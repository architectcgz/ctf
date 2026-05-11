#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "check.js").read_text(encoding="utf-8")
    raw = re.search(r"const data = \[(.*?)\];", text, re.S).group(1)
    values = [int(item.strip(), 16) for item in raw.split(",")]
    print("".join(chr(value ^ 0x13) for value in values))


if __name__ == "__main__":
    main()
