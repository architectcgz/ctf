#!/usr/bin/env python3
from __future__ import annotations

import base64
import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / 'attachments' / 'page.html').read_text(encoding='utf-8')
    encoded = re.search(r'backup-note:\s*([A-Za-z0-9+/=]+)', text).group(1)
    print(base64.b64decode(encoded).decode())


if __name__ == '__main__':
    main()
