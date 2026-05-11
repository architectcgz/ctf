#!/usr/bin/env python3
from __future__ import annotations

import os
import re
import urllib.parse
import urllib.request


def main() -> None:
    base_url = os.getenv('BASE_URL', 'http://127.0.0.1:8080').rstrip('/')
    text = urllib.request.urlopen(f'{base_url}/download/briefing.md').read().decode('utf-8')
    bits = []
    for ch in text:
        if ch == '\u200b':
            bits.append('0')
        elif ch == '\u200c':
            bits.append('1')
    raw = ''.join(bits)
    out = []
    for idx in range(0, len(raw), 8):
        chunk = raw[idx:idx + 8]
        if len(chunk) == 8:
            out.append(chr(int(chunk, 2)))
    code = ''.join(out)
    body = urllib.request.urlopen(
        f'{base_url}/redeem?code={urllib.parse.quote(code)}'
    ).read().decode()
    match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', body)
    if not match:
        raise SystemExit('flag not found')
    print(match.group(0))


if __name__ == '__main__':
    main()
