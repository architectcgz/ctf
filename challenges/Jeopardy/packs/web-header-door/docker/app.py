from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import os


FLAG = os.environ.get("FLAG", "flag{web_header_door_opened}")


class Handler(BaseHTTPRequestHandler):
    def _send(self, status, body, content_type="text/plain; charset=utf-8"):
        data = body.encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def do_GET(self):
        if self.path == "/":
            self._send(
                200,
                "Welcome to the tiny intranet.\nCheck the crawler rules before knocking.\n",
            )
            return
        if self.path == "/robots.txt":
            self._send(200, "User-agent: *\nDisallow: /staff-door\n")
            return
        if self.path == "/staff-door":
            token = self.headers.get("X-CTF-Token", "")
            if token != "open-sesame":
                self._send(
                    403,
                    "Access denied.\nRequired header: X-CTF-Token: open-sesame\n",
                )
                return
            self._send(200, f"Door opened.\n{FLAG}\n")
            return
        self._send(404, "Not found.\n")

    def log_message(self, fmt, *args):
        return


if __name__ == "__main__":
    ThreadingHTTPServer(("0.0.0.0", 8080), Handler).serve_forever()
