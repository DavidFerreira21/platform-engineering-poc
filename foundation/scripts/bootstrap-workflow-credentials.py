#!/usr/bin/env python3
"""Bootstrap TEMPORÁRIO de identidade AWS para o Terraform Executor no Hub kind (Etapa 08).

Sem dependências Python externas. NUNCA grava credenciais em arquivos nem Git.
Requer uma role de origem na conta da plataforma e role de destino em UAT já criadas.
"""
import argparse
import base64
import json
import os
import re
import shutil
import subprocess
import sys

CONTEXT = "kind-platform-management"
NAMESPACE = "argo-workflows"
SECRET_NAME = "terraform-executor-bootstrap"
CONFIG_NAME = "terraform-executor-aws-config"
PROFILE = "uat-executor"
ROLE_RE = re.compile(r"^arn:aws:iam::([0-9]{12}):role/([A-Za-z0-9+=,.@_/-]+)$")


def run(argv, *, env=None, input_text=None):
    completed = subprocess.run(argv, input=input_text, text=True, capture_output=True, env=env, check=False)
    if completed.returncode:
        # stderr de AWS/kubectl não deve conter credenciais; não repassar stdout de comandos com credenciais.
        raise RuntimeError(f"Falhou {argv[0]} {argv[1]} (código {completed.returncode}): "
                           f"{completed.stderr.strip()[:350]}")
    return completed.stdout.strip()


def aws(args, *, env=None):
    return json.loads(run(["aws", *args, "--output", "json"], env=env))


def kubectl(args, *, data=None):
    return run(["kubectl", "--context", CONTEXT, "-n", NAMESPACE, *args], input_text=data)


def validate_cli():
    for tool in ("aws", "kubectl"):
        if not shutil.which(tool):
            raise RuntimeError(f"Instale {tool} antes do bootstrap")
    current = run(["kubectl", "config", "current-context"])
    if current != CONTEXT:
        raise RuntimeError(f"Contexto ativo {current!r}; selecione explicitamente {CONTEXT}")
    kubectl(["get", "namespace", NAMESPACE, "-o", "name"])


def check_role(role, account=None):
    match = ROLE_RE.fullmatch(role)
    if not match:
        raise ValueError(f"Role ARN IAM inválido: {role!r}")
    if account and match.group(1) != account:
        raise ValueError("A role de destino não pertence à conta UAT esperada")


def assume(role, name, env=None):
    return aws(["sts", "assume-role", "--role-arn", role, "--role-session-name", name,
                "--duration-seconds", "3600"], env=env)["Credentials"]


def credentials_env(creds, region):
    return {
        **os.environ,
        "AWS_ACCESS_KEY_ID": creds["AccessKeyId"],
        "AWS_SECRET_ACCESS_KEY": creds["SecretAccessKey"],
        "AWS_SESSION_TOKEN": creds["SessionToken"],
        "AWS_DEFAULT_REGION": region,
        "AWS_REGION": region,
        "AWS_EC2_METADATA_DISABLED": "true",
        "AWS_PAGER": "",
    }


def object_create(obj):
    # create, e não apply: evita guardar credenciais no annotation last-applied-configuration.
    kubectl(["create", "-f", "-"], data=json.dumps(obj))


