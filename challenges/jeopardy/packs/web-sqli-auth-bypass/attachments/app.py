import sqlite3
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs
import os

FLAG = "flag{web_sqli_auth_bypass}"
DB = sqlite3.connect(":memory:", check_same_thread=False)
DB.execute("create table users(username text, password text, role text)")
DB.executemany("insert into users values(?,?,?)", [("admin","S3aled!","admin"),("guest","guest","user")])
DB.commit()

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/":
            body = "<form method='post' action='/login'><input name='username'><input name='password'><button>login</button></form>"
            self._send(200, body, "text/html; charset=utf-8")
        else:
            self._send(404, "not found")

    def do_POST(self):
        if self.path != "/login":
            self._send(404, "not found")
            return
        length = int(self.headers.get("Content-Length", "0"))
        form = parse_qs(self.rfile.read(length).decode("utf-8"))
        username = form.get("username", [""])[0]
        password = form.get("password", [""])[0]
        query = f"select role from users where username = '{username}' and password = '{password}'"
        row = DB.execute(query).fetchone()
        if row and row[0] == "admin":
            self._send(200, FLAG)
        else:
            self._send(403, "denied")

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
