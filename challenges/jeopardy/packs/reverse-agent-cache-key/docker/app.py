import os
from pathlib import Path

from flask import Flask, abort, render_template_string, request, send_file

APP_DIR = Path(__file__).resolve().parent
DOWNLOADS_DIR = APP_DIR / "notes"
PYC_PATH = DOWNLOADS_DIR / "cache_gate.pyc"
REDEEM_CODE = "agent-cache-2026-key"

app = Flask(__name__)


def current_flag() -> str:
    return os.getenv("FLAG", "").strip() or "flag{local_reverse_agent_cache_key}"


INDEX_TPL = """
<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="utf-8" />
    <title>Agent Cache Keygen</title>
    <style>
      body {
        font-family: "Noto Sans SC", sans-serif;
        max-width: 760px;
        margin: 2rem auto;
        line-height: 1.7;
        color: #1f2937;
      }
      .panel {
        background: #f8fafc;
        border: 1px solid #dbe4ee;
        border-radius: 14px;
        padding: 1rem 1.25rem;
      }
      code {
        background: #e2e8f0;
        border-radius: 8px;
        padding: 0.15rem 0.35rem;
      }
    </style>
  </head>
  <body>
    <h1>Agent Cache Keygen</h1>
    <p>本地调试 helper 只剩下编译后的 Python 字节码，研发同学让你自己恢复兑换码。</p>
    <div class="panel">
      <p><a href="/download/cache_gate.pyc">下载 cache_gate.pyc</a></p>
      <p>拿到兑换码后访问 <code>/redeem?code=...</code>。</p>
    </div>
  </body>
</html>
"""


@app.get("/")
def index():
    return render_template_string(INDEX_TPL)


@app.get("/download/cache_gate.pyc")
def download_pyc():
    if not PYC_PATH.exists():
        abort(500, "helper artifact missing")
    return send_file(PYC_PATH, as_attachment=True, download_name="cache_gate.pyc")


@app.get("/redeem")
def redeem():
    code = request.args.get("code", "").strip()
    if code != REDEEM_CODE:
        abort(403, "invalid redeem code")
    return {"flag": current_flag(), "source": "agent-cache"}


if __name__ == "__main__":
    port = int(os.getenv("PORT", "8080"))
    app.run(host="0.0.0.0", port=port)
