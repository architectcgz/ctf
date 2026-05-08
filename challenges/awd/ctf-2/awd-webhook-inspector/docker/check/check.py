import json
import os
import sys
import urllib.request


BASE = sys.argv[1].rstrip("/") if len(sys.argv) > 1 else "http://127.0.0.1:18086"
TOKEN = os.environ.get("CHECKER_TOKEN", "demo-checker-token")


def fail(message):
    raise SystemExit(f"checker failed: {message}")


def expect(condition, message):
    if not condition:
        fail(message)


def get(path):
    with urllib.request.urlopen(BASE + path, timeout=5) as resp:
        return resp.status, resp.read().decode(errors="replace")


def get_json(path):
    status, body = get(path)
    try:
        return status, json.loads(body)
    except json.JSONDecodeError:
        fail(f"{path} did not return JSON: {body[:120]}")


def checker_get(path):
    req = urllib.request.Request(BASE + path, headers={"X-AWD-Checker-Token": TOKEN})
    with urllib.request.urlopen(req, timeout=5) as resp:
        return resp.status, resp.read().decode(errors="replace")


def put(path, body):
    req = urllib.request.Request(
        BASE + path,
        data=body.encode(),
        method="PUT",
        headers={"X-AWD-Checker-Token": TOKEN},
    )
    with urllib.request.urlopen(req, timeout=5) as resp:
        return resp.status, resp.read().decode(errors="replace")


status, payload = get_json("/health")
expect(status == 200 and payload.get("status") == "ok", "health endpoint is not healthy")
status, body = get("/")
expect(status == 200 and "Webhook Inspector" in body, "index page missing")
status, body = get("/manifest/demo")
expect(status == 200 and "webhook-inspector-demo" in body, "demo manifest is unavailable")
flag = "flag{local_webhook_inspector_check}"
status, _ = put("/api/flag", flag)
expect(status == 200, "put_flag failed")
status, body = checker_get("/api/flag")
expect(status == 200 and flag in body, "get_flag did not return the flag written by put_flag")
print("ok awd-webhook-inspector")