def prepare(args):
    validate_cli()
    check_role(args.bootstrap_role_arn)
    check_role(args.target_role_arn, args.uat_account)
    if not re.fullmatch(r"[0-9]{12}", args.uat_account):
        raise ValueError("A conta UAT deve ter 12 dígitos")
    if not re.fullmatch(r"[a-z]{2}-[a-z0-9-]+-[0-9]", args.region):
        raise ValueError("Região AWS inválida")
    if args.bootstrap_role_arn == args.target_role_arn:
        raise ValueError("A role de origem deve ser distinta da role de destino")

    # Sem acesso à role de bootstrap via SSO a operação falha sem criar Secret.
    human = aws(["sts", "get-caller-identity", "--profile", args.profile, "--region", args.region])
    if human['Account'] != args.uat_account:
        raise RuntimeError("O profile SSO não está autenticado na conta UAT esperada; valide a Etapa 07")
    print(f"Identidade humana autenticada na UAT: {human['Arn']}")
    source = assume(args.bootstrap_role_arn, "platform-poc-bootstrap",
                    env={**os.environ, "AWS_PROFILE": args.profile})
    source_env = credentials_env(source, args.region)
    origin = aws(["sts", "get-caller-identity", "--region", args.region], env=source_env)
    if args.bootstrap_role_arn.split(":")[4] != origin["Account"]:
        raise RuntimeError("STS retornou conta de bootstrap diferente da esperada")

    # Verificar o encadeamento ANTES de armazenar a sessão temporária no cluster.
    config_content = (
        f"[profile {PROFILE}]\nrole_arn = {args.target_role_arn}\n"
        "credential_source = Environment\n"
        f"region = {args.region}\n"
    )
    dest = assume(args.target_role_arn, "platform-poc-uat-check", env=source_env)
    target = aws(["sts", "get-caller-identity", "--region", args.region],
                 env=credentials_env(dest, args.region))
    if target["Account"] != args.uat_account:
        raise RuntimeError("STS do destino não corresponde a UAT")

    # Requer namespace existente. Se houver Secret anterior, falhar fechado.
    existing = kubectl(["get", "secret", SECRET_NAME, "--ignore-not-found", "-o", "name"])
    if existing:
        raise RuntimeError("Secret já existe. Faça cleanup antes de renovar a sessão.")
    if kubectl(["get", "configmap", CONFIG_NAME, "--ignore-not-found", "-o", "name"]):
        raise RuntimeError("ConfigMap já existe. Faça cleanup antes de renovar a sessão.")

    expiry = source["Expiration"]
    if isinstance(expiry, str):
        expiry_txt = expiry
    else:
        expiry_txt = str(expiry)
    annotations = {"platform.example.io/bootstrap-expiration": expiry_txt,
                   "platform.example.io/disposable-lab": "08"}
    secret_values = {
        "AWS_ACCESS_KEY_ID": source["AccessKeyId"],
        "AWS_SECRET_ACCESS_KEY": source["SecretAccessKey"],
        "AWS_SESSION_TOKEN": source["SessionToken"],
        "AWS_DEFAULT_REGION": args.region,
        "AWS_REGION": args.region,
        "AWS_EC2_METADATA_DISABLED": "true",
    }
    secret = {
        "apiVersion": "v1", "kind": "Secret",
        "metadata": {"name": SECRET_NAME, "namespace": NAMESPACE, "annotations": annotations,
                     "labels": {"app.kubernetes.io/part-of": "terraformstack-executor"}},
        "type": "Opaque",
        "data": {k: base64.b64encode(v.encode()).decode() for k, v in secret_values.items()},
    }
    configmap = {
        "apiVersion": "v1", "kind": "ConfigMap",
        "metadata": {"name": CONFIG_NAME, "namespace": NAMESPACE,
                     "labels": {"app.kubernetes.io/part-of": "terraformstack-executor"}},
        "data": {"config": config_content},
    }
    object_create(configmap)
    try:
        object_create(secret)
    except Exception:
        kubectl(["delete", "configmap", CONFIG_NAME, "--ignore-not-found"])
        raise
    print("OK: configuração de role e credencial STS efêmera criadas em argo-workflows.")
    print(f"Conta UAT verificada via STS: {target['Account']}")
    print(f"Expiração da sessão bootstrap: {expiry_txt} (renovar manualmente).")
    print("Não execute Terraform enquanto o Workflow ainda não validar a role, a aprovação e o state.")


def cleanup():
    validate_cli()
    # Remoção explícita; nunca excluir credenciais de outros componentes.
    kubectl(["delete", "secret", SECRET_NAME, "--ignore-not-found"])
    kubectl(["delete", "configmap", CONFIG_NAME, "--ignore-not-found"])
    print("OK: material temporário do executor removido")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    p = sub.add_parser("prepare", help="Configurar credenciais temporárias no kind")
    p.add_argument("--profile", default="platform-poc-uat")
    p.add_argument("--uat-account", required=True)
    p.add_argument("--region", default="us-east-1")
    p.add_argument("--bootstrap-role-arn", required=True,
                   help="Role mínima de origem assumida via SSO; somente STS AssumeRole UAT")
    p.add_argument("--target-role-arn", required=True,
                   help="Role de execução autorizada na conta UAT")
    sub.add_parser("cleanup", help="Excluir Secret/ConfigMap efêmeros")
    args = parser.parse_args()
    try:
        prepare(args) if args.command == "prepare" else cleanup()
    except (ValueError, RuntimeError, KeyError, json.JSONDecodeError) as error:
        print(f"ERRO: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
