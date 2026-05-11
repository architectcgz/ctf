#!/usr/bin/env python3
from __future__ import annotations

import re
import sqlite3
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    conn = sqlite3.connect(root / "attachments" / "History.db")
    cur = conn.cursor()
    rows = cur.execute("SELECT url, title FROM urls").fetchall()
    conn.close()
    text = "\n".join(f"{url} {title}" for url, title in rows)
    match = re.search(r"flag\{[a-z0-9_\-]+\}", text)
    if not match:
        raise SystemExit("flag not found")
    print(match.group(0))


if __name__ == "__main__":
    main()
