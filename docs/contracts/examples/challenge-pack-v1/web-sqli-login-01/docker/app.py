import os
import secrets
import sqlite3
from pathlib import Path

from flask import Flask, redirect, render_template_string, request, session, url_for

APP_DIR = Path(__file__).resolve().parent
DATA_DIR = APP_DIR / "data"
DB_PATH = DATA_DIR / "app.db"

app = Flask(__name__)
app.secret_key = os.getenv("APP_SECRET", "") or secrets.token_hex(32)


def get_db() -> sqlite3.Connection:
    DATA_DIR.mkdir(parents=True, exist_ok=True)
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    return conn


def init_db() -> None:
    flag_value = os.getenv("CTF_FLAG", "")
    with get_db() as conn:
        conn.execute(
            "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, password TEXT)"
        )
        conn.execute(
            "CREATE TABLE IF NOT EXISTS secrets (id INTEGER PRIMARY KEY AUTOINCREMENT, secret TEXT)"
        )
        conn.execute("INSERT INTO users (username, password) VALUES ('guest', 'guest')")
        if flag_value:
            conn.execute("INSERT INTO secrets (secret) VALUES (?)", (flag_value,))
        else:
            conn.execute("INSERT INTO secrets (secret) VALUES ('flag{CTF_FLAG_not_set}')")


@app.before_request
def ensure_db():
    if not DB_PATH.exists():
        init_db()


INDEX_TPL = """
<!doctype html>
<meta charset="utf-8" />
<title>Web-02</title>
<h1>Web-02 SQL Injection: Login Bypass</h1>
<p><a href="{{ url_for('login') }}">Login</a></p>
<p><a href="{{ url_for('me') }}">Me</a></p>
"""


LOGIN_TPL = """
<!doctype html>
<meta charset="utf-8" />
<title>Login</title>
<h1>Login</h1>
{% if error %}<pre style="color:#b00">{{ error }}</pre>{% endif %}
<form method="post">
  <label>Username <input name="username" autocomplete="off"></label><br>
  <label>Password <input name="password" type="password" autocomplete="off"></label><br>
  <button type="submit">Sign in</button>
</form>
<p>Hint: try guest/guest</p>
"""


ME_TPL = """
<!doctype html>
<meta charset="utf-8" />
<title>Me</title>
<h1>Me</h1>
{% if user %}
  <pre>user={{ user }}</pre>
  <p><a href="{{ url_for('logout') }}">Logout</a></p>
{% else %}
  <p>Not logged in. <a href="{{ url_for('login') }}">Login</a></p>
{% endif %}
"""


@app.get("/")
def index():
    return render_template_string(INDEX_TPL)


@app.route("/login", methods=["GET", "POST"])
def login():
    if request.method == "GET":
        return render_template_string(LOGIN_TPL, error=None)

    username = request.form.get("username", "")
    password = request.form.get("password", "")

    # Intentionally vulnerable SQLi for CTF purposes.
    sql = (
        "SELECT username FROM users "
        f"WHERE username = '{username}' AND password = '{password}' "
        "LIMIT 1"
    )

    try:
        with get_db() as conn:
            row = conn.execute(sql).fetchone()
    except Exception as exc:
        return render_template_string(LOGIN_TPL, error=f"SQL error: {exc}"), 400

    if not row:
        return render_template_string(LOGIN_TPL, error="Invalid credentials"), 401

    session["user"] = str(row["username"])
    return redirect(url_for("me"))


@app.get("/me")
def me():
    return render_template_string(ME_TPL, user=session.get("user"))


@app.get("/logout")
def logout():
    session.clear()
    return redirect(url_for("index"))

