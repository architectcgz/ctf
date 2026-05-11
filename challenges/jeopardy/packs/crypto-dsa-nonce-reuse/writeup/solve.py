#!/usr/bin/env python3
from __future__ import annotations

import zlib
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


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    values = {}
    for line in (root / "attachments" / "signatures.txt").read_text(encoding="utf-8").splitlines():
        key, value = line.split("=", 1)
        values[key] = value
    q = int(values["q"])
    h1 = int(values["h1"])
    h2 = int(values["h2"])
    r = int(values["r"])
    s1 = int(values["s1"])
    s2 = int(values["s2"])
    cipher = bytes.fromhex(values["cipher_hex"])
    k = ((h1 - h2) * modinv((s1 - s2) % q, q)) % q
    x = ((s1 * k - h1) * modinv(r, q)) % q
    key = zlib.crc32(str(x).encode("utf-8")).to_bytes(4, "big")
    plain = bytes(byte ^ key[idx % len(key)] for idx, byte in enumerate(cipher))
    print(plain.decode())


if __name__ == "__main__":
    main()
