#!/usr/bin/env python3
from __future__ import annotations

import base64
import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / 'attachments' / 'source.html').read_text(encoding='utf-8')
    values = re.findall(r'var\s+\w+="([A-Za-z0-9+/=]+)";', text)
    final = values[-1]
    print(base64.b64decode(base64.b64decode(final)).decode())


if __name__ == '__main__':
    main()
