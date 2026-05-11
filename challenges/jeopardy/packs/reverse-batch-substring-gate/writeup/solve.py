#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "gate.bat").read_text(encoding="utf-8")
    pool = re.search(r"set pool=(.*)", text).group(1)
    parts = re.findall(r"%pool:~(\d+),(\d+)%", text)
    out = "".join(pool[int(start): int(start) + int(length)] for start, length in parts)
    print(out)


if __name__ == "__main__":
    main()
