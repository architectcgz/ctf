import os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

FLAG = "flag{web_workflow_step_bypass}"

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        parsed = urlparse(self.path)
        if parsed.path == "/":
            self._send(200, "step1 at /step1")
            return
        if parsed.path == "/step1":
            self._send(200, "submit ticket=ready to /step2")
            return
        if parsed.path == "/step2":
            self._send(200, "final page is /final?ok=1")
            return
        if parsed.path == "/final":
            ok = parse_qs(parsed.query).get("ok", ["0"])[0]
            if ok == "1":
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
