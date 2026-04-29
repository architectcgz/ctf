import os
import secrets
from pathlib import Path

from flask import Flask, Response, abort, redirect, request, send_file, url_for
from werkzeug.utils import secure_filename

APP = Flask(__name__)
DATA_DIR = Path("/data/uploads")
META_DIR = DATA_DIR / ".meta"
FLAG_PATH = Path("/flag")
ALLOWED_EXT = {".txt", ".png", ".jpg", ".jpeg", ".pdf"}
BLOCKED_SUFFIXES = {".php", ".phtml", ".phar", ".sh", ".py"}


def setup():
    DATA_DIR.mkdir(parents=True, exist_ok=True)
    META_DIR.mkdir(parents=True, exist_ok=True)
    if not FLAG_PATH.exists():
        FLAG_PATH.write_text(os.environ.get("FLAG", "flag{local_demo}"), encoding="utf-8")
    sample = DATA_DIR / "course-notice.txt"
    if not sample.exists():
        sample.write_text("本周课程资料已更新。", encoding="utf-8")
    sample_meta = META_DIR / "course-notice.txt.meta"
    if not sample_meta.exists():
        sample_meta.write_text("course-notice.txt", encoding="utf-8")


def page(body):
    return (
        "<!doctype html><title>校园网盘</title>"
        "<style>body{font-family:sans-serif;max-width:900px;margin:32px auto;line-height:1.5}"
        "input{margin:8px 0}</style>"
        "<h1>校园网盘</h1>"
        "<p><a href='/'>文件列表</a> <a href='/upload'>上传文件</a></p>"
        + body
    )


def list_visible_files():
    result = []
    for meta in sorted(META_DIR.glob("*.meta")):
        storage_name = meta.stem
        original_name = meta.read_text(encoding="utf-8").strip() or storage_name
        file_path = DATA_DIR / storage_name
        if file_path.is_file():
            result.append((storage_name, original_name))
    return result


def is_allowed_upload(name):
    suffixes = [suffix.lower() for suffix in Path(name).suffixes]
    if not suffixes:
        return False
    if suffixes[-1] not in ALLOWED_EXT:
        return False
    return not any(suffix in BLOCKED_SUFFIXES for suffix in suffixes[:-1])


def resolve_storage_name(file_id):
    safe_id = secure_filename(file_id)
    if safe_id != file_id or not safe_id:
        abort(404)
    file_path = DATA_DIR / safe_id
    if not file_path.is_file():
        abort(404)
    return safe_id, file_path


@APP.before_request
def before_request():
    setup()


@APP.get("/")
def index():
    links = []
    for storage_name, original_name in list_visible_files():
        links.append(
            f"<li>{original_name} "
            f"<a href='/download/{storage_name}'>下载</a> "
            f"<a href='/preview/{storage_name}'>预览</a></li>"
        )
    return page("<ul>" + "".join(links) + "</ul>")


@APP.route("/upload", methods=["GET", "POST"])
def upload():
    if request.method == "POST":
        file = request.files["file"]
        original_name = secure_filename(file.filename or "")
        if not original_name or not is_allowed_upload(original_name):
            return page("<p>文件类型不允许。</p>"), 400
        storage_name = f"{secrets.token_hex(8)}{Path(original_name).suffix.lower()}"
        file.save(DATA_DIR / storage_name)
        (META_DIR / f"{storage_name}.meta").write_text(original_name, encoding="utf-8")
        return redirect(url_for("index"))
    return page("<form method='post' enctype='multipart/form-data'><input type='file' name='file'><button>上传</button></form>")


@APP.get("/download/<file_id>")
def download(file_id):
    _, file_path = resolve_storage_name(file_id)
    return send_file(file_path, as_attachment=True)


@APP.get("/preview/<file_id>")
def preview(file_id):
    _, file_path = resolve_storage_name(file_id)
    data = file_path.read_text(encoding="utf-8", errors="replace")
    return page("<pre>" + data.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;") + "</pre>")


@APP.get("/health")
def health():
    return {"status": "ok"}


@APP.route("/api/flag", methods=["GET", "PUT"])
def checker_flag():
    expected_token = os.environ.get("CHECKER_TOKEN", "demo-checker-token")
    if request.headers.get("X-AWD-Checker-Token") != expected_token:
        return {"error": "not found"}, 404
    if request.method == "PUT":
        candidate = request.get_data(as_text=True).strip()
        if not candidate:
            return {"error": "empty flag"}, 400
        FLAG_PATH.write_text(candidate, encoding="utf-8")
        return {"status": "accepted"}
    return Response(FLAG_PATH.read_text(encoding="utf-8").strip(), mimetype="text/plain")


if __name__ == "__main__":
    APP.run(host="0.0.0.0", port=8080)
