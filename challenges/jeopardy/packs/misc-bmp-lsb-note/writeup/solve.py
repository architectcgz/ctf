#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    data = (root / "attachments" / "cover.bmp").read_bytes()
    offset = int.from_bytes(data[10:14], "little")
    width = int.from_bytes(data[18:22], "little", signed=True)
    row_size = ((width * 3 + 3) // 4) * 4
    row = data[offset : offset + row_size]
    bits = "".join(str(row[idx * 3] & 1) for idx in range(width))
    out = bytearray()
    for idx in range(0, len(bits), 8):
        byte = int(bits[idx : idx + 8], 2)
        if byte == 0:
            break
        out.append(byte)
    print(out.decode())


if __name__ == "__main__":
    main()
