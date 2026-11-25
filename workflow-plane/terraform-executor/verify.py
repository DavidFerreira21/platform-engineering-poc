#!/usr/bin/env python3
"""Lab 08: valida AWS STS, backend S3 com lock e Configuration sem apply."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import sys

SOURCE = "/opt/foundation/lab08/validation-stack"


def run(cmd, env):
    p = subprocess.run(cmd, cwd=SOURCE, env=env, text=True, check=False)
    if p.returncode:
        raise RuntimeError(f"{cmd[0]} retornou exit {p.returncode}")
    return p.returncode


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--bucket", required=True)
    parser.add_argument("--account", required=True)
    parser.add_argument("--region", required=True)
    parser.add_argument("--entity", default="entity-a")
    parser.add_argument("--environment", default="uat")
    parser.add_argument("--stack", choices=("network", "eks"), default="network")
    args = parser.parse_args()

    # O diretório de trabalho é read-only na imagem; -chdir não grava fontes.
    # Copiar para emptyDir pertencente ao UID não privilegiado.
    import shutil
    shutil.copytree("/opt/foundation/modules/configuration", "/workspace/foundation/modules/configuration", dirs_exist_ok=True)
    shutil.copytree("/opt/foundation/config", "/workspace/foundation/config", dirs_exist_ok=True)
    shutil.copytree(SOURCE, "/workspace/foundation/lab08/validation-stack", dirs_exist_ok=True)
    cwd = "/workspace/foundation/lab08/validation-stack"
    env = dict(os.environ, AWS_PROFILE="uat-executor",
               AWS_DEFAULT_REGION=args.region, AWS_REGION=args.region,
               TF_IN_AUTOMATION="true", AWS_EC2_METADATA_DISABLED="true")
    who = subprocess.run(["aws", "--profile", "uat-executor", "sts", "get-caller-identity",
                          "--output", "json"], text=True, capture_output=True, env=env)
    if who.returncode:
        raise RuntimeError(f"Falha no STS da role destino: {who.stderr[:250]}")
    caller = json.loads(who.stdout)
    if caller["Account"] != args.account:
        raise RuntimeError("Conta efetiva do Pod diverge do destino permitido")
    print(f"STS: destino confirmado {args.account}/{args.region}", flush=True)

    backend = [
        "terraform", "init", "-input=false", "-reconfigure", "-no-color",
        f"-backend-config=bucket={args.bucket}",
        f"-backend-config=key=platform/{args.entity}/{args.environment}/lab08/validation.tfstate",
        f"-backend-config=region={args.region}",
        "-backend-config=use_lockfile=true",
        "-backend-config=encrypt=true",
    ]
    subprocess.run(["terraform", "fmt", "-check", "-recursive", "-no-color", cwd],
                   check=True, env=env)
    subprocess.run(backend, check=True, cwd=cwd, env=env)
    subprocess.run(["terraform", "validate", "-no-color"], check=True, cwd=cwd, env=env)
    plan = subprocess.run([
        "terraform", "plan", "-input=false", "-no-color", "-detailed-exitcode",
        "-lock-timeout=60s",
        "-var", f"entity={args.entity}",
        "-var", f"environment={args.environment}",
        "-var", f"stack={args.stack}",
    ], cwd=cwd, env=env, check=False)
    if plan.returncode not in (0, 2):
        raise RuntimeError(f"terraform plan falhou com exit {plan.returncode}")
    print("OK: STS, S3 backend e Configuration validados; nenhum terraform apply foi executado.", flush=True)


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, subprocess.CalledProcessError) as exc:
        print(f"ERRO: {exc}", file=sys.stderr)
        sys.exit(1)
