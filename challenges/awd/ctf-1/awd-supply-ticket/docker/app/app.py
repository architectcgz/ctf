import os
import sqlite3
from pathlib import Path

from flask import Flask, Response, g, redirect, render_template_string, request, session, url_for

APP = Flask(__name__)
APP.secret_key = os.environ.get("SECRET_KEY", "dev-secret-change-me")
DB_PATH = Path("/data/ticket.db")
FLAG_PATH = Path("/flag")


def db():
    if "db" not in g:
        DB_PATH.parent.mkdir(parents=True, exist_ok=True)
        g.db = sqlite3.connect(DB_PATH)
        g.db.row_factory = sqlite3.Row
    return g.db


@APP.teardown_appcontext
def close_db(_error):
    conn = g.pop("db", None)
    if conn is not None:
        conn.close()


def init_db():
    if not FLAG_PATH.exists():
        FLAG_PATH.write_text(os.environ.get("FLAG", "flag{local_demo}"), encoding="utf-8")
    conn = db()
    conn.execute(
        "create table if not exists tickets "
        "(id integer primary key autoincrement, title text, content text, status text)"
    )
    conn.execute(
        "create table if not exists users "
        "(username text primary key, password text, role text)"
    )
    admin_password = os.environ.get("ADMIN_PASSWORD", "admin123")
    conn.execute(
        "insert or ignore into users values (?, ?, ?)",
        ("admin", admin_password, "admin"),
    )
    conn.commit()


LAYOUT = """
<!doctype html>
<title>供应链工单系统</title>
<style>
body{font-family:sans-serif;max-width:960px;margin:32px auto;line-height:1.5}
input,textarea{display:block;width:100%;margin:8px 0;padding:8px}
a,button{margin-right:12px}
.ticket{border:1px solid #ddd;padding:12px;margin:12px 0}
</style>
<h1>供应链工单系统</h1>
<p><a href="/">首页</a><a href="/new">创建工单</a><a href="/admin">后台</a><a href="/logout">退出</a></p>
{{ body|safe }}
"""


def page(body):
    return render_template_string(LAYOUT, body=body)


@APP.before_request
def before_request():
    init_db()


@APP.get("/")
def index():
    rows = db().execute("select * from tickets order by id desc").fetchall()
    items = "".join(
        f"<div class='ticket'><b>#{r['id']} {r['title']}</b><p>{r['content']}</p><em>{r['status']}</em></div>"
        for r in rows
    )
    return page(items or "<p>暂无工单。</p>")


@APP.route("/login", methods=["GET", "POST"])
def login():
    if request.method == "POST":
        user = db().execute(
            "select * from users where username=? and password=?",
            (request.form["username"], request.form["password"]),
        ).fetchone()
        if user:
            session["user"] = user["username"]
            session["role"] = user["role"]
            return redirect(url_for("index"))
        return page("<p>登录失败。</p>")
    return page(
        "<form method='post'>"
        "<input name='username' placeholder='username'>"
        "<input name='password' placeholder='password' type='password'>"
        "<button>登录</button></form>"
    )


@APP.get("/logout")
def logout():
    session.clear()
    return redirect(url_for("index"))


@APP.route("/new", methods=["GET", "POST"])
def new_ticket():
    if request.method == "POST":
        db().execute(
            "insert into tickets(title, content, status) values (?, ?, 'open')",
            (request.form["title"], request.form["content"]),
        )
        db().commit()
        return redirect(url_for("index"))
    return page(
        "<form method='post'>"
        "<input name='title' placeholder='工单标题'>"
        "<textarea name='content' placeholder='问题描述'></textarea>"
        "<button>提交</button></form>"
    )


@APP.get("/notify/<int:ticket_id>")
def notify(ticket_id):
    row = db().execute("select * from tickets where id=?", (ticket_id,)).fetchone()
    if row is None:
        return page("<p>工单不存在。</p>"), 404
    template = f"通知：工单 #{row['id']} - {row['title']}，状态 {row['status']}"
    return page(render_template_string(template))


@APP.get("/admin")
def admin():
    if session.get("role") != "admin":
        return redirect(url_for("login"))
    flag = FLAG_PATH.read_text(encoding="utf-8").strip()
    return page(f"<p>管理员面板</p><p>本轮 Flag：<code>{flag}</code></p>")


@APP.get("/health")
def health():
    return {"status": "ok"}


@APP.route("/api/flag", methods=["GET", "PUT"])
def checker_flag():
    expected_token = os.environ.get("CHECKER_TOKEN", "demo-checker-token")
    if request.headers.get("X-AWD-Checker-Token") != expected_token:
        return {"error": "not found"}, 404
    if request.method == "PUT":
        candidate = request.get_data(as_text=True).strip()
        if not candidate:
            return {"error": "empty flag"}, 400
        FLAG_PATH.write_text(candidate, encoding="utf-8")
        return {"status": "accepted"}
    return Response(FLAG_PATH.read_text(encoding="utf-8").strip(), mimetype="text/plain")


if __name__ == "__main__":
    APP.run(host="0.0.0.0", port=8080)
