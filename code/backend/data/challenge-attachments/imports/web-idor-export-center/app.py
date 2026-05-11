import json, os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

EXPORTS = {
    "1001": {"owner": "alice", "body": "report one"},
    "1002": {"owner": "alice", "body": "report two"},
    "9001": {"owner": "admin", "body": "flag{web_idor_export_center}" },
}

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        parsed = urlparse(self.path)
        if parsed.path == "/":
            self._send(200, "your exports: 1001,1002")
            return
        if parsed.path == "/export":
            item = parse_qs(parsed.query).get("id", [""])[0]
            record = EXPORTS.get(item)
            if not record:
                self._send(404, "missing")
                return
            self._send(200, json.dumps(record))
            return
        self._send(404, "not found")

    def _send(self, status, body):
        data = body.encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def log_message(self, *args):
        return

if __name__ == "__main__":
    port = int(os.getenv("PORT", "8080"))
    ThreadingHTTPServer(("127.0.0.1", port), Handler).serve_forever()
