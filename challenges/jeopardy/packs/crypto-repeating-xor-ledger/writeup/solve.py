#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def hamming(left: bytes, right: bytes) -> int:
    return sum((a ^ b).bit_count() for a, b in zip(left, right))


def score(data: bytes) -> float:
    freq = {
        ord(" "): 13.0,
        ord("e"): 12.7,
        ord("t"): 9.1,
        ord("a"): 8.2,
        ord("o"): 7.5,
        ord("i"): 7.0,
        ord("n"): 6.7,
        ord("s"): 6.3,
        ord("h"): 6.1,
        ord("r"): 6.0,
    }
    total = 0.0
    for byte in data.lower():
        if 32 <= byte <= 126:
            total += freq.get(byte, 0.25)
        else:
            total -= 12.0
    return total


def recover_keysizes(data: bytes) -> list[int]:
    scored = []
    for keysize in range(2, 13):
        blocks = [data[idx : idx + keysize] for idx in range(0, keysize * 4, keysize)]
        if len(blocks[-1]) != keysize:
            continue
        pairs = []
        for idx in range(len(blocks) - 1):
            pairs.append(hamming(blocks[idx], blocks[idx + 1]) / keysize)
        scored.append((sum(pairs) / len(pairs), keysize))
    scored.sort()
    return [keysize for _, keysize in scored]


def recover_key(data: bytes, keysize: int) -> bytes:
    key = bytearray()
    for offset in range(keysize):
        column = data[offset::keysize]
        best = max(
            ((score(bytes(byte ^ candidate for byte in column)), candidate) for candidate in range(256)),
            key=lambda item: item[0],
        )
        key.append(best[1])
    return bytes(key)


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    cipher = (root / "attachments" / "ledger.bin").read_bytes()
    for keysize in recover_keysizes(cipher):
        key = recover_key(cipher, keysize)
        plain = bytes(byte ^ key[idx % len(key)] for idx, byte in enumerate(cipher))
        match = re.search(rb"flag\{[a-z0-9_\-]+\}", plain)
        if match:
            print(match.group().decode())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
