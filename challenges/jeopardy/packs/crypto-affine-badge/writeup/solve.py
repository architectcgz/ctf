#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def egcd(a: int, b: int):
    if b == 0:
        return a, 1, 0
    g, x1, y1 = egcd(b, a % b)
    return g, y1, x1 - (a // b) * y1


def modinv(a: int, m: int) -> int:
    g, x, _ = egcd(a, m)
    if g != 1:
        raise ValueError
    return x % m


def decode(text: str, a: int, b: int) -> str:
    inv_a = modinv(a, 26)
    out = []
    for ch in text:
        if "a" <= ch <= "z":
            value = ord(ch) - ord("a")
            out.append(chr(((inv_a * (value - b)) % 26) + ord("a")))
        else:
            out.append(ch)
    return "".join(out)


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    cipher = (root / "attachments" / "badge.txt").read_text(encoding="utf-8")
    for a in range(1, 26):
        if a % 2 == 0 or a == 13:
            continue
        for b in range(26):
            plain = decode(cipher, a, b)
            match = re.search(r"flag\{[a-z0-9_\-]+\}", plain)
            if match:
                print(match.group(0))
                return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
