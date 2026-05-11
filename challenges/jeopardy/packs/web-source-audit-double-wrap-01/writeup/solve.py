#!/usr/bin/env python3
from __future__ import annotations

import base64
import os
import re


def main() -> None:
    import urllib.request

    base_url = os.getenv('BASE_URL', 'http://127.0.0.1:8080').rstrip('/')
    text = urllib.request.urlopen(base_url).read().decode('utf-8')
    values = re.findall(r'var\s+\w+="([A-Za-z0-9+/=]+)";', text)
    final = values[-1]
    print(base64.b64decode(base64.b64decode(final)).decode())


if __name__ == '__main__':
    main()
