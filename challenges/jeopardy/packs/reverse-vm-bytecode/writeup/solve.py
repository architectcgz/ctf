#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    program = (root / "attachments" / "program.vmb").read_bytes()
    stack = []
    out = bytearray()
    pc = 0
    while pc < len(program):
        op = program[pc]
        pc += 1
        if op == 0x01:
            stack.append(program[pc])
            pc += 1
        elif op == 0x02:
            stack[-1] ^= program[pc]
            pc += 1
        elif op == 0x03:
            out.append(stack.pop())
        elif op == 0xFF:
            break
        else:
            raise SystemExit(f"unknown opcode {op}")
    print(out.decode())


if __name__ == "__main__":
    main()
