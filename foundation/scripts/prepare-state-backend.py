#!/usr/bin/env python3
"""Provisiona/verifica S3 de state na conta UAT, somente após --execute explícito.

Usar AWS SSO humano do Lab 07. A conta e bucket são informados pelo operador,
nunca extraídos de inputs não confiáveis do TerraformStack.
"""
import argparse
import json
import re
import subprocess
import sys


def aws(profile, region, *args, json_output=False):
    cmd = ["aws", "--profile", profile, "--region", region, *args]
    p = subprocess.run(cmd, text=True, capture_output=True)
    if p.returncode:
        raise RuntimeError(f"{args[0]} falhou (exit {p.returncode}): {p.stderr[:300]}")
    return json.loads(p.stdout) if json_output else p.stdout.strip()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--profile", default="platform-poc-uat")
    parser.add_argument("--account", required=True)
    parser.add_argument("--region", required=True)
    parser.add_argument("--bucket", required=True)
    parser.add_argument("--execute", action="store_true", help="Autoriza criar/configurar bucket; sem flag somente exibe o plano")
    args = parser.parse_args()
    if not re.fullmatch(r"\d{12}", args.account):
        parser.error("Account ID precisa ter 12 dígitos")
    if not re.fullmatch(r"[a-z]{2}-[a-z0-9-]+-[0-9]", args.region):
        parser.error("Região inválida")
    if not re.fullmatch(r"[a-z0-9][a-z0-9.-]{3,61}[a-z0-9]", args.bucket):
        parser.error("Bucket S3 inválido")
    if args.bucket.startswith("xn--") or ".." in args.bucket:
        parser.error("Bucket inválido")

    identity = aws(args.profile, args.region, "sts", "get-caller-identity", json_output=True)
    if identity["Account"] != args.account:
        raise RuntimeError("Conta autenticada não corresponde ao Environment UAT")
    if not args.execute:
        print("PREVIEW: criar/verificar bucket S3 com BlockPublicAccess, versionamento e SSE-S3.")
        print(f"Account={args.account} Region={args.region} Bucket={args.bucket}")
        print("Nenhum recurso criado. Para autorizar, acrescente --execute.")
        return

    # HeadBucket não é suficiente para comprovar ownership; GetBucketLocation
    # deve apontar à região, e ExpectedBucketOwner acompanha chamadas suportadas.
    exists = True
    p = subprocess.run(
        ["aws", "--profile", args.profile, "--region", args.region,
         "s3api", "head-bucket", "--bucket", args.bucket,
         "--expected-bucket-owner", args.account], capture_output=True, text=True)
    if p.returncode:
        # Não presumir "não existe": AccessDenied pode significar bucket de terceiro.
        # Tentar create-bucket com nome global: qualquer colisão falha fechado.
        exists = False
    if not exists:
        call = ["s3api", "create-bucket", "--bucket", args.bucket]
        if args.region != "us-east-1":
            call += ["--create-bucket-configuration", f"LocationConstraint={args.region}"]
        aws(args.profile, args.region, *call)
    location = aws(args.profile, args.region, "s3api", "get-bucket-location",
                   "--bucket", args.bucket, "--expected-bucket-owner", args.account,
                   "--query", "LocationConstraint", "--output", "text")
    if location in ("None", "null", ""):
        location = "us-east-1"
    if location != args.region:
        raise RuntimeError("Bucket em região diferente da autorizada")

    aws(args.profile, args.region, "s3api", "put-public-access-block", "--bucket", args.bucket,
        "--expected-bucket-owner", args.account, "--public-access-block-configuration",
        "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true")
    aws(args.profile, args.region, "s3api", "put-bucket-versioning", "--bucket", args.bucket,
        "--expected-bucket-owner", args.account, "--versioning-configuration", "Status=Enabled")
    aws(args.profile, args.region, "s3api", "put-bucket-encryption", "--bucket", args.bucket,
        "--expected-bucket-owner", args.account, "--server-side-encryption-configuration",
        '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}')
    print(f"OK: state bucket protegido: {args.bucket} ({args.account}/{args.region})")
    print("NÃO excluir o bucket na limpeza do laboratório: estados e histórico são duráveis.")


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, ValueError) as error:
        print(f"ERRO: {error}", file=sys.stderr)
        raise SystemExit(1)
