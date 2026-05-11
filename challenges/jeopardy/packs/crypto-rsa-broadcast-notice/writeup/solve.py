#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


def crt(ns, cs):
    total = 0
    big_n = 1
    for n in ns:
        big_n *= n
    for n, c in zip(ns, cs):
        m = big_n // n
        total += c * m * pow(m, -1, n)
    return total % big_n


def icbrt(value: int) -> int:
    low, high = 0, value
    while low <= high:
        mid = (low + high) // 2
        cube = mid * mid * mid
        if cube == value:
            return mid
        if cube < value:
            low = mid + 1
        else:
            high = mid - 1
    return high


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    values = {}
    for line in (root / "attachments" / "broadcast.txt").read_text(encoding="utf-8").splitlines():
        key, value = line.split("=", 1)
        values[key] = int(value)
    ns = [values["n1"], values["n2"], values["n3"]]
    cs = [values["c1"], values["c2"], values["c3"]]
    cube = crt(ns, cs)
    m = icbrt(cube)
    data = m.to_bytes((m.bit_length() + 7) // 8, "big")
    print(f"flag{{{data.decode()}}}")


if __name__ == "__main__":
    main()
