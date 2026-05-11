#!/usr/bin/env python3
from __future__ import annotations

import io
import re
import tarfile
from pathlib import Path
from zipfile import ZipFile


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    with ZipFile(root / "attachments" / "image.zip") as zf:
        layer = zf.read("layer1.tar.gz")
    with tarfile.open(fileobj=io.BytesIO(layer), mode="r:gz") as tf:
        data = tf.extractfile("app/.env").read().decode("utf-8")
    match = re.search(r"flag\{[a-z0-9_\-]+\}", data)
    if not match:
        raise SystemExit("flag not found")
    print(match.group(0))


if __name__ == "__main__":
    main()
