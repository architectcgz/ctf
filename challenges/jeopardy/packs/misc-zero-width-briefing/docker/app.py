import os
from pathlib import Path

from flask import Flask, abort, render_template_string, request, send_file

APP_DIR = Path(__file__).resolve().parent
DOWNLOADS_DIR = APP_DIR / "notes"
BRIEFING_PATH = DOWNLOADS_DIR / "briefing.md"
REDEEM_CODE = "invisible-sprint-2026"

app = Flask(__name__)


def current_flag() -> str:
    return os.getenv("FLAG", "").strip() or "flag{local_misc_zero_width_briefing}"


INDEX_TPL = """
<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="utf-8" />
    <title>Invisible Sprint Briefing</title>
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
    <h1>Invisible Sprint Briefing</h1>
    <div class="panel">
      <p>下载最新冲刺周报：</p>
      <p><a href="/download/briefing.md">briefing.md</a></p>
      <p>恢复水印后访问 <code>/redeem?code=...</code>。</p>
    </div>
  </body>
</html>
"""


@app.get("/")
def index():
    return render_template_string(INDEX_TPL)


@app.get("/download/briefing.md")
def download_briefing():
    if not BRIEFING_PATH.exists():
        abort(500, "briefing not ready")
    return send_file(BRIEFING_PATH, as_attachment=True, download_name="briefing.md")


@app.get("/redeem")
def redeem():
    code = request.args.get("code", "").strip()
    if code != REDEEM_CODE:
        abort(403, "invalid briefing watermark")
    return {"flag": current_flag(), "source": "briefing"}


if __name__ == "__main__":
    port = int(os.getenv("PORT", "8080"))
    app.run(host="0.0.0.0", port=port)
