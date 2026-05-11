#!/usr/bin/env python3
from __future__ import annotations

import subprocess
import tarfile
import tempfile
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with tempfile.TemporaryDirectory() as tmp_dir:
        with tarfile.open(root / "attachments" / "makepack.tar.gz", "r:gz") as tf:
            tf.extractall(tmp_dir)
        output = subprocess.check_output(["make", "reveal", "ROLE=admin"], cwd=tmp_dir, text=True)
    print(output.strip())


if __name__ == "__main__":
    main()
