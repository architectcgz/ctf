#!/usr/bin/env python3
"""Project-specific documentation consistency checks."""

from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
FAILURES: list[str] = []
ARCHITECTURE_DOC_ROOT = ROOT / "docs/architecture"

ARCHITECTURE_EVIDENCE_RE = re.compile(
    r"("
    r"code/(backend|frontend)/|"
    r"docs/(architecture|contracts|operations|requirements)/|"
    r"scripts/|"
    r"challenges/|"
    r"\b(GET|POST|PUT|PATCH|DELETE)\s+/|"
    r"API[:：]|"
    r"测试[:：]|"
    r"依据[:：]|"
    r"Guardrail|"
    r"migration|"
    r"router|"
    r"composable|"
    r"store|"
    r"handler|"
    r"service|"
    r"repository"
    r")",
    re.IGNORECASE,
)
VAGUE_ARCHITECTURE_PHRASES = ["统一处理", "自动同步", "平台负责", "系统管理"]
PLANNING_WORD_RE = re.compile(r"(待落地|TODO|后续计划|计划支持)")


def fail(message: str) -> None:
    FAILURES.append(message)


def read(path: str) -> str:
    return (ROOT / path).read_text(encoding="utf-8")


def line_number(text: str, offset: int) -> int:
    return text.count("\n", 0, offset) + 1


def markdown_section(text: str, heading: str) -> str | None:
    pattern = re.compile(rf"^##\s+{re.escape(heading)}\s*$", re.MULTILINE)
    match = pattern.search(text)
    if match is None:
        return None
    next_heading = re.search(r"^##\s+", text[match.end() :], flags=re.MULTILINE)
    end = len(text) if next_heading is None else match.end() + next_heading.start()
    return text[match.end() : end]


def paragraphs_with_offsets(text: str) -> list[tuple[int, str]]:
    paragraphs: list[tuple[int, str]] = []
    start = 0
    for part in re.split(r"\n\s*\n", text):
        stripped = part.strip()
        if stripped:
            offset = text.find(part, start)
            paragraphs.append((offset, stripped))
            start = offset + len(part)
    return paragraphs


def git_changed_architecture_docs() -> set[Path]:
    paths: set[Path] = set()
    commands = [
        ["git", "diff", "--name-only", "--cached", "--", "docs/architecture"],
        ["git", "diff", "--name-only", "--", "docs/architecture"],
    ]
    for command in commands:
        result = subprocess.run(command, cwd=ROOT, text=True, capture_output=True, check=False)
        if result.returncode != 0:
            continue
        for raw in result.stdout.splitlines():
            if raw.endswith(".md"):
                path = ROOT / raw
                if path.exists():
                    paths.add(path)
    return paths


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
        "harness",
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


def check_current_design_quality(path: Path, strict: bool) -> None:
    text = path.read_text(encoding="utf-8")
    rel = path.relative_to(ROOT)
    current_design = markdown_section(text, "当前设计")
    is_index_doc = path.name == "README.md" or "索引" in path.name

    if strict and not is_index_doc and "pages" not in path.relative_to(ARCHITECTURE_DOC_ROOT).parts and current_design is None:
        fail(f"{rel}: changed architecture doc must include `## 当前设计` or stay out of docs/architecture")
        return
    if current_design is None:
        return

    if "负责" not in current_design or "不负责" not in current_design:
        fail(f"{rel}: `## 当前设计` must define component boundaries with 负责 / 不负责")
    if not ARCHITECTURE_EVIDENCE_RE.search(current_design):
        fail(f"{rel}: `## 当前设计` must cite code paths, API, contract docs, tests, or guardrails")

    for offset, paragraph in paragraphs_with_offsets(current_design):
        for phrase in VAGUE_ARCHITECTURE_PHRASES:
            if phrase in paragraph and not ARCHITECTURE_EVIDENCE_RE.search(paragraph):
                fail(
                    f"{rel}:{line_number(text, text.find(current_design) + offset)}: "
                    f"`{phrase}` needs nearby code/API/data/test evidence"
                )
        if "支持" in paragraph:
            has_required_detail = all(word in paragraph for word in ["入口", "数据结构", "状态", "测试"])
            if not has_required_detail and not ARCHITECTURE_EVIDENCE_RE.search(paragraph):
                fail(
                    f"{rel}:{line_number(text, text.find(current_design) + offset)}: "
                    "`支持` claim needs entry, data structure, state change, and test evidence"
                )
        if PLANNING_WORD_RE.search(paragraph) and "待确认" not in paragraph and "已知限制" not in paragraph:
            fail(
                f"{rel}:{line_number(text, text.find(current_design) + offset)}: "
                "Current architecture facts must not describe plans as implemented facts; use 待确认 or 已知限制"
            )


def check_architecture_doc_quality() -> None:
    changed_docs = git_changed_architecture_docs()
    current_design_docs = {
        path
        for path in ARCHITECTURE_DOC_ROOT.rglob("*.md")
        if "## 当前设计" in path.read_text(encoding="utf-8")
    }

    for path in sorted(changed_docs | current_design_docs):
        check_current_design_quality(path, strict=path in changed_docs)


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
        "harness/prompts/AGENTS.md",
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
    check_architecture_doc_quality()
    check_diagrams()

    if FAILURES:
        for item in FAILURES:
            print(f"  FAIL — {item}")
        return 1
    print("  PASS — documentation references and diagram sources")
    return 0


if __name__ == "__main__":
    sys.exit(main())
