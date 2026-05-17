#!/usr/bin/env python3
from __future__ import annotations

import base64
import os
from pathlib import Path
import re


def load_source_text() -> str:
    import urllib.request

    base_url = os.getenv('BASE_URL', '').strip().rstrip('/')
    if base_url:
        return urllib.request.urlopen(base_url).read().decode('utf-8')

    pack_dir = Path(__file__).resolve().parents[1]
    return (pack_dir / 'statement.md').read_text(encoding='utf-8')


def main() -> None:
    text = load_source_text()
    values = re.findall(r'var\s+\w+="([A-Za-z0-9+/=]+)";', text)
    final = values[-1]
    print(base64.b64decode(base64.b64decode(final)).decode())


if __name__ == '__main__':
    main()
