#!/usr/bin/env python3
from __future__ import annotations

import os
import re
import urllib.parse
import urllib.request


def main() -> None:
    base_url = os.getenv('BASE_URL', 'http://127.0.0.1:8080').rstrip('/')
    body = urllib.request.urlopen(f'{base_url}/').read().decode()
    blocks = re.findall(r'<pre>(.*?)</pre>', body, re.S)
    known_plain = blocks[0].encode()
    known_cipher = bytes.fromhex(blocks[1].strip())
    secret_cipher = bytes.fromhex(blocks[2].strip())
    keystream = bytes(a ^ b for a, b in zip(known_plain, known_cipher))
    code = bytes(a ^ b for a, b in zip(secret_cipher, keystream)).decode()
    redeem = urllib.request.urlopen(
        f'{base_url}/redeem?code={urllib.parse.quote(code)}'
    ).read().decode()
    match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', redeem)
    if not match:
        raise SystemExit('flag not found')
    print(match.group(0))


if __name__ == '__main__':
    main()
