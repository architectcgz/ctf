import cgi
import os
import subprocess
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import unquote

FLAG = "flag{web_file_upload_double_ext}"
BASE = os.path.dirname(__file__)
UPLOADS = os.path.join(BASE, "uploads")
os.makedirs(UPLOADS, exist_ok=True)
open(os.path.join(BASE, "flag.txt"), "w", encoding="utf-8").write(FLAG + "\n")

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/":
            body = "<form method='post' action='/upload' enctype='multipart/form-data'><input type='file' name='file'><button>upload</button></form>"
            self._send(200, body, "text/html; charset=utf-8")
            return
        if self.path.startswith("/files/"):
            name = unquote(self.path[len("/files/"):])
            path = os.path.join(UPLOADS, name)
            if not os.path.exists(path):
                self._send(404, "missing")
                return
            if name.endswith(".py"):
                out = subprocess.getoutput("python3 " + path)
                self._send(200, out)
            else:
                self._send(200, open(path, "r", encoding="utf-8", errors="ignore").read())
            return
        self._send(404, "not found")

    def do_POST(self):
        if self.path != "/upload":
            self._send(404, "not found")
            return
        form = cgi.FieldStorage(
            fp=self.rfile,
            headers=self.headers,
            environ={
                "REQUEST_METHOD": "POST",
                "CONTENT_TYPE": self.headers["Content-Type"],
                "CONTENT_LENGTH": self.headers.get("Content-Length", "0"),
            },
        )
        item = form["file"]
        name = os.path.basename(item.filename)
        first_ext = name.split(".")[1] if "." in name else ""
        if first_ext not in {"jpg", "png"}:
            self._send(400, "bad ext")
            return
        path = os.path.join(UPLOADS, name)
        with open(path, "wb") as f:
            f.write(item.file.read())
        self._send(200, f"stored /files/{name}")

    def _send(self, status, body, ctype="text/plain; charset=utf-8"):
        data = body.encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def log_message(self, *args):
        return

if __name__ == "__main__":
    port = int(os.getenv("PORT", "8080"))
    ThreadingHTTPServer(("127.0.0.1", port), Handler).serve_forever()
