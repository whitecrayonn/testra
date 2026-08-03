#!/usr/bin/env python3
"""Audit markdown files under testra/docs for broken internal links and missing code refs."""
import re
import json
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
TESTRA_ROOT = REPO_ROOT / "testra"
DOCS = TESTRA_ROOT / "docs"

LINK_RE = re.compile(r"!?\[([^\]]*)\]\(([^)\s]+)\)")
REF_LINK_RE = re.compile(r"^\[[^\]]+\]:\s*(\S+)", re.MULTILINE)
# Inline code paths inside testra or at repo root
CODE_RE = re.compile(r"`((?:testra/)?(?:apps|packages|infra|scripts|docs|04_Architecture)/[^`]+)`")

EXTERNAL_PREFIXES = ("http://", "https://", "mailto:", "ftp://", "//")

def is_external(target):
    return any(target.startswith(p) for p in EXTERNAL_PREFIXES)

def resolve_link(file_dir, target):
    """Resolve a markdown link target to a filesystem Path (or None)."""
    if not target or target.startswith("#"):
        return None
    if is_external(target):
        return None
    target = target.split("#")[0]
    if target.startswith("/"):
        return (TESTRA_ROOT / target.lstrip("/")).resolve()
    return (file_dir / target).resolve()

def exists_link(path):
    if path is None:
        return True
    if path.exists():
        return True
    if not path.suffix and path.with_suffix(".md").exists():
        return True
    if path.is_dir() and (path / "README.md").exists():
        return True
    return False

def resolve_code_ref(file_dir, target):
    """Resolve an inline code path reference; returns Path or None to skip."""
    target = target.strip()
    # Skip dynamic/placeholder/wildcard patterns
    if any(c in target for c in "*{}<>[]"):
        return None
    # Strip trailing punctuation that may be captured
    target = target.rstrip(".,;:!?')}\"")
    # Strip #Lnn line anchors
    target = re.sub(r"#L?\d+(?:-\d+)?$", "", target)
    # Repo-root-relative paths (outside testra/)
    if target.startswith("04_Architecture/"):
        return (REPO_ROOT / target).resolve()
    # Root-level product docs (testra-*.md)
    if target.startswith("testra-") and target.endswith(".md"):
        return (REPO_ROOT / target).resolve()
    # Strip leading testra/ to make it relative to testra root
    if target.startswith("testra/"):
        target = target[len("testra/"):]
    # All other recognized prefixes live inside testra
    if any(target.startswith(p) for p in ("apps/", "packages/", "infra/", "scripts/", "docs/")):
        return (TESTRA_ROOT / target).resolve()
    return None

def audit():
    active_files = []
    archive_files = []
    broken_links = []
    broken_refs = []

    for path in sorted(DOCS.rglob("*.md")):
        rel = path.relative_to(TESTRA_ROOT)
        if ".git" in path.parts:
            continue
        if "archive" in path.parts:
            archive_files.append(str(rel.as_posix()))
        else:
            active_files.append(str(rel.as_posix()))
        text = path.read_text(encoding="utf-8", errors="ignore")
        file_dir = path.parent

        # Markdown links
        for _label, target in LINK_RE.findall(text):
            resolved = resolve_link(file_dir, target)
            if resolved is None:
                continue
            if not exists_link(resolved):
                broken_links.append({
                    "file": str(rel.as_posix()),
                    "target": target,
                    "resolved": str(resolved.relative_to(REPO_ROOT).as_posix()) if resolved.is_relative_to(REPO_ROOT) else str(resolved),
                })

        # Reference links
        for target in REF_LINK_RE.findall(text):
            resolved = resolve_link(file_dir, target)
            if resolved is None:
                continue
            if not exists_link(resolved):
                broken_links.append({
                    "file": str(rel.as_posix()),
                    "target": target,
                    "resolved": str(resolved.relative_to(REPO_ROOT).as_posix()) if resolved.is_relative_to(REPO_ROOT) else str(resolved),
                })

        # Code references in backticks
        for raw in CODE_RE.findall(text):
            if isinstance(raw, tuple):
                raw = raw[0]
            resolved = resolve_code_ref(file_dir, raw)
            if resolved is None:
                continue
            if not resolved.exists():
                broken_refs.append({
                    "file": str(rel.as_posix()),
                    "target": raw,
                    "resolved": str(resolved.relative_to(REPO_ROOT).as_posix()) if resolved.is_relative_to(REPO_ROOT) else str(resolved),
                })

    print(json.dumps({
        "active_files": active_files,
        "archive_files": archive_files,
        "active_count": len(active_files),
        "archive_count": len(archive_files),
        "broken_links": broken_links,
        "broken_refs": broken_refs,
    }, indent=2))

if __name__ == "__main__":
    audit()
