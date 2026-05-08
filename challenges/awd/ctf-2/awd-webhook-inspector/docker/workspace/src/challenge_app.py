import json
from pathlib import Path
from urllib.error import URLError
from urllib.request import urlopen

from flask import Response, abort, request

FLAG_PATH = Path("/flag")
DATA_DIR = Path("/workspace/data")


def setup_business_data():
    DATA_DIR.mkdir(parents=True, exist_ok=True)


def page(body):
    return (
        "<!doctype html><title>Webhook Inspector</title>"
        "<style>body{font-family:sans-serif;max-width:900px;margin:32px auto;line-height:1.6}"
        "input{margin:8px 0;padding:8px 10px;width:100%;box-sizing:border-box}"
        "button{padding:8px 14px}pre{background:#f5f5f5;padding:14px;border-radius:10px;overflow:auto}</style>"
        "<h1>Webhook Inspector</h1>"
        "<p><a href='/'>首页</a> <a href='/manifest/demo'>示例 Manifest</a></p>"
        + body
    )


def current_flag() -> str:
    return FLAG_PATH.read_text(encoding="utf-8").strip()


def register_challenge_routes(app):
    @app.before_request
    def before_request():
        setup_business_data()

    @app.get("/")
    def index():
        return page(
            "<p>输入 webhook / manifest 地址，服务端会代你抓取并预览内容。</p>"
            "<form action='/preview' method='get'>"
            "<input name='url' value='https://example.com/.well-known/manifest.json'>"
            "<button type='submit'>预览</button>"
            "</form>"
            "<p>开发限制：已阻止 localhost / 127.0.0.1 / 0.0.0.0。</p>"
        )

    @app.get("/manifest/demo")
    def demo_manifest():
        payload = {
            "name": "webhook-inspector-demo",
            "version": "2026.05",
            "fields": ["event", "signature", "payload"],
        }
        return Response(json.dumps(payload, ensure_ascii=False), mimetype="application/json")

    @app.get("/internal/snapshot")
    def internal_snapshot():
        if request.remote_addr not in {"127.0.0.1", "::1"}:
            abort(403, "local access only")
        return {"flag": current_flag(), "scope": "inspector-local"}

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
        try:
            with urlopen(target_url, timeout=3) as resp:
                data = resp.read(4096).decode("utf-8", errors="replace")
        except URLError as exc:
            abort(502, f"preview failed: {exc.reason}")
        return page("<h2>Preview</h2><pre>" + data + "</pre>")
