#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def decrypt(cipher: str, width: int) -> str:
    rows = len(cipher) // width
    extra = len(cipher) % width
    lengths = [rows + (1 if idx < extra else 0) for idx in range(width)]
    cols = []
    cursor = 0
    for length in lengths:
        cols.append(cipher[cursor : cursor + length])
        cursor += length
    out = []
    for row in range(max(lengths)):
        for col in range(width):
            if row < len(cols[col]):
                out.append(cols[col][row])
    return "".join(out)


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    cipher = (root / "attachments" / "archive.txt").read_text(encoding="utf-8")
    for width in range(2, 9):
        plain = decrypt(cipher, width)
        match = re.search(r"flag\{[a-z0-9_\-]+\}", plain)
        if match:
            print(match.group(0))
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
