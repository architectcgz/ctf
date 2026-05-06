import os
from pathlib import Path

from flask import Response, request

FLAG_PATH = Path("/flag")


def ensure_flag_file():
    if not FLAG_PATH.exists():
        FLAG_PATH.write_text(os.environ.get("FLAG", "flag{local_demo}"), encoding="utf-8")


def register_runtime_routes(app):
    @app.get("/health")
    def health():
        return {"status": "ok"}

    @app.route("/api/flag", methods=["GET", "PUT"])
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
        ensure_flag_file()
        return Response(FLAG_PATH.read_text(encoding="utf-8").strip(), mimetype="text/plain")
