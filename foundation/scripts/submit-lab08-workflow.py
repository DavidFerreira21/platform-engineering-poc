#!/usr/bin/env python3
"""Inicia validação Lab08, sem depender de argo CLI ou escrever arquivo temporário."""
import argparse
import json
import re
import subprocess
import sys


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--bucket", required=True)
    p.add_argument("--account", required=True)
    p.add_argument("--region", required=True)
    p.add_argument("--stack", choices=("network", "eks"), default="network")
    args = p.parse_args()
    if not re.fullmatch(r"\d{12}", args.account):
        p.error("Account ID inválido")
    if not re.fullmatch(r"[a-z0-9][a-z0-9.-]{3,61}[a-z0-9]", args.bucket):
        p.error("Bucket inválido")
    if not re.fullmatch(r"[a-z]{2}-[a-z0-9-]+-[0-9]", args.region):
        p.error("Região inválida")
    workflow = {
        "apiVersion": "argoproj.io/v1alpha1", "kind": "Workflow",
        "metadata": {"generateName": "lab08-terraform-check-", "namespace": "argo-workflows"},
        "spec": {
            "workflowTemplateRef": {"name": "lab08-terraform-executor-check"},
            "arguments": {"parameters": [
                {"name": "bucket", "value": args.bucket},
                {"name": "account", "value": args.account},
                {"name": "region", "value": args.region},
                {"name": "stack", "value": args.stack},
            ]},
        },
    }
    command = ["kubectl", "--context", "kind-platform-management", "-n", "argo-workflows",
               "create", "-f", "-", "-o", "name"]
    result = subprocess.run(command, input=json.dumps(workflow), text=True, capture_output=True)
    if result.returncode:
        print(f"ERRO: {result.stderr[:350]}", file=sys.stderr)
        return 1
    print(result.stdout.strip())
    return 0


if __name__ == "__main__":
    sys.exit(main())
