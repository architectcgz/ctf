from pathlib import Path

from flask import redirect, request, send_file, url_for
from werkzeug.utils import secure_filename

DATA_DIR = Path("/data/uploads")
ALLOWED_EXT = {".txt", ".png", ".jpg", ".jpeg", ".pdf"}


def setup_business_data():
    DATA_DIR.mkdir(parents=True, exist_ok=True)
    sample = DATA_DIR / "course-notice.txt"
    if not sample.exists():
        sample.write_text("本周课程资料已更新。", encoding="utf-8")


def page(body):
    return (
        "<!doctype html><title>校园网盘</title>"
        "<style>body{font-family:sans-serif;max-width:900px;margin:32px auto;line-height:1.5}"
        "input{margin:8px 0}</style>"
        "<h1>校园网盘</h1>"
        "<p><a href='/'>文件列表</a> <a href='/upload'>上传文件</a></p>"
        + body
    )


def register_challenge_routes(app):
    @app.before_request
    def before_request():
        setup_business_data()

    @app.get("/")
    def index():
        links = []
        for path in sorted(DATA_DIR.iterdir()):
            if path.is_file():
                name = path.name
                links.append(
                    f"<li>{name} "
                    f"<a href='/download/{name}'>下载</a> "
                    f"<a href='/preview?path={name}'>预览</a></li>"
                )
        return page("<ul>" + "".join(links) + "</ul>")

    @app.route("/upload", methods=["GET", "POST"])
    def upload():
        if request.method == "POST":
            file = request.files["file"]
            name = secure_filename(file.filename)
            suffix = Path(name).suffix.lower()
            if suffix not in ALLOWED_EXT and not name.endswith(".jpg.php"):
                return page("<p>文件类型不允许。</p>"), 400
            file.save(DATA_DIR / name)
            return redirect(url_for("index"))
        return page("<form method='post' enctype='multipart/form-data'><input type='file' name='file'><button>上传</button></form>")

    @app.get("/download/<name>")
    def download(name):
        safe_name = secure_filename(name)
        return send_file(DATA_DIR / safe_name, as_attachment=True)

    @app.get("/preview")
    def preview():
        raw_path = request.args.get("path", "")
        target = DATA_DIR / raw_path
        if not target.exists():
            return page("<p>文件不存在。</p>"), 404
        data = target.read_text(encoding="utf-8", errors="replace")
        return page("<pre>" + data + "</pre>")
