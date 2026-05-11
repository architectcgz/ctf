#!/usr/bin/env python3
from __future__ import annotations

import re
import shutil
import subprocess
import tarfile
import tempfile
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with tempfile.TemporaryDirectory() as tmp_dir:
        tar_path = root / "attachments" / "repo.tar.gz"
        with tarfile.open(tar_path, "r:gz") as tf:
            tf.extractall(tmp_dir)
        repo = Path(tmp_dir)
        output = subprocess.check_output(["git", "stash", "show", "-p"], cwd=repo, text=True)
        match = re.search(r"flag\{[a-z0-9_\-]+\}", output)
        if not match:
            raise SystemExit("flag not found")
        print(match.group(0))


if __name__ == "__main__":
    main()
