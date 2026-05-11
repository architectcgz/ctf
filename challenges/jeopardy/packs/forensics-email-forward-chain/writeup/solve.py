#!/usr/bin/env python3
from __future__ import annotations

from email import policy
from email.parser import BytesParser
from pathlib import Path
import re


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    raw = (root / "attachments" / "thread.eml").read_bytes()
    msg = BytesParser(policy=policy.default).parsebytes(raw)
    text_parts = []
    for part in msg.walk():
        if part.get_content_maintype() == "multipart":
            continue
        payload = part.get_payload(decode=True) or b""
        text_parts.append(payload.decode("utf-8", errors="ignore"))
    text = "\n".join(text_parts)
    match = re.search(r"flag\{[a-z0-9_\-]+\}", text)
    if not match:
        raise SystemExit("flag not found")
    print(match.group(0))


if __name__ == "__main__":
    main()
