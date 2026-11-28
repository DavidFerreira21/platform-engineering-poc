#!/usr/bin/env python3
"""Validação leve de links Markdown locais, âncoras e cercas de código."""
from collections import Counter
from functools import lru_cache
from pathlib import Path
from urllib.parse import unquote
import re
import sys

ROOT = Path(__file__).resolve().parents[1]
LINK = re.compile(r"\]\(([^)]+)\)")
HEAD = re.compile(r"^#{1,6}\s+(.+?)\s*#*\s*$")
SKIP = ("https://", "http://", "mailto:")
problems = []
checked = 0
anchors_checked = 0


def slug(text):
    """Slug aproximado de cabeçalhos GitHub Markdown, preservando acentos."""
    text = re.sub(r"\[([^\]]+)\]\([^)]+\)", r"\1", text)
    text = re.sub(r"<[^>]*>", "", text)
    text = re.sub(r"[" + chr(96) + r"*_~]", "", text)
    text = text.lower().strip()
    text = re.sub(r"[^\w\s-]", "", text, flags=re.UNICODE)
    return re.sub(r"\s", "-", text)


@lru_cache(maxsize=None)
def headings(path):
    anchors = set()
    duplicates = Counter()
    in_fence = False
    for line in path.read_text(encoding="utf-8").splitlines():
        if line.lstrip().startswith(chr(96) * 3) or line.lstrip().startswith("~~~"):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        match = HEAD.match(line)
        if match:
            base = slug(match.group(1))
            suffix = duplicates[base]
            anchors.add(f"{base}-{suffix}" if suffix else base)
            duplicates[base] += 1
    return anchors


for doc in sorted(ROOT.rglob("*.md")):
    if any(part in {".git", ".terraform"} for part in doc.parts):
        continue
    content = doc.read_text(encoding="utf-8")
    if content.count(chr(96) * 3) % 2:
        problems.append(f"{doc.relative_to(ROOT)}: cercas de código desequilibradas")
    for raw in LINK.findall(content):
        if raw.startswith(SKIP):
            continue
        target, sep, anchor = raw.partition("#")
        if target.startswith("/") or "://" in target:
            continue
        resolved = (doc.parent / unquote(target)).resolve() if target else doc.resolve()
        checked += 1
        if not resolved.is_relative_to(ROOT.resolve()) or not resolved.is_file():
            problems.append(f"{doc.relative_to(ROOT)}: destino não encontrado: {raw}")
            continue
        if sep and anchor and resolved.suffix.lower() == ".md":
            anchors_checked += 1
            if unquote(anchor) not in headings(resolved):
                problems.append(f"{doc.relative_to(ROOT)}: âncora não encontrada: {raw}")

for issue in problems:
    print("ERRO:", issue)
print(f"Documentos verificados; {checked} links internos e {anchors_checked} âncoras conferidos.")
if problems:
    sys.exit(1)
