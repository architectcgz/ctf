"""Helpers for validating current documentation code-path references."""

from __future__ import annotations

import re
from pathlib import Path


BACKTICK_RE = re.compile(r"`([^`]+)`")
LINE_SUFFIX_RE = re.compile(r":\d+$")
SKIP_MARKERS = ("*", "...", "<", ">", "(", ")", "{", "}", "$")


def _is_relative_to(path: Path, parent: Path) -> bool:
    try:
        path.relative_to(parent)
        return True
    except ValueError:
        return False


def _normalize_token(token: str) -> str | None:
    value = token.strip().rstrip(".,;:")
    if "/" not in value or re.search(r"\s", value):
        return None
    if value.startswith(("http://", "https://", "#")):
        return None
    if any(marker in value for marker in SKIP_MARKERS):
        return None
    return LINE_SUFFIX_RE.sub("", value)


def _resolve_path(root: Path, token: str) -> Path | None:
    if token.startswith("frontend/"):
        return root / "code/frontend" / token.removeprefix("frontend/")
    if token.startswith("internal/"):
        return root / "code/backend" / token
    if token.startswith(("code/", "docs/", "scripts/", "challenges/", "harness/")):
        return root / token
    return None


def missing_changed_feature_doc_code_refs(root: Path, changed_docs: set[Path]) -> list[str]:
    """Return missing code-path references in changed feature architecture docs."""

    feature_root = root / "docs/architecture/features"
    failures: list[str] = []
    for path in sorted(changed_docs):
        if not _is_relative_to(path, feature_root):
            continue
        text = path.read_text(encoding="utf-8")
        rel = path.relative_to(root)
        for match in BACKTICK_RE.finditer(text):
            token = _normalize_token(match.group(1))
            if token is None:
                continue
            target = _resolve_path(root, token)
            if target is None or target.exists():
                continue
            line = text.count("\n", 0, match.start()) + 1
            failures.append(f"{rel}:{line}: missing current code path `{token}`")
    return failures
