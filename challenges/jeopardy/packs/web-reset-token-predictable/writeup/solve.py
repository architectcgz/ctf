#!/usr/bin/env python3
from __future__ import annotations

import os, random, socket, subprocess, sys, time, urllib.request
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
    proc = start(root / "attachments" / "app.py", 18091)
    try:
        issued_text = urllib.request.urlopen("http://127.0.0.1:18091/reset/request?user=admin").read().decode()
        issued = int(issued_text.split("=", 1)[1])
        token = random.Random(issued).randint(100000, 999999)
        body = urllib.request.urlopen(f"http://127.0.0.1:18091/reset/confirm?user=admin&token={token}").read().decode()
        print(body.strip())
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
