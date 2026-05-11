#!/usr/bin/env python3
from __future__ import annotations

import bz2
import gzip
import io
import tarfile
from pathlib import Path


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    text = (root / "attachments" / "dump.txt").read_text(encoding="utf-8")
    data = bytes.fromhex("".join(text.split()))
    stage1 = bz2.decompress(data)
    stage2 = gzip.decompress(stage1)
    with tarfile.open(fileobj=io.BytesIO(stage2), mode="r:gz") as tf:
        print(tf.extractfile("flag.txt").read().decode().strip())


if __name__ == "__main__":
    main()
