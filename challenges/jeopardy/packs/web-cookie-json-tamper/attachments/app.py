import base64, json, os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

FLAG = "flag{web_cookie_json_tamper}"

def encode(obj):
    return base64.urlsafe_b64encode(json.dumps(obj).encode()).decode()

def decode(value):
    return json.loads(base64.urlsafe_b64decode(value.encode()).decode())

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/":
            cookie = self.headers.get("Cookie", "")
            if "session=" not in cookie:
                token = encode({"user":"guest","role":"user"})
                self.send_response(200)
                self.send_header("Set-Cookie", f"session={token}; Path=/")
                self.end_headers()
                self.wfile.write(b"guest page")
                return
            self._send(200, "guest page")
            return
        if self.path == "/admin":
            cookie = self.headers.get("Cookie", "")
            token = cookie.split("session=", 1)[1].split(";", 1)[0]
            session = decode(token)
            if session.get("role") == "admin":
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
