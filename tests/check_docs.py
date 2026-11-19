#!/usr/bin/env python3
"""Validação leve dos links locais Markdown; sem dependências externas."""
from pathlib import Path
import re
import sys

ROOT = Path(__file__).resolve().parents[1]
LINK = re.compile(r"\]\(([^)]+)\)")
SKIP = ("https://", "http://", "mailto:", "#")
problems = []
checked = 0

for doc in sorted(ROOT.rglob("*.md")):
    if ".git" in doc.parts:
        continue
    content = doc.read_text(encoding="utf-8")
    if content.count("```") % 2:
        problems.append(f"{doc.relative_to(ROOT)}: cercas de código desequilibradas")
    for raw in LINK.findall(content):
        target = raw.split("#", 1)[0]
        if not target or raw.startswith(SKIP):
            continue
        # Apenas links internos relativos neste teste.
        if target.startswith("/") or "://" in target:
            continue
        resolved = (doc.parent / target).resolve()
        checked += 1
        if not resolved.is_relative_to(ROOT.resolve()) or not resolved.is_file():
            problems.append(f"{doc.relative_to(ROOT)}: destino não encontrado: {raw}")

for issue in problems:
    print("ERRO:", issue)
print(f"Documentos verificados; {checked} links internos conferidos.")
if problems:
    sys.exit(1)
