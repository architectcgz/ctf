#!/usr/bin/env python3
from __future__ import annotations

import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    data = (root / "attachments" / "snapshot.bin").read_bytes()
    match = re.search(rb"flag\{[a-z0-9_\-]+\}", data)
    if not match:
        raise SystemExit("flag not found")
    print(match.group().decode())


if __name__ == "__main__":
    main()
