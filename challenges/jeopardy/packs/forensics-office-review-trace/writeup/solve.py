#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path
from zipfile import ZipFile


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with ZipFile(root / "attachments" / "review.docx") as zf:
        text = zf.read("word/comments.xml").decode("utf-8")
    match = re.search(r"flag\{[a-z0-9_\-]+\}", text)
    if not match:
        raise SystemExit("flag not found")
    print(match.group(0))


if __name__ == "__main__":
    main()
