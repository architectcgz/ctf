import os, re
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

FLAG = "flag{web_ssti_render_lab}"

def render(tpl: str) -> str:
    def repl(match):
        expr = match.group(1)
        return str(eval(expr, {"__builtins__": {}}, {"flag": FLAG, "site": "ssti-lab"}))
    return re.sub(r"\x7b\x7b(.*?)\x7d\x7d", repl, tpl)

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        parsed = urlparse(self.path)
        if parsed.path == "/":
            self._send(200, "use /render?template=hello")
            return
        if parsed.path == "/render":
            tpl = parse_qs(parsed.query).get("template", [""])[0]
            self._send(200, render(tpl))
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
