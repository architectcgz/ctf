#!/usr/bin/env python3
from __future__ import annotations

import ast
import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / 'attachments' / 'checker.py').read_text(encoding='utf-8')
    raw = re.search(r'data\s*=\s*(\[[^\]]+\])', text, re.S).group(1)
    data = ast.literal_eval(raw)
    print(''.join(chr(value ^ 0x7A) for value in data))


if __name__ == '__main__':
    main()
