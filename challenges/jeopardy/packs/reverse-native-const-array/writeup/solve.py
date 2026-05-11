#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    values = [int(line.strip()) for line in (root / "attachments" / "constants.txt").read_text(encoding="utf-8").splitlines() if line.strip()]
    text = "".join(chr(value - idx) for idx, value in enumerate(values))
    print(text)


if __name__ == "__main__":
    main()
