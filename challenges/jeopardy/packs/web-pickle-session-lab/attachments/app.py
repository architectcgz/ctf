import base64, os, pickle
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

FLAG = "flag{web_pickle_session_lab}"
BASE = os.path.dirname(__file__)
open(os.path.join(BASE, "flag.txt"), "w", encoding="utf-8").write(FLAG + "\n")

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        parsed = urlparse(self.path)
        if parsed.path == "/":
            self._send(200, "use /restore?blob=...")
            return
        if parsed.path == "/restore":
            blob = parse_qs(parsed.query).get("blob", [""])[0]
            obj = pickle.loads(base64.urlsafe_b64decode(blob.encode()))
            self._send(200, str(obj))
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
