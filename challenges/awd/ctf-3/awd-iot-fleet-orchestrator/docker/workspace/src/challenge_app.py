import html
import json
import os
from pathlib import Path
from urllib.error import URLError
from urllib.parse import quote
from urllib.request import urlopen

from flask import Response, abort, request

FLAG_PATH = Path(os.environ.get("FLAG_PATH", "/shared/flag"))
DATA_DIR = Path("/workspace/data")
DEVICE_PATH = DATA_DIR / "fleet_devices.json"
ROLE = os.environ.get("ROLE", "fleet-web")
AGENT_BASE_URL = os.environ.get("AGENT_BASE_URL", "http://fleet-agent:9091")
VAULT_BASE_URL = os.environ.get("VAULT_BASE_URL", "http://config-vault:9092")
DEFAULT_FLEET_KEY = os.environ.get("DEFAULT_FLEET_KEY", "fleet-support-2026")
DEFAULT_DEVICES = {
    "demo-sensor": {
        "site": "lab-a",
        "status": "online",
        "device_key": "sensor-demo-local",
    }
}


def setup_business_data():
    DATA_DIR.mkdir(parents=True, exist_ok=True)
    if not DEVICE_PATH.exists():
        DEVICE_PATH.write_text(
            json.dumps(DEFAULT_DEVICES, ensure_ascii=False, indent=2),
            encoding="utf-8",
        )


def current_flag() -> str:
    return FLAG_PATH.read_text(encoding="utf-8").strip()


def load_devices() -> dict:
    setup_business_data()
    return json.loads(DEVICE_PATH.read_text(encoding="utf-8"))


def page(body: str) -> str:
    return (
        "<!doctype html><title>IoT Fleet Orchestrator</title>"
        "<style>body{font-family:sans-serif;max-width:920px;margin:32px auto;line-height:1.6}"
        "input{margin:8px 0;padding:8px 10px;width:100%;box-sizing:border-box}"
        "button{padding:8px 14px}pre{background:#f5f5f5;padding:14px;border-radius:10px;overflow:auto}"
        "code{background:#f3f4f6;padding:2px 6px;border-radius:6px}</style>"
        "<h1>IoT Fleet Orchestrator</h1>"
        "<p><a href='/'>首页</a> <a href='/devices/demo'>示例设备</a></p>"
        + body
    )


def fetch_text(target_url: str) -> str:
    try:
        with urlopen(target_url, timeout=3) as resp:
            return resp.read(4096).decode("utf-8", errors="replace")
    except URLError as exc:
        abort(502, f"dispatch failed: {exc.reason}")


def register_fleet_web_routes(app):
    @app.get("/")
    def index():
        return page(
            "<p>IoT Fleet Orchestrator 会把公网控制台的操作转发给内部 fleet-agent。</p>"
            "<form action='/dispatch' method='get'>"
            "<input name='device' value='demo-sensor'>"
            "<input name='key' value='sensor-demo-local'>"
            "<input name='op' value='status'>"
            "<button type='submit'>发送</button>"
            "</form>"
        )

    @app.get("/devices/demo")
    def devices_demo():
        payload = {
            "device": "demo-sensor",
            "sample_status": "/dispatch?device=demo-sensor&key=sensor-demo-local&op=status",
            "ops": ["status", "ping", "pull-config"],
        }
        return Response(json.dumps(payload, ensure_ascii=False), mimetype="application/json")

    @app.get("/dispatch")
    def dispatch():
        device = request.args.get("device", "").strip()
        key = request.args.get("key", "").strip()
        op = request.args.get("op", "").strip()
        if not device or not key or not op:
            abort(400, "device, key and op are required")
        target = (
            AGENT_BASE_URL
            + "/internal/dispatch?device="
            + quote(device, safe="")
            + "&key="
            + quote(key, safe="")
            + "&op="
            + quote(op, safe="")
        )
        data = fetch_text(target)
        return page("<h2>Dispatch Result</h2><pre>" + html.escape(data) + "</pre>")


def register_fleet_agent_routes(app):
    @app.get("/internal/dispatch")
    def internal_dispatch():
        device = request.args.get("device", "").strip()
        key = request.args.get("key", "").strip()
        op = request.args.get("op", "").strip()
        devices = load_devices()
        record = devices.get(device)
        if record is None:
            abort(404, "unknown device")
        if key not in {record["device_key"], DEFAULT_FLEET_KEY}:
            abort(403, "invalid key")
        if op == "status":
            return {"device": device, "status": record["status"], "site": record["site"]}
        if op == "ping":
            return {"device": device, "result": "pong"}
        if op == "pull-config":
            data = fetch_text(VAULT_BASE_URL + "/internal/device?device=" + quote(device, safe=""))
            return Response(data, mimetype="application/json")
        abort(400, "unsupported op")


def register_config_vault_routes(app):
    @app.get("/internal/device")
    def internal_device():
        device = request.args.get("device", "").strip()
        devices = load_devices()
        record = devices.get(device)
        if record is None:
            abort(404, "unknown device")
        payload = {
            "device": device,
            "site": record["site"],
            "status": record["status"],
            "flag": current_flag(),
        }
        return Response(json.dumps(payload, ensure_ascii=False), mimetype="application/json")


def register_challenge_routes(app):
    @app.before_request
    def before_request():
        setup_business_data()

    if ROLE == "fleet-web":
        register_fleet_web_routes(app)
        return
    if ROLE == "fleet-agent":
        register_fleet_agent_routes(app)
        return
    if ROLE == "config-vault":
        register_config_vault_routes(app)
        return

    @app.get("/")
    def unknown_role():
        abort(500, f"unknown role: {ROLE}")
