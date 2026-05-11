#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "check.ps1").read_text(encoding="utf-8")
    key = int(re.search(r"\$key = (\d+)", text).group(1))
    numbers = [int(item) for item in re.search(r"@\(([^)]*)\)", text, re.S).group(1).split(",")]
    print(bytes(value ^ key for value in numbers).decode())


if __name__ == "__main__":
    main()
