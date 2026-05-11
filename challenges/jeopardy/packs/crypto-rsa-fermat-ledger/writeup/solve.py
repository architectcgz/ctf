#!/usr/bin/env python3
from __future__ import annotations

from math import isqrt
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    values = {}
    for line in (root / "attachments" / "rsa.txt").read_text(encoding="utf-8").splitlines():
        key, value = line.split("=", 1)
        values[key] = int(value)
    n = values["n"]
    e = values["e"]
    c = values["c"]
    a = isqrt(n)
    if a * a < n:
        a += 1
    while True:
        b2 = a * a - n
        b = isqrt(b2)
        if b * b == b2:
            p = a - b
            q = a + b
            break
        a += 1
    phi = (p - 1) * (q - 1)
    d = pow(e, -1, phi)
    m = pow(c, d, n)
    data = m.to_bytes((m.bit_length() + 7) // 8, "big")
    print(f"flag{{{data.decode()}}}")


if __name__ == "__main__":
    main()
