#!/usr/bin/env python3
from __future__ import annotations

import subprocess
from pathlib import Path


RECORD_SIZE = 7


def ror8(value: int, bits: int) -> int:
    bits %= 8
    return ((value >> bits) | ((value << (8 - bits)) & 0xFF)) & 0xFF


def decode_record(raw: bytes) -> tuple[int, int, int, int, int, int, int]:
    block_id = raw[0]
    decoded = [block_id]
    for pos, item in enumerate(raw[1:], start=1):
        mask = (block_id * 29 + pos * 11 + 0x5A) & 0xFF
        decoded.append(item ^ mask)
    return tuple(decoded)  # type: ignore[return-value]


def recover_flag(program_path: Path) -> str:
    data = program_path.read_bytes()
    if len(data) < 7 or data[:4] != b"BVM1":
        raise SystemExit("invalid blob header")

    entry = data[4]
    state = data[5]
    count = data[6]

    records = {}
    offset = 7
    for _ in range(count):
        chunk = data[offset : offset + RECORD_SIZE]
        if len(chunk) != RECORD_SIZE:
            raise SystemExit("truncated blob")
        record = decode_record(chunk)
        records[record[0]] = record
        offset += RECORD_SIZE

    output = [0] * count
    current = entry
    for _ in range(count):
        block_id, next_block, index, xor_key, add_key, rol_bits, expected = records[current]
        value = ror8(expected, rol_bits)
        value = (value - add_key) & 0xFF
        value ^= xor_key
        value ^= state
        output[index] = value
        state = expected ^ index ^ 0xA5
        current = next_block

    return bytes(output).decode("utf-8")


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    program = root / "attachments" / "program.blk"
    flag = recover_flag(program)

    proc = subprocess.run(
        [binary, program],
        input=flag + "\n",
        text=True,
        capture_output=True,
    )
    if "ok" not in proc.stdout:
        raise SystemExit("recovered input did not validate")
    print(flag)


if __name__ == "__main__":
    main()
