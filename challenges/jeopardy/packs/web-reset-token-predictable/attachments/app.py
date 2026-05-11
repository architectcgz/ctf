import os, random, time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

FLAG = "flag{web_reset_token_predictable}"
ISSUED = {}

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        parsed = urlparse(self.path)
        params = parse_qs(parsed.query)
        if parsed.path == "/reset/request":
            user = params.get("user", ["guest"])[0]
            issued = int(time.time())
            token = random.Random(issued).randint(100000, 999999)
            ISSUED[user] = (issued, token)
            self._send(200, f"issued={issued}")
            return
        if parsed.path == "/reset/confirm":
            user = params.get("user", ["guest"])[0]
            token = int(params.get("token", ["0"])[0])
            item = ISSUED.get(user)
            if item and token == item[1] and user == "admin":
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
