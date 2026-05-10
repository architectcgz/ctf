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
TENANT_PATH = DATA_DIR / "audit_tenants.json"
ROLE = os.environ.get("ROLE", "edge-gateway")
ADMIN_BASE_URL = os.environ.get("ADMIN_BASE_URL", "http://admin-app:9091")
AUDIT_BASE_URL = os.environ.get("AUDIT_BASE_URL", "http://audit-db:9092")
DEFAULT_TENANTS = {
    "blue-lab": {
        "owner": "ops-class",
        "ticket": "gw-2026-041",
        "status": "healthy",
    }
}


def setup_business_data():
    DATA_DIR.mkdir(parents=True, exist_ok=True)
    if not TENANT_PATH.exists():
        TENANT_PATH.write_text(
            json.dumps(DEFAULT_TENANTS, ensure_ascii=False, indent=2),
            encoding="utf-8",
        )


def current_flag() -> str:
    return FLAG_PATH.read_text(encoding="utf-8").strip()


def load_tenants() -> dict:
    setup_business_data()
    return json.loads(TENANT_PATH.read_text(encoding="utf-8"))


def page(body: str) -> str:
    return (
        "<!doctype html><title>Forwarded Admin Gateway</title>"
        "<style>body{font-family:sans-serif;max-width:920px;margin:32px auto;line-height:1.6}"
        "input{margin:8px 0;padding:8px 10px;width:100%;box-sizing:border-box}"
        "button{padding:8px 14px}pre{background:#f5f5f5;padding:14px;border-radius:10px;overflow:auto}"
        "code{background:#f3f4f6;padding:2px 6px;border-radius:6px}</style>"
        "<h1>Forwarded Admin Gateway</h1>"
        "<p><a href='/'>首页</a> <a href='/reports/demo'>示例报告</a></p>"
        + body
    )


def fetch_text(target_url: str, headers: dict | None = None) -> str:
    req = Request(target_url, headers=headers or {})
    try:
        with urlopen(req, timeout=3) as resp:
            return resp.read(4096).decode("utf-8", errors="replace")
    except URLError as exc:
        abort(502, f"proxy failed: {exc.reason}")


def register_gateway_routes(app):
    @app.get("/")
    def index():
        return page(
            "<p>公网网关会代运维同学转发内部管理报告，方便查看租户健康状态。</p>"
            "<form action='/proxy' method='get'>"
            "<input name='path' value='/internal/status?tenant=blue-lab'>"
            "<button type='submit'>转发</button>"
            "</form>"
            "<p>真正的 admin 服务并不在公网，所有访问都经过这层 gateway。</p>"
        )

    @app.get("/reports/demo")
    def reports_demo():
        payload = {
            "tenant": "blue-lab",
            "preview_path": "/proxy?path=/internal/status?tenant=blue-lab",
            "note": "gateway forwards selected admin paths",
        }
        return Response(json.dumps(payload, ensure_ascii=False), mimetype="application/json")

    @app.get("/proxy")
    def proxy():
        path = request.args.get("path", "").strip()
        if not path or not path.startswith("/"):
            abort(400, "path must start with /")
        headers = {}
        for name in ("X-Forwarded-For", "X-Forwarded-Host"):
            value = request.headers.get(name)
            if value:
                headers[name] = value
        data = fetch_text(ADMIN_BASE_URL + path, headers=headers)
        return page("<h2>Proxy Result</h2><pre>" + html.escape(data) + "</pre>")


def register_admin_routes(app):
    @app.get("/internal/status")
    def internal_status():
        tenant = request.args.get("tenant", "blue-lab").strip() or "blue-lab"
        tenants = load_tenants()
        record = tenants.get(tenant)
        if record is None:
            abort(404, "unknown tenant")
        payload = {"tenant": tenant, "status": record["status"], "ticket": record["ticket"]}
        return Response(json.dumps(payload, ensure_ascii=False), mimetype="application/json")

    @app.get("/internal/export")
    def internal_export():
        trusted_proxy = request.headers.get("X-Forwarded-For", "") in {"127.0.0.1", "::1"}
        trusted_host = request.headers.get("X-Forwarded-Host", "") == "ops-gateway.local"
        if not trusted_proxy and not trusted_host:
            abort(403, "ops proxy only")
        tenant = request.args.get("tenant", "blue-lab").strip() or "blue-lab"
        data = fetch_text(AUDIT_BASE_URL + "/internal/tenant?name=" + quote(tenant, safe=""))
        return Response(data, mimetype="application/json")


def register_audit_routes(app):
    @app.get("/internal/tenant")
    def internal_tenant():
        tenant = request.args.get("name", "blue-lab").strip() or "blue-lab"
        tenants = load_tenants()
        record = tenants.get(tenant)
        if record is None:
            abort(404, "unknown tenant")
        payload = {
            "tenant": tenant,
            "owner": record["owner"],
            "ticket": record["ticket"],
            "flag": current_flag(),
        }
        return Response(json.dumps(payload, ensure_ascii=False), mimetype="application/json")


def register_challenge_routes(app):
    @app.before_request
    def before_request():
        setup_business_data()

    if ROLE == "edge-gateway":
        register_gateway_routes(app)
        return
    if ROLE == "admin-app":
        register_admin_routes(app)
        return
    if ROLE == "audit-db":
        register_audit_routes(app)
        return

    @app.get("/")
    def unknown_role():
        abort(500, f"unknown role: {ROLE}")
