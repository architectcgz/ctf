#!/usr/bin/env python3
from __future__ import annotations

import re
import tarfile
import tempfile
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with tempfile.TemporaryDirectory() as tmp_dir:
        with tarfile.open(root / "attachments" / "cron.tar.gz", "r:gz") as tf:
            tf.extractall(tmp_dir)
        script = (Path(tmp_dir) / "usr/local/bin/night-export.sh").read_text(encoding="utf-8")
    match = re.search(r"flag\{[a-z0-9_\-]+\}", script)
    if not match:
        raise SystemExit("flag not found")
    print(match.group(0))


if __name__ == "__main__":
    main()
