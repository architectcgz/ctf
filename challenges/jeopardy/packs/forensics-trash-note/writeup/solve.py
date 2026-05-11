#!/usr/bin/env python3
from __future__ import annotations

import re
import tarfile
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with tarfile.open(root / 'attachments' / 'evidence.tar') as tf:
        data = tf.extractfile('evidence/.trash/deleted-note.txt').read().decode()
    match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', data)
    if not match:
        raise SystemExit('flag not found')
    print(match.group(0))


if __name__ == '__main__':
    main()
