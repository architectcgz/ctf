import base64, hashlib, hmac, json, os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

FLAG = "flag{web_jwt_weak_secret}"
SECRET = b"changeme123"

def b64url(data):
    return base64.urlsafe_b64encode(data).rstrip(b"=")

def sign(header, payload):
    body = b".".join([b64url(json.dumps(header).encode()), b64url(json.dumps(payload).encode())])
    sig = b64url(hmac.new(SECRET, body, hashlib.sha256).digest())
    return body.decode() + "." + sig.decode()

def verify(token):
    body, sig = token.rsplit(".", 1)
    expected = b64url(hmac.new(SECRET, body.encode(), hashlib.sha256).digest()).decode()
    if not hmac.compare_digest(sig, expected):
        return None
    payload = json.loads(base64.urlsafe_b64decode(body.split(".")[1] + "==").decode())
    return payload

GUEST = sign({"alg":"HS256","typ":"JWT"}, {"user":"guest","role":"user"})

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        parsed = urlparse(self.path)
        if parsed.path == "/token":
            self._send(200, GUEST)
            return
        if parsed.path == "/admin":
            token = parse_qs(parsed.query).get("token", [""])[0]
            payload = verify(token)
            if payload and payload.get("role") == "admin":
                self._send(200, FLAG)
            else:
                self._send(403, "denied")
            return
        self._send(404, "not found")

    def _send(self, status, body):
        data = body.encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "text/plain; charset=utf-8")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def log_message(self, *args):
        return

if __name__ == "__main__":
    port = int(os.getenv("PORT", "8080"))
    ThreadingHTTPServer(("127.0.0.1", port), Handler).serve_forever()
