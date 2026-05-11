#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path
import re


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "checker.txt").read_text(encoding="utf-8")
    raw = re.search(r"expected = \[(.*?)\]", text).group(1)
    nums = [int(item.strip(), 16) for item in raw.split(",")]
    print("flag{" + "".join(f"{value:02x}" for value in nums) + "}")


if __name__ == "__main__":
    main()
