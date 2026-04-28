import os
from pathlib import Path

from flask import Flask, abort, render_template_string, request, send_file

APP_DIR = Path(__file__).resolve().parent
NOTES_DIR = APP_DIR / "notes"
RUNTIME_DIR = APP_DIR / "runtime"
FLAG_PATH = RUNTIME_DIR / "flag.txt"

app = Flask(__name__)


def ensure_runtime_files() -> None:
    RUNTIME_DIR.mkdir(parents=True, exist_ok=True)
    flag_value = os.getenv("FLAG", "").strip() or "flag{FLAG_not_set}"
    FLAG_PATH.write_text(flag_value + "\n", encoding="utf-8")


def list_note_names() -> list[str]:
    return sorted(item.name for item in NOTES_DIR.iterdir() if item.is_file())


INDEX_TPL = """
<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="utf-8" />
    <title>内部笔记下载器</title>
    <style>
      body {
        font-family: "Noto Sans SC", sans-serif;
        margin: 2rem auto;
        max-width: 760px;
        line-height: 1.6;
        color: #1f2937;
      }
      h1 {
        margin-bottom: 0.5rem;
      }
      .panel {
        background: #f8fafc;
        border: 1px solid #dbe4ee;
        border-radius: 12px;
        padding: 1rem 1.25rem;
      }
      code {
        background: #eef2f7;
        padding: 0.1rem 0.35rem;
        border-radius: 6px;
      }
    </style>
  </head>
  <body>
    <h1>内部笔记下载器</h1>
    <p>下面是当前可下载的排障笔记：</p>
    <div class="panel">
      <ul>
        {% for note in notes %}
        <li><a href="/download?file={{ note }}">{{ note }}</a></li>
        {% endfor %}
      </ul>
    </div>
    <p>下载接口示例：<code>/download?file=network-check.txt</code></p>
  </body>
</html>
"""


@app.before_request
def prepare() -> None:
    ensure_runtime_files()


@app.get("/")
def index():
    return render_template_string(INDEX_TPL, notes=list_note_names())


@app.get("/download")
def download():
    file_name = request.args.get("file", "").strip()
    if not file_name:
        abort(400, "missing file parameter")

    target_path = NOTES_DIR / file_name
    if not target_path.exists() or not target_path.is_file():
        abort(404, "file not found")

    return send_file(target_path, as_attachment=False, download_name=target_path.name)


if __name__ == "__main__":
    port = int(os.getenv("PORT", "8080"))
    app.run(host="0.0.0.0", port=port)
