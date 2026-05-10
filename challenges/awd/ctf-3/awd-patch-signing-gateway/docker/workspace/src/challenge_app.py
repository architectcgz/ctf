import html
import json
import os
from pathlib import Path
from urllib.error import URLError
from urllib.parse import quote
from urllib.request import Request, urlopen

from flask import Response, abort, request

FLAG_PATH = Path(os.environ.get("FLAG_PATH", "/shared/flag"))
DATA_DIR = Path("/workspace/data")
CHANNEL_PATH = DATA_DIR / "patch_channels.json"
ROLE = os.environ.get("ROLE", "patch-gateway")
SIGNER_BASE_URL = os.environ.get("SIGNER_BASE_URL", "http://signer-core:9091")
KEY_BASE_URL = os.environ.get("KEY_BASE_URL", "http://key-vault:9092")
DEFAULT_CHANNELS = {
    "stable": {"build": "2026.05.1", "notes": "lts release"},
    "beta": {"build": "2026.05.2-beta", "notes": "preview release"},
}


def setup_business_data():
    DATA_DIR.mkdir(parents=True, exist_ok=True)
    if not CHANNEL_PATH.exists():
        CHANNEL_PATH.write_text(
            json.dumps(DEFAULT_CHANNELS, ensure_ascii=False, indent=2),
            encoding="utf-8",
        )


def current_flag() -> str:
    return FLAG_PATH.read_text(encoding="utf-8").strip()


def load_channels() -> dict:
    setup_business_data()
    return json.loads(CHANNEL_PATH.read_text(encoding="utf-8"))


def page(body: str) -> str:
    return (
        "<!doctype html><title>Patch Signing Gateway</title>"
        "<style>body{font-family:sans-serif;max-width:920px;margin:32px auto;line-height:1.6}"
        "input{margin:8px 0;padding:8px 10px;width:100%;box-sizing:border-box}"
        "button{padding:8px 14px}pre{background:#f5f5f5;padding:14px;border-radius:10px;overflow:auto}"
        "code{background:#f3f4f6;padding:2px 6px;border-radius:6px}</style>"
        "<h1>Patch Signing Gateway</h1>"
        "<p><a href='/'>首页</a> <a href='/channels/demo'>示例渠道</a></p>"
        + body
    )


def fetch_text(target_url: str, headers: dict | None = None) -> str:
    req = Request(target_url, headers=headers or {})
    try:
        with urlopen(req, timeout=3) as resp:
            return resp.read(4096).decode("utf-8", errors="replace")
    except URLError as exc:
        abort(502, f"bundle failed: {exc.reason}")


def register_gateway_routes(app):
    @app.get("/")
    def index():
        return page(
            "<p>Patch Signing Gateway 会代公网客户端请求内部 signer，返回补丁 bundle 预览。</p>"
            "<form action='/bundle' method='get'>"
            "<input name='channel' value='stable'>"
            "<input name='signer_role' value='viewer'>"
            "<button type='submit'>生成</button>"
            "</form>"
        )

    @app.get("/channels/demo")
    def channels_demo():
        payload = {
            "channel": "stable",
            "preview_path": "/bundle?channel=stable",
            "roles": ["viewer", "release-manager"],
        }
        return Response(json.dumps(payload, ensure_ascii=False), mimetype="application/json")

    @app.get("/bundle")
    def bundle():
        channel = request.args.get("channel", "").strip()
        signer_role = request.args.get("signer_role", "viewer").strip() or "viewer"
        if not channel:
            abort(400, "missing channel parameter")
        target = SIGNER_BASE_URL + "/internal/bundle?channel=" + quote(channel, safe="")
        data = fetch_text(target, headers={"X-Signer-Role": signer_role})
        return page("<h2>Bundle</h2><pre>" + html.escape(data) + "</pre>")


def register_signer_routes(app):
    @app.get("/internal/bundle")
    def internal_bundle():
        channel = request.args.get("channel", "stable").strip() or "stable"
        signer_role = request.headers.get("X-Signer-Role", "viewer")
        if signer_role == "release-manager":
            data = fetch_text(KEY_BASE_URL + "/internal/keyset?channel=" + quote(channel, safe=""))
            return Response(data, mimetype="application/json")
        data = fetch_text(KEY_BASE_URL + "/internal/public-manifest?channel=" + quote(channel, safe=""))
        return Response(data, mimetype="application/json")


def register_key_vault_routes(app):
    @app.get("/internal/public-manifest")
    def public_manifest():
        channel = request.args.get("channel", "stable").strip() or "stable"
        channels = load_channels()
        record = channels.get(channel)
        if record is None:
            abort(404, "unknown channel")
        payload = {"channel": channel, "build": record["build"], "notes": record["notes"]}
        return Response(json.dumps(payload, ensure_ascii=False), mimetype="application/json")

    @app.get("/internal/keyset")
    def keyset():
        channel = request.args.get("channel", "stable").strip() or "stable"
        channels = load_channels()
        record = channels.get(channel)
        if record is None:
            abort(404, "unknown channel")
        payload = {
            "channel": channel,
            "build": record["build"],
            "role": "release-manager",
            "flag": current_flag(),
        }
        return Response(json.dumps(payload, ensure_ascii=False), mimetype="application/json")


def register_challenge_routes(app):
    @app.before_request
    def before_request():
        setup_business_data()

    if ROLE == "patch-gateway":
        register_gateway_routes(app)
        return
    if ROLE == "signer-core":
        register_signer_routes(app)
        return
    if ROLE == "key-vault":
        register_key_vault_routes(app)
        return

    @app.get("/")
    def unknown_role():
        abort(500, f"unknown role: {ROLE}")
