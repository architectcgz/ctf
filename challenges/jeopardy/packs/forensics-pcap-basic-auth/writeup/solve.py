#!/usr/bin/env python3
from __future__ import annotations

import base64
import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    data = (root / "attachments" / "capture.pcap").read_bytes()
    match = re.search(rb"Authorization: Basic ([A-Za-z0-9+/=]+)", data)
    if not match:
        raise SystemExit("auth header not found")
    plain = base64.b64decode(match.group(1)).decode()
    print(plain.split(":", 1)[1])


if __name__ == "__main__":
    main()
