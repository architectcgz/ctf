#!/usr/bin/env python3
from __future__ import annotations

import tarfile
import tempfile
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with tempfile.TemporaryDirectory() as tmp_dir:
        with tarfile.open(root / "attachments" / "tree.tar.gz", "r:gz") as tf:
            tf.extractall(tmp_dir)
        for path in Path(tmp_dir).rglob("*"):
            if path.is_file() and "vault" in path.parts and (path.stat().st_mode & 0o111):
                text = path.read_text(encoding="utf-8")
                if text.startswith("flag{"):
                    print(text)
                    return
    raise SystemExit("flag not found")


if __name__ == "__main__":
    main()
