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
ASSET_PATH = DATA_DIR / "render_assets.json"
ROLE = os.environ.get("ROLE", "render-web")
RENDER_BASE_URL = os.environ.get("RENDER_BASE_URL", "http://render-worker:9091")
ASSET_BASE_URL = os.environ.get("ASSET_BASE_URL", "http://asset-cache:9092")
DEFAULT_ASSETS = {
    "receipt": {"title": "Receipt Template", "theme": "classic"},
    "label": {"title": "Shipping Label", "theme": "mono"},
}


def setup_business_data():
    DATA_DIR.mkdir(parents=True, exist_ok=True)
    if not ASSET_PATH.exists():
        ASSET_PATH.write_text(
            json.dumps(DEFAULT_ASSETS, ensure_ascii=False, indent=2),
            encoding="utf-8",
        )


def current_flag() -> str:
    return FLAG_PATH.read_text(encoding="utf-8").strip()


def load_assets() -> dict:
    setup_business_data()
    return json.loads(ASSET_PATH.read_text(encoding="utf-8"))


def page(body: str) -> str:
    return (
        "<!doctype html><title>Preview Render Farm</title>"
        "<style>body{font-family:sans-serif;max-width:920px;margin:32px auto;line-height:1.6}"
        "input{margin:8px 0;padding:8px 10px;width:100%;box-sizing:border-box}"
        "button{padding:8px 14px}pre{background:#f5f5f5;padding:14px;border-radius:10px;overflow:auto}"
        "code{background:#f3f4f6;padding:2px 6px;border-radius:6px}</style>"
        "<h1>Preview Render Farm</h1>"
        "<p><a href='/'>首页</a> <a href='/catalog/demo'>素材目录</a></p>"
        + body
    )


def fetch_text(target_url: str) -> str:
    try:
        with urlopen(target_url, timeout=3) as resp:
            return resp.read(4096).decode("utf-8", errors="replace")
    except URLError as exc:
        abort(502, f"render failed: {exc.reason}")


def register_render_web_routes(app):
    @app.get("/")
    def index():
        return page(
            "<p>Preview Render Farm 会把用户选择的素材交给内部 renderer 生成预览。</p>"
            "<form action='/preview' method='get'>"
            "<input name='asset' value='assets/receipt'>"
            "<button type='submit'>预览</button>"
            "</form>"
        )

    @app.get("/catalog/demo")
    def catalog_demo():
        payload = {
            "default_asset": "assets/receipt",
            "preview_path": "/preview?asset=assets/receipt",
            "alternate_asset": "assets/label",
        }
        return Response(json.dumps(payload, ensure_ascii=False), mimetype="application/json")

    @app.get("/preview")
    def preview():
        asset = request.args.get("asset", "").strip()
        if not asset:
            abort(400, "missing asset parameter")
        data = fetch_text(RENDER_BASE_URL + "/internal/render?asset=" + quote(asset, safe="/"))
        return page("<h2>Preview</h2><pre>" + html.escape(data) + "</pre>")


def register_render_worker_routes(app):
    @app.get("/internal/render")
    def internal_render():
        asset = request.args.get("asset", "assets/receipt").strip() or "assets/receipt"
        asset_url = ASSET_BASE_URL + "/internal/" + asset + ".json"
        data = fetch_text(asset_url)
        return Response(data, mimetype="application/json")


def register_asset_cache_routes(app):
    @app.get("/internal/assets/<path:asset_name>.json")
    def asset_json(asset_name: str):
        assets = load_assets()
        record = assets.get(asset_name)
        if record is None:
            abort(404, "unknown asset")
        return Response(json.dumps({"asset": asset_name, "record": record}, ensure_ascii=False), mimetype="application/json")

    @app.get("/internal/debug/flag.json")
    def debug_flag():
        payload = {"asset": "debug/flag", "flag": current_flag(), "scope": "render-cache"}
        return Response(json.dumps(payload, ensure_ascii=False), mimetype="application/json")


def register_challenge_routes(app):
    @app.before_request
    def before_request():
        setup_business_data()

    if ROLE == "render-web":
        register_render_web_routes(app)
        return
    if ROLE == "render-worker":
        register_render_worker_routes(app)
        return
    if ROLE == "asset-cache":
        register_asset_cache_routes(app)
        return

    @app.get("/")
    def unknown_role():
        abort(500, f"unknown role: {ROLE}")
