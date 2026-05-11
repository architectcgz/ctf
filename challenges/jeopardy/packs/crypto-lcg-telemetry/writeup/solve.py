#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


MOD = 2147483647


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
    text = (root / "attachments" / "telemetry.txt").read_text(encoding="utf-8")
    sample_part, cipher_part = text.split("cipher_hex:\n", 1)
    samples = [int(line) for line in sample_part.splitlines() if line.isdigit()]
    cipher = bytes.fromhex(cipher_part.strip())
    a = ((samples[2] - samples[1]) * modinv((samples[1] - samples[0]) % MOD, MOD)) % MOD
    c = (samples[1] - a * samples[0]) % MOD
    state = samples[-1]
    stream = []
    for _ in range(len(cipher)):
        state = (a * state + c) % MOD
        stream.append(state & 0xFF)
    plain = bytes(byte ^ key for byte, key in zip(cipher, stream))
    match = re.search(rb"flag\{[a-z0-9_\-]+\}", plain)
    if not match:
        raise SystemExit("flag not found")
    print(match.group().decode())


if __name__ == "__main__":
    main()
