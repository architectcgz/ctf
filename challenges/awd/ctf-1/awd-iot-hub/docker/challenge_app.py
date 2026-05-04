import os
import sqlite3
from pathlib import Path

from flask import g, jsonify, request

DB_PATH = Path("/data/iot.db")
FIRMWARE_DIR = Path("/data/firmware")


def db():
    if "db" not in g:
        DB_PATH.parent.mkdir(parents=True, exist_ok=True)
        g.db = sqlite3.connect(DB_PATH)
        g.db.row_factory = sqlite3.Row
    return g.db


def close_db(_error):
    conn = g.pop("db", None)
    if conn is not None:
        conn.close()


def setup_business_data():
    FIRMWARE_DIR.mkdir(parents=True, exist_ok=True)
    conn = db()
    conn.execute(
        "create table if not exists telemetry "
        "(id integer primary key autoincrement, topic text, payload text)"
    )
    conn.execute(
        "create table if not exists commands "
        "(id integer primary key autoincrement, topic text, command text)"
    )
    conn.commit()
    fw = FIRMWARE_DIR / "sensor-v1.bin"
    if not fw.exists():
        fw.write_text("demo firmware", encoding="utf-8")


def valid_device_key():
    return request.headers.get("X-Device-Key") == os.environ.get("DEVICE_KEY", "device-demo-key")


def register_challenge_routes(app):
    app.teardown_appcontext(close_db)

    @app.before_request
    def before_request():
        setup_business_data()

    @app.get("/")
    def index():
        rows = db().execute("select topic, payload from telemetry order by id desc limit 20").fetchall()
        items = "".join(f"<li><code>{r['topic']}</code>: {r['payload']}</li>" for r in rows)
        return (
            "<!doctype html><title>IoT 设备管理平台</title>"
            "<h1>IoT 设备管理平台</h1>"
            "<ul>" + items + "</ul>"
        )

    @app.post("/api/telemetry")
    def telemetry():
        if not valid_device_key():
            return jsonify({"error": "bad device key"}), 403
        topic = request.json.get("topic", "")
        payload = request.json.get("payload", "")
        db().execute("insert into telemetry(topic, payload) values (?, ?)", (topic, payload))
        db().commit()
        return jsonify({"status": "stored"})

    @app.get("/api/topic")
    def topic_read():
        requested = request.args.get("topic", "")
        allowed_prefix = request.args.get("allow", "team1/")
        if not requested.startswith(allowed_prefix):
            return jsonify({"error": "topic denied"}), 403
        rows = db().execute(
            "select topic, payload from telemetry where topic like ? order by id desc",
            (requested + "%",),
        ).fetchall()
        return jsonify([dict(row) for row in rows])

    @app.post("/api/command")
    def command():
        topic = request.json.get("topic", "")
        command_text = request.json.get("command", "")
        db().execute("insert into commands(topic, command) values (?, ?)", (topic, command_text))
        db().commit()
        return jsonify({"status": "queued"})

    @app.get("/firmware")
    def firmware():
        name = request.args.get("name", "sensor-v1.bin")
        target = FIRMWARE_DIR / name
        if not target.exists():
            return jsonify({"error": "not found"}), 404
        return target.read_text(encoding="utf-8", errors="replace")
