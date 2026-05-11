#!/usr/bin/env python3
from __future__ import annotations

import subprocess
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    binary = root / "attachments" / "challenge.bin"
    for value in range(1000, 10000):
        proc = subprocess.run([binary], input=f"{value}\n", text=True, capture_output=True)
        if "ok" in proc.stdout:
            print(f"flag{{{value}}}")
            return
    raise SystemExit("serial not found")


if __name__ == "__main__":
    main()
