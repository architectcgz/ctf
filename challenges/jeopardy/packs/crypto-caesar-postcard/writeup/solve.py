#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    cipher = (root / 'attachments' / 'postcard.txt').read_text(encoding='utf-8').strip()
    out = []
    for ch in cipher:
        if 'a' <= ch <= 'z':
            out.append(chr((ord(ch) - ord('a') - 7) % 26 + ord('a')))
        else:
            out.append(ch)
    print(''.join(out))


if __name__ == '__main__':
    main()
