#!/usr/bin/env python3
from __future__ import annotations

import re
import struct
from pathlib import Path


def unshift_right_xor(value: int, shift: int) -> int:
    result = 0
    for bit in range(32):
        src = 31 - bit
        shifted = src + shift
        known = ((result >> shifted) & 1) if shifted < 32 else 0
        current = ((value >> src) & 1) ^ known
        result |= current << src
    return result


def unshift_left_xor_mask(value: int, shift: int, mask: int) -> int:
    result = 0
    for bit in range(32):
        shifted = bit - shift
        known = ((result >> shifted) & 1) if shifted >= 0 else 0
        mask_bit = (mask >> bit) & 1
        current = ((value >> bit) & 1) ^ (known & mask_bit)
        result |= current << bit
    return result


def untemper(value: int) -> int:
    value = unshift_right_xor(value, 18)
    value = unshift_left_xor_mask(value, 15, 0xEFC60000)
    value = unshift_left_xor_mask(value, 7, 0x9D2C5680)
    value = unshift_right_xor(value, 11)
    return value & 0xFFFFFFFF


class MiniMT:
    def __init__(self, state):
        self.state = list(state)
        self.index = 624

    def twist(self):
        for idx in range(624):
            y = (self.state[idx] & 0x80000000) + (self.state[(idx + 1) % 624] & 0x7FFFFFFF)
            self.state[idx] = self.state[(idx + 397) % 624] ^ (y >> 1)
            if y & 1:
                self.state[idx] ^= 0x9908B0DF
        self.index = 0

    def rand_u32(self):
        if self.index >= 624:
            self.twist()
        y = self.state[self.index]
        self.index += 1
        y ^= y >> 11
        y ^= (y << 7) & 0x9D2C5680
        y ^= (y << 15) & 0xEFC60000
        y ^= y >> 18
        return y & 0xFFFFFFFF


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "reset.txt").read_text(encoding="utf-8")
    output_part, cipher_part = text.split("cipher_hex:\n", 1)
    outputs = [int(line) for line in output_part.splitlines() if line.isdigit()]
    cipher = bytes.fromhex(cipher_part.strip())
    state = [untemper(value) for value in outputs[:624]]
    mt = MiniMT(state)
    stream = bytearray()
    while len(stream) < len(cipher):
        stream.extend(struct.pack(">I", mt.rand_u32()))
    plain = bytes(byte ^ stream[idx] for idx, byte in enumerate(cipher))
    match = re.search(rb"flag\{[a-z0-9_\-]+\}", plain)
    if not match:
        raise SystemExit("flag not found")
    print(match.group().decode())


if __name__ == "__main__":
    main()
