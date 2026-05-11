#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    lines = (root / "attachments" / "session.cast").read_text(encoding="utf-8").splitlines()[1:]
    output = "".join(json.loads(line)[2] for line in lines)
    match = re.search(r"flag\{[a-z0-9_\-]+\}", output)
    if not match:
        raise SystemExit("flag not found")
    print(match.group(0))


if __name__ == "__main__":
    main()
