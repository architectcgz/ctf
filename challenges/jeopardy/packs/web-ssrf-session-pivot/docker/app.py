from __future__ import annotations

import base64
import hashlib
import hmac
import http.client
import json
import os
import urllib.error
import urllib.parse
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer


FLAG = os.environ.get("FLAG", "flag{local_web_ssrf_session_pivot}")
SESSION_SECRET = os.environ.get("SESSION_SECRET", "relay-signing-secret")
COOKIE_NAME = "session"


def b64u_encode(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).decode("utf-8").rstrip("=")


def b64u_decode(text: str) -> bytes:
    padding = "=" * ((4 - len(text) % 4) % 4)
    return base64.urlsafe_b64decode(text + padding)


def sign_value(value: str) -> str:
    return hmac.new(SESSION_SECRET.encode("utf-8"), value.encode("utf-8"), hashlib.sha256).hexdigest()


def build_token(user: str, role: str) -> str:
    payload = json.dumps({"role": role, "user": user}, separators=(",", ":"), sort_keys=True).encode("utf-8")
    encoded = b64u_encode(payload)
    return f"{encoded}.{sign_value(encoded)}"


def parse_token(token: str) -> dict[str, str] | None:
    if "." not in token:
        return None
    encoded, signature = token.split(".", 1)
    if not hmac.compare_digest(sign_value(encoded), signature):
        return None
    try:
        data = json.loads(b64u_decode(encoded).decode("utf-8"))
    except (ValueError, json.JSONDecodeError, UnicodeDecodeError):
        return None
    if not isinstance(data, dict):
        return None
    user = str(data.get("user", "guest"))
    role = str(data.get("role", "user"))
    return {"user": user, "role": role}


def fetch_remote_text(target: str) -> str:
    parsed = urllib.parse.urlsplit(target)
    if parsed.scheme != "http" or not parsed.hostname:
        raise urllib.error.URLError("unsupported target")

    path = parsed.path or "/"
    if parsed.query:
        path += "?" + parsed.query

    conn = http.client.HTTPConnection(parsed.hostname, parsed.port or 80, timeout=3)
    try:
        conn.request("GET", path)
        response = conn.getresponse()
        return response.read().decode("utf-8", errors="replace")
    finally:
        conn.close()


class Handler(BaseHTTPRequestHandler):
    def _send(self, status: int, body: str, *, content_type: str = "text/plain; charset=utf-8", headers: dict[str, str] | None = None) -> None:
        data = body.encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", str(len(data)))
        for key, value in (headers or {}).items():
            self.send_header(key, value)
        self.end_headers()
        self.wfile.write(data)

    def _cookie_value(self, name: str) -> str:
        cookie = self.headers.get("Cookie", "")
        for part in cookie.split(";"):
            part = part.strip()
            if part.startswith(name + "="):
                return part.split("=", 1)[1]
        return ""

    def _session(self) -> dict[str, str] | None:
        token = self._cookie_value(COOKIE_NAME)
        if not token:
            return None
        return parse_token(token)

    def do_GET(self) -> None:
        parsed = urllib.parse.urlparse(self.path)
        if parsed.path == "/":
            self._send(
                200,
                "Support relay online.\nLogin first, then try /support/fetch?url=http://relay.local/... .\n",
            )
            return
        if parsed.path == "/me":
            session = self._session()
            if not session:
                self._send(403, "Login required.\n")
                return
            self._send(200, json.dumps(session, ensure_ascii=False) + "\n", content_type="application/json; charset=utf-8")
            return
        if parsed.path == "/support/fetch":
            session = self._session()
            if not session:
                self._send(403, "Login required.\n")
                return
            params = urllib.parse.parse_qs(parsed.query)
            target = params.get("url", [""])[0]
            if not target.startswith("http://relay.local"):
                self._send(400, "Only relay.local is allowed.\n")
                return
            try:
                body = fetch_remote_text(target)
            except urllib.error.URLError as exc:
                self._send(502, f"Fetch failed: {exc}\n")
                return
            self._send(200, body + ("\n" if not body.endswith("\n") else ""))
            return
        if parsed.path == "/internal/secret":
            if self.client_address[0] not in {"127.0.0.1", "::1"}:
                self._send(403, "Local requests only.\n")
                return
            self._send(200, f"SESSION_SECRET={SESSION_SECRET}\n")
            return
        if parsed.path == "/admin/export":
            session = self._session()
            if not session or session.get("role") != "admin":
                self._send(403, "Admin only.\n")
                return
            self._send(200, f"Export ready.\n{FLAG}\n")
            return
        self._send(404, "Not found.\n")

    def do_POST(self) -> None:
        parsed = urllib.parse.urlparse(self.path)
        if parsed.path != "/login":
            self._send(404, "Not found.\n")
            return
        length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(length).decode("utf-8")
        params = urllib.parse.parse_qs(body)
        username = params.get("username", ["guest"])[0] or "guest"
        token = build_token(username, "user")
        self._send(
            200,
            f"Logged in as {username}.\n",
            headers={"Set-Cookie": f"{COOKIE_NAME}={token}; Path=/; HttpOnly"},
        )

    def log_message(self, fmt: str, *args: object) -> None:
        return


if __name__ == "__main__":
    ThreadingHTTPServer(("0.0.0.0", 8080), Handler).serve_forever()
