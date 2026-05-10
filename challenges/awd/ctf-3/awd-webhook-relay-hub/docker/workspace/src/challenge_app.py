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
ARCHIVE_PATH = DATA_DIR / "relay_archives.json"
ROLE = os.environ.get("ROLE", "relay-web")
RELAY_WORKER_URL = os.environ.get("RELAY_WORKER_URL", "http://relay-worker:9091")
RELAY_STORE_URL = os.environ.get("RELAY_STORE_URL", "http://relay-store:9092")
DEFAULT_ARCHIVES = {
    "demo": {
        "target": "https://hooks.academy.example/demo",
        "last_event": "user.sync",
        "note": "demo feed for teacher preview",
    }
}


def setup_business_data():
    DATA_DIR.mkdir(parents=True, exist_ok=True)
    if not ARCHIVE_PATH.exists():
        ARCHIVE_PATH.write_text(
            json.dumps(DEFAULT_ARCHIVES, ensure_ascii=False, indent=2),
            encoding="utf-8",
        )


def current_flag() -> str:
    return FLAG_PATH.read_text(encoding="utf-8").strip()


def load_archives() -> dict:
    setup_business_data()
    return json.loads(ARCHIVE_PATH.read_text(encoding="utf-8"))


def page(body: str) -> str:
    return (
        "<!doctype html><title>Webhook Relay Hub</title>"
        "<style>body{font-family:sans-serif;max-width:920px;margin:32px auto;line-height:1.6}"
        "input{margin:8px 0;padding:8px 10px;width:100%;box-sizing:border-box}"
        "button{padding:8px 14px}pre{background:#f5f5f5;padding:14px;border-radius:10px;overflow:auto}"
        "code{background:#f3f4f6;padding:2px 6px;border-radius:6px}</style>"
        "<h1>Webhook Relay Hub</h1>"
        "<p><a href='/'>首页</a> <a href='/feeds/demo'>示例 Feed</a></p>"
        + body
    )


def fetch_text(target_url: str) -> str:
    try:
        with urlopen(target_url, timeout=3) as resp:
            return resp.read(4096).decode("utf-8", errors="replace")
    except URLError as exc:
        abort(502, f"preview failed: {exc.reason}")


def register_relay_web_routes(app):
    @app.get("/")
    def index():
        return page(
            "<p>Webhook Relay Hub 会代运维同学抓取 feed 文档并预览回放结果。</p>"
            "<form action='/preview' method='get'>"
            "<input name='url' value='https://example.com/.well-known/feed.json'>"
            "<button type='submit'>预览</button>"
            "</form>"
            "<p>开发限制：已阻止 <code>localhost</code> / <code>127.0.0.1</code> / <code>0.0.0.0</code>。</p>"
        )

    @app.get("/feeds/demo")
    def demo_feed():
        payload = {
            "name": "relay-demo",
            "version": "2026.05",
            "replay_endpoint": "http://relay-worker:9091/internal/replay?feed=demo",
            "fields": ["event", "signature", "payload"],
        }
        return Response(json.dumps(payload, ensure_ascii=False), mimetype="application/json")

    @app.get("/preview")
    def preview():
        target_url = request.args.get("url", "").strip()
        if not target_url:
            abort(400, "missing url parameter")
        lowered = target_url.lower()
        if not lowered.startswith(("http://", "https://")):
            abort(400, "http or https only")
        if "127.0.0.1" in lowered or "localhost" in lowered or "0.0.0.0" in lowered:
            abort(403, "target blocked")
        data = fetch_text(target_url)
        return page("<h2>Preview</h2><pre>" + html.escape(data) + "</pre>")


def register_relay_worker_routes(app):
    @app.get("/internal/replay")
    def internal_replay():
        feed = request.args.get("feed", "demo").strip() or "demo"
        store_url = RELAY_STORE_URL + "/internal/archive?feed=" + quote(feed, safe="")
        data = fetch_text(store_url)
        return Response(data, mimetype="application/json")


def register_relay_store_routes(app):
    @app.get("/internal/archive")
    def internal_archive():
        feed = request.args.get("feed", "demo").strip() or "demo"
        archives = load_archives()
        record = archives.get(feed)
        if record is None:
            abort(404, "unknown feed")
        payload = {
            "feed": feed,
            "delivery": record,
            "debug_scope": "relay-store",
            "flag": current_flag(),
        }
        return Response(json.dumps(payload, ensure_ascii=False), mimetype="application/json")


def register_challenge_routes(app):
    @app.before_request
    def before_request():
        setup_business_data()

    if ROLE == "relay-web":
        register_relay_web_routes(app)
        return
    if ROLE == "relay-worker":
        register_relay_worker_routes(app)
        return
    if ROLE == "relay-store":
        register_relay_store_routes(app)
        return

    @app.get("/")
    def unknown_role():
        abort(500, f"unknown role: {ROLE}")
