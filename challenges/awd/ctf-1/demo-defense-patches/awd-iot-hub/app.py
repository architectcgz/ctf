import os
import secrets
import sqlite3
from pathlib import Path

from flask import Flask, Response, g, jsonify, request

APP = Flask(__name__)
DB_PATH = Path("/data/iot.db")
FIRMWARE_DIR = Path("/data/firmware")
FLAG_PATH = Path("/flag")
ADMIN_TOKEN = os.environ.get("ADMIN_TOKEN") or secrets.token_hex(16)


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


def setup():
    FIRMWARE_DIR.mkdir(parents=True, exist_ok=True)
    if not FLAG_PATH.exists():
        FLAG_PATH.write_text(os.environ.get("FLAG", "flag{local_demo}"), encoding="utf-8")
    conn = db()
    conn.execute(
        "create table if not exists telemetry "
        "(id integer primary key autoincrement, team_id text, device_id text, topic text, payload text)"
    )
    conn.execute(
        "create table if not exists commands "
        "(id integer primary key autoincrement, team_id text, device_id text, topic text, command text)"
    )
    conn.execute(
        "create table if not exists devices "
        "(device_id text primary key, team_id text, device_key text)"
    )
    conn.execute(
        "create table if not exists firmware_files "
        "(name text primary key, file_name text)"
    )
    if conn.execute("select count(*) from devices").fetchone()[0] == 0:
        conn.execute(
            "insert into devices(device_id, team_id, device_key) values (?, ?, ?)",
            ("sensor-a", "team1", os.environ.get("DEVICE_KEY_SENSOR_A", secrets.token_hex(12))),
        )
    fw = FIRMWARE_DIR / "sensor-v1.bin"
    if not fw.exists():
        fw.write_text("demo firmware", encoding="utf-8")
    conn.execute(
        "insert or ignore into firmware_files(name, file_name) values (?, ?)",
        ("sensor-v1", "sensor-v1.bin"),
    )
    conn.commit()


@APP.before_request
def before_request():
    setup()


def require_device():
    device_id = request.headers.get("X-Device-Id", "").strip()
    device_key = request.headers.get("X-Device-Key", "").strip()
    if not device_id or not device_key:
        return None
    row = db().execute(
        "select device_id, team_id, device_key from devices where device_id=?",
        (device_id,),
    ).fetchone()
    if row is None or row["device_key"] != device_key:
        return None
    return row


def split_topic(raw_topic):
    parts = [part for part in raw_topic.strip("/").split("/") if part]
    if len(parts) < 3:
        return None
    return parts[0], parts[1], "/".join(parts[2:])


@APP.get("/")
def index():
    rows = db().execute(
        "select topic, payload from telemetry order by id desc limit 20"
    ).fetchall()
    items = "".join(f"<li><code>{r['topic']}</code>: {r['payload']}</li>" for r in rows)
    return (
        "<!doctype html><title>IoT 设备管理平台</title>"
        "<h1>IoT 设备管理平台</h1>"
        "<ul>" + items + "</ul>"
    )


@APP.post("/api/telemetry")
def telemetry():
    device = require_device()
    if device is None:
        return jsonify({"error": "bad device key"}), 403
    topic = request.json.get("topic", "")
    payload = request.json.get("payload", "")
    parsed = split_topic(topic)
    if parsed is None:
        return jsonify({"error": "invalid topic"}), 400
    team_id, device_id, _rest = parsed
    if team_id != device["team_id"] or device_id != device["device_id"]:
        return jsonify({"error": "topic denied"}), 403
    db().execute(
        "insert into telemetry(team_id, device_id, topic, payload) values (?, ?, ?, ?)",
        (team_id, device_id, topic, payload),
    )
    db().commit()
    return jsonify({"status": "stored"})


@APP.get("/api/topic")
def topic_read():
    requested = request.args.get("topic", "")
    parsed = split_topic(requested)
    if parsed is None:
        return jsonify({"error": "invalid topic"}), 400
    team_id, device_id, _rest = parsed
    rows = db().execute(
        "select topic, payload from telemetry where team_id=? and device_id=? and topic like ? order by id desc",
        (team_id, device_id, requested.rstrip("/") + "%"),
    ).fetchall()
    return jsonify([dict(row) for row in rows])


@APP.post("/api/command")
def command():
    if request.headers.get("X-Admin-Token") != ADMIN_TOKEN:
        return jsonify({"error": "forbidden"}), 403
    topic = request.json.get("topic", "")
    command_text = request.json.get("command", "")
    parsed = split_topic(topic)
    if parsed is None:
        return jsonify({"error": "invalid topic"}), 400
    team_id, device_id, _rest = parsed
    db().execute(
        "insert into commands(team_id, device_id, topic, command) values (?, ?, ?, ?)",
        (team_id, device_id, topic, command_text),
    )
    db().commit()
    return jsonify({"status": "queued"})


@APP.get("/firmware")
def firmware():
    firmware_name = request.args.get("name", "sensor-v1")
    row = db().execute(
        "select file_name from firmware_files where name=?",
        (firmware_name,),
    ).fetchone()
    if row is None:
        return jsonify({"error": "not found"}), 404
    target = FIRMWARE_DIR / row["file_name"]
    if not target.is_file():
        return jsonify({"error": "not found"}), 404
    return target.read_text(encoding="utf-8", errors="replace")


@APP.get("/health")
def health():
    return jsonify({"status": "ok"})


@APP.route("/api/flag", methods=["GET", "PUT"])
def checker_flag():
    expected_token = os.environ.get("CHECKER_TOKEN", "demo-checker-token")
    if request.headers.get("X-AWD-Checker-Token") != expected_token:
        return jsonify({"error": "not found"}), 404
    if request.method == "PUT":
        candidate = request.get_data(as_text=True).strip()
        if not candidate:
            return jsonify({"error": "empty flag"}), 400
        FLAG_PATH.write_text(candidate, encoding="utf-8")
        return jsonify({"status": "accepted"})
    return Response(FLAG_PATH.read_text(encoding="utf-8").strip(), mimetype="text/plain")


if __name__ == "__main__":
    APP.run(host="0.0.0.0", port=8080)
