#!/usr/bin/env python3
from __future__ import annotations

import marshal
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
    proc = start(root / 'docker' / 'app.py', 18105)
    try:
        pyc = urllib.request.urlopen('http://127.0.0.1:18105/download/cache_gate.pyc').read()
        code = marshal.loads(pyc[16:])
        key = next(item for item in code.co_consts if isinstance(item, int))
        prefix = next(item for item in code.co_consts if isinstance(item, str) and item)
        encoded = next(item for item in code.co_consts if isinstance(item, tuple))
        redeem_code = prefix + ''.join(chr(value ^ key) for value in encoded)
        body = urllib.request.urlopen(f'http://127.0.0.1:18105/redeem?code={urllib.parse.quote(redeem_code)}').read().decode()
        match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', body)
        if not match:
            raise SystemExit('flag not found')
        print(match.group(0))
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == '__main__':
    main()
