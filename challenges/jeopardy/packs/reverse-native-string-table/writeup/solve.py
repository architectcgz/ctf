#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    values = [int(line) for line in (root / "attachments" / "table.txt").read_text(encoding="utf-8").splitlines() if line.strip()]
    key = int((root / "attachments" / "key.txt").read_text(encoding="utf-8").strip())
    print(bytes(value ^ key for value in values).decode())


if __name__ == "__main__":
    main()
