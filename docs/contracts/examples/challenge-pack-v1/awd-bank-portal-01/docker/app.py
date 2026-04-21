from flask import Flask, jsonify, request

app = Flask(__name__)

vault = {
    "flag": "BOOTSTRAP",
    "updated_by": "bootstrap",
    "updated_at": "init",
}


@app.get("/")
def index():
    return jsonify(
        {
            "service": "bank-portal",
            "status": "ok",
            "features": ["transfer", "vault", "audit"],
        }
    )


@app.get("/healthz")
def healthz():
    return jsonify({"ok": True, "flag_length": len(vault["flag"])})


@app.put("/api/flag")
def put_flag():
    content_type = (request.headers.get("Content-Type") or "").lower()
    if "application/json" in content_type:
        payload = request.get_json(silent=True) or {}
        flag_value = str(payload.get("flag", "")).strip()
    else:
        flag_value = request.get_data(as_text=True).strip()

    if not flag_value:
        return jsonify({"error": "empty flag"}), 400

    vault["flag"] = flag_value
    vault["updated_by"] = request.headers.get("X-Team", "checker")
    vault["updated_at"] = "runtime"
    return jsonify({"stored": True, "length": len(flag_value)})


@app.get("/api/flag")
def get_flag():
    return jsonify({"flag": vault["flag"]})


@app.get("/api/debug/export")
def debug_export():
    # Vulnerability by design for AWD sample:
    # exporting internal vault data should never be public.
    if request.args.get("scope") == "all":
        return jsonify({"vault": vault, "service": "bank-portal"})
    return jsonify({"error": "scope required"}), 400


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8080)
