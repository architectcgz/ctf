#!/usr/bin/env python3
from __future__ import annotations

import itertools
import zlib


TARGET = 0x176fb4f3


def main() -> None:
    alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    for chars in itertools.product(alphabet, repeat=4):
        word = "".join(chars)
        if zlib.crc32(word.encode()) & 0xFFFFFFFF == TARGET:
            print(f"flag{{{word}}}")
            return
    raise SystemExit("not found")


if __name__ == "__main__":
    main()
