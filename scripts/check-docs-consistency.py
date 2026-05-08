#!/usr/bin/env python3
"""Project-specific documentation consistency checks."""

from __future__ import annotations

import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
FAILURES: list[str] = []


def fail(message: str) -> None:
    FAILURES.append(message)


def read(path: str) -> str:
    return (ROOT / path).read_text(encoding="utf-8")


def resolve_ref(source: Path, token: str) -> Path | None:
    if token.startswith(("http://", "https://", "#")):
        return None
    if "*" in token or "<" in token or ">" in token:
        return None
    if token.startswith("@/"):
        return ROOT / "code/frontend/src" / token[2:]
    if token.startswith("./") or token.startswith("../"):
        return (source.parent / token).resolve()
    first = token.split("/", 1)[0]
    if first in {
        "AGENTS.md",
        "challenges",
        "code",
        "concepts",
        "docs",
        "feedback",
        "practice",
        "prompts",
        "references",
        "scripts",
        "thinking",
        "works",
    }:
        return ROOT / token
    if token.endswith(".md"):
        return source.parent / token
    return None


def check_backtick_refs(path: str) -> None:
    source = ROOT / path
    text = source.read_text(encoding="utf-8")
    for match in re.finditer(r"`([^`]+)`", text):
        token = match.group(1).strip()
        if not token.endswith((".md", ".yaml", ".yml", ".mmd")):
            continue
        target = resolve_ref(source, token)
        if target is None:
            continue
        if not target.exists():
            line = text.count("\n", 0, match.start()) + 1
            fail(f"{path}:{line}: missing referenced file `{token}`")


def check_no_stale_refs(path: str, patterns: list[str]) -> None:
    text = read(path)
    for pattern in patterns:
        if pattern in text:
            fail(f"{path}: stale reference still present: {pattern}")


def check_architecture_status() -> None:
    for path in (ROOT / "docs/architecture").rglob("*.md"):
        text = path.read_text(encoding="utf-8")
        head = "\n".join(text.splitlines()[:20])
        if re.search(r"状态：\s*(Draft|初稿)", head, flags=re.IGNORECASE):
            fail(f"{path.relative_to(ROOT)}: active architecture doc must not be Draft/初稿")


def check_diagrams() -> None:
    allowed = re.compile(r"\b(flowchart|sequenceDiagram|stateDiagram-v2|classDiagram|C4Context|C4Container|C4Component)\b")
    sensitive = re.compile(r"(password|passwd|secret|token|cookie|private[_-]?key|DATABASE_URL=|redis://|postgres://)", re.IGNORECASE)
    for path in list((ROOT / "docs").rglob("*.mmd")) + list((ROOT / "docs").rglob("*.puml")):
        text = path.read_text(encoding="utf-8")
        rel = path.relative_to(ROOT)
        if path.suffix == ".mmd" and not allowed.search(text):
            fail(f"{rel}: Mermaid source must declare a supported diagram type")
        if sensitive.search(text):
            fail(f"{rel}: diagram source appears to contain sensitive configuration text")


def main() -> int:
    for path in [
        "docs/architecture/README.md",
        "docs/architecture/features/专题架构索引.md",
        "docs/design/README.md",
        "prompts/AGENTS.md",
    ]:
        check_backtick_refs(path)

    check_no_stale_refs(
        "docs/design/README.md",
        [
            "docs/superpowers/plans/",
            "docs/superpowers/specs/",
            "docs/superpowers/specs/*.md",
        ],
    )
    check_no_stale_refs(
        "docs/architecture/frontend/01-architecture-overview.md",
        ["design-system/MASTER.md", "frontend/design-system/"],
    )
    check_no_stale_refs(
        "docs/architecture/frontend/06-components.md",
        ["design-system/MASTER.md", "frontend/design-system/"],
    )

    check_architecture_status()
    check_diagrams()

    if FAILURES:
        for item in FAILURES:
            print(f"  FAIL — {item}")
        return 1
    print("  PASS — documentation references and diagram sources")
    return 0


if __name__ == "__main__":
    sys.exit(main())
