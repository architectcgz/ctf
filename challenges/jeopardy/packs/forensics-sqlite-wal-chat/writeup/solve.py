#!/usr/bin/env python3
from __future__ import annotations

import re
import shutil
import sqlite3
import tempfile
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with tempfile.TemporaryDirectory() as tmp_dir:
        tmp = Path(tmp_dir)
        for name in ("chat.db", "chat.db-wal", "chat.db-shm"):
            shutil.copy2(root / "attachments" / name, tmp / name)
        conn = sqlite3.connect(tmp / "chat.db")
        rows = conn.execute("SELECT sender, body FROM messages").fetchall()
        conn.close()
    text = "\n".join(f"{sender}: {body}" for sender, body in rows)
    match = re.search(r"flag\{[a-z0-9_\-]+\}", text)
    if not match:
        raise SystemExit("flag not found")
    print(match.group(0))


if __name__ == "__main__":
    main()
