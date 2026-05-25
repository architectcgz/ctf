#!/usr/bin/env python3
from __future__ import annotations

import argparse
import sys
from pathlib import Path

from common import ROOT, validate_reuse_decision


def resolve_document_path(document: str) -> Path:
    path = Path(document)
    if not path.suffix:
        path = Path(".harness/reuse-decisions") / f"{document}.md"
    if not path.is_absolute():
        path = ROOT / path
    return path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="verify that a task-scoped reuse decision exists before protected implementation starts"
    )
    parser.add_argument(
        "--document",
        required=True,
        help="reuse decision path or task slug; slugs resolve to .harness/reuse-decisions/<slug>.md",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    path = resolve_document_path(args.document)
    if not path.is_file():
        rel = path.relative_to(ROOT).as_posix()
        print(f"FAIL: reuse decision document does not exist: {rel}", file=sys.stderr)
        print("Create it before touching protected implementation files.", file=sys.stderr)
        return 1

    text = path.read_text(encoding="utf-8").strip()
    if not text:
        rel = path.relative_to(ROOT).as_posix()
        print(f"FAIL: reuse decision document is empty: {rel}", file=sys.stderr)
        return 1

    errors = validate_reuse_decision(text)
    if errors:
        rel = path.relative_to(ROOT).as_posix()
        print(f"FAIL: reuse decision startup gate failed for {rel}", file=sys.stderr)
        for error in errors:
            print(f"- {error}", file=sys.stderr)
        return 1

    print("PASS: reuse decision startup gate passed")
    print(f"- {path.relative_to(ROOT).as_posix()}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
