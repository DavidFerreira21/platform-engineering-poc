"""Publica os outputs permitidos após a verificação da stack pelo Workflow."""
import argparse
import json
import re
import subprocess
import sys


def aws(region, *args):
    return subprocess.check_output(["aws", "--region", region, *args], text=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("entity", "environment", "region", "account-id"):
        parser.add_argument("--" + name, required=True)
    args = parser.parse_args()
    if not all(re.fullmatch(r"[a-z0-9][a-z0-9-]*", part)
               for part in (args.entity, args.environment)):
        parser.error("Entity ou Environment inválido.")
    values = json.load(sys.stdin)
    allowed = r"network/(vpc-id|(web|database)-(nonpci|pci)-subnet-ids)|eks/(hub|management|nonpci|pci)/cluster-(name|endpoint)"
    if not isinstance(values, dict) or any(
        not re.fullmatch(allowed, key) or not isinstance(value, str) or not value
        for key, value in values.items()
    ):
        parser.error("Output fora do contrato do Bridge.")
    if not values:
        return
    account = aws(args.region, "sts", "get-caller-identity", "--query", "Account", "--output", "text").strip()
    if account != args.account_id:
        parser.error("Identidade do publicador fora da conta esperada.")
    for key, value in sorted(values.items()):
        path = f"/platform/{args.entity}/{args.environment}/{key}"
        aws(args.region, "ssm", "put-parameter", "--name", path, "--type", "String",
            "--value", value, "--overwrite")


if __name__ == "__main__":
    main()
