#!/usr/bin/env python3
from __future__ import annotations

import base64
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    data = (root / "attachments" / "note.png").read_bytes()
    cursor = 8
    while cursor < len(data):
        length = int.from_bytes(data[cursor:cursor+4], "big")
        ctype = data[cursor+4:cursor+8]
        body = data[cursor+8:cursor+8+length]
        cursor += 12 + length
        if ctype == b"tEXt":
            _, value = body.split(b"=", 1)
            print(base64.b64decode(value).decode())
            return
    raise SystemExit("tEXt not found")


if __name__ == "__main__":
    main()
