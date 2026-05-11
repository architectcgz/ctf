#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


def egcd(a: int, b: int):
    if b == 0:
        return a, 1, 0
    g, x1, y1 = egcd(b, a % b)
    return g, y1, x1 - (a // b) * y1


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    values = {}
    for line in (root / "attachments" / "common.txt").read_text(encoding="utf-8").splitlines():
        key, value = line.split("=", 1)
        values[key] = int(value)
    n = values["n"]
    e1 = values["e1"]
    e2 = values["e2"]
    c1 = values["c1"]
    c2 = values["c2"]
    _, s, t = egcd(e1, e2)
    if s < 0:
        c1 = pow(c1, -1, n)
        s = -s
    if t < 0:
        c2 = pow(c2, -1, n)
        t = -t
    m = (pow(c1, s, n) * pow(c2, t, n)) % n
    data = m.to_bytes((m.bit_length() + 7) // 8, "big")
    print(f"flag{{{data.decode()}}}")


if __name__ == "__main__":
    main()
