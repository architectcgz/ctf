#!/usr/bin/env python3
from __future__ import annotations

import base64
import re
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    data = (root / "attachments" / "dns.pcap").read_bytes()
    parts = re.findall(rb"\x02(\d\d)[\x01-\x0a]([A-Z2-7]{1,10})\x05exfil\x03lab\x00", data)
    ordered = [chunk.decode() for _, chunk in sorted(parts, key=lambda item: item[0])]
    joined = "".join(ordered)
    padded = joined + "=" * ((8 - len(joined) % 8) % 8)
    print(base64.b32decode(padded).decode())


if __name__ == "__main__":
    main()
