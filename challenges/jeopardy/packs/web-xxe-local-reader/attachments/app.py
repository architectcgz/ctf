import os, re
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

FLAG = "flag{web_xxe_local_reader}"
BASE = os.path.dirname(__file__)
open(os.path.join(BASE, "flag.txt"), "w", encoding="utf-8").write(FLAG + "\n")

class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path != "/import":
            self._send(404, "not found")
            return
        length = int(self.headers.get("Content-Length", "0"))
        xml = self.rfile.read(length).decode("utf-8")
        entity = re.search(r'<!ENTITY\s+(\w+)\s+SYSTEM\s+"file://([^"]+)">', xml)
        if entity:
            name, path = entity.groups()
            if os.path.exists(path):
                xml = xml.replace(f"&{name};", open(path, "r", encoding="utf-8").read())
        content = re.search(r"<data>(.*?)</data>", xml, re.S)
        self._send(200, content.group(1) if content else "empty")

    def do_GET(self):
        if self.path == "/":
            self._send(200, "post xml to /import")
        else:
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
