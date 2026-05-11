#!/usr/bin/env python3
from __future__ import annotations

import marshal
import os
import re
import urllib.parse
import urllib.request


def main() -> None:
    base_url = os.getenv('BASE_URL', 'http://127.0.0.1:8080').rstrip('/')
    pyc = urllib.request.urlopen(f'{base_url}/download/cache_gate.pyc').read()
    code = marshal.loads(pyc[16:])
    key = next(item for item in code.co_consts if isinstance(item, int))
    prefix = next(item for item in code.co_consts if isinstance(item, str) and item)
    encoded = next(item for item in code.co_consts if isinstance(item, tuple))
    redeem_code = prefix + ''.join(chr(value ^ key) for value in encoded)
    body = urllib.request.urlopen(
        f'{base_url}/redeem?code={urllib.parse.quote(redeem_code)}'
    ).read().decode()
    match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', body)
    if not match:
        raise SystemExit('flag not found')
    print(match.group(0))


if __name__ == '__main__':
    main()
