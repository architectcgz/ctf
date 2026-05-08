import os
from pathlib import Path

from flask import Flask, abort, render_template_string, request, send_file

APP_DIR = Path(__file__).resolve().parent
DOWNLOADS_DIR = APP_DIR / "notes"
BUNDLE_PATH = DOWNLOADS_DIR / "preview-bundle.zip"
TOKEN = "preview-artifact-2026-token"

app = Flask(__name__)


def current_flag() -> str:
    return os.getenv("FLAG", "").strip() or "flag{local_forensics_ci_preview_artifact}"


INDEX_TPL = """
<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="utf-8" />
    <title>CI Preview Artifact Leak</title>
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
    <h1>CI Preview Artifact Leak</h1>
    <div class="panel">
      <p>下载误公开的工件包：</p>
      <p><a href="/download/preview-bundle.zip">preview-bundle.zip</a></p>
      <p>拿到 token 后访问 <code>/redeem?token=...</code>。</p>
    </div>
  </body>
</html>
"""


@app.get("/")
def index():
    return render_template_string(INDEX_TPL)


@app.get("/download/preview-bundle.zip")
def download_bundle():
    if not BUNDLE_PATH.exists():
        abort(500, "artifact bundle missing")
    return send_file(BUNDLE_PATH, as_attachment=True, download_name="preview-bundle.zip")


@app.get("/redeem")
def redeem():
    token = request.args.get("token", "").strip()
    if token != TOKEN:
        abort(403, "invalid preview token")
    return {"flag": current_flag(), "source": "ci-preview"}


if __name__ == "__main__":
    port = int(os.getenv("PORT", "8080"))
    app.run(host="0.0.0.0", port=port)
