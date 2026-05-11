#!/usr/bin/env python3
from __future__ import annotations

import tarfile
import tempfile
from pathlib import Path


def load_env(path: Path) -> dict[str, str]:
    values = {}
    for line in path.read_text(encoding="utf-8").splitlines():
        if "=" in line:
            key, value = line.split("=", 1)
            values[key] = value
    return values


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with tempfile.TemporaryDirectory() as tmp_dir:
        with tarfile.open(root / "attachments" / "envpack.tar.gz", "r:gz") as tf:
            tf.extractall(tmp_dir)
        tmp = Path(tmp_dir)
        merged = {}
        for name in (".env", "service.env", "compose.env"):
            merged.update(load_env(tmp / name))
    print(merged["EXPORT_CODE"])


if __name__ == "__main__":
    main()
