#!/usr/bin/env python3
from __future__ import annotations

import tarfile
import tempfile
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with tempfile.TemporaryDirectory() as tmp_dir:
        with tarfile.open(root / "attachments" / "sample.tar.gz", "r:gz") as tf:
            tf.extractall(tmp_dir)
        print((Path(tmp_dir) / "-").read_text(encoding="utf-8").strip())


if __name__ == "__main__":
    main()
