#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "check.sh").read_text(encoding="utf-8")
    pool = re.search(r'pool="([^"]+)"', text).group(1)
    left = pool[2:]
    right = left[:-2]
    print(right)


if __name__ == "__main__":
    main()
