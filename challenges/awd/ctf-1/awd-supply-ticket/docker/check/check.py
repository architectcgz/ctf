import json
import os
import sys
import time
import urllib.parse
import urllib.request


BASE = sys.argv[1].rstrip("/") if len(sys.argv) > 1 else "http://127.0.0.1:18081"
TOKEN = os.environ.get("CHECKER_TOKEN", "demo-checker-token")


def fail(message):
    raise SystemExit(f"checker failed: {message}")


def expect(condition, message):
    if not condition:
        fail(message)


def request(path, data=None):
    body = None
    headers = {}
    if data is not None:
        body = urllib.parse.urlencode(data).encode()
        headers["Content-Type"] = "application/x-www-form-urlencoded"
    req = urllib.request.Request(BASE + path, data=body, headers=headers)
    with urllib.request.urlopen(req, timeout=5) as resp:
        return resp.status, resp.read().decode(errors="replace")


def get_json(path):
    status, body = request(path)
    try:
        return status, json.loads(body)
    except json.JSONDecodeError:
        fail(f"{path} did not return JSON: {body[:120]}")


def put(path, body):
    req = urllib.request.Request(
        BASE + path,
        data=body.encode(),
        method="PUT",
        headers={"X-AWD-Checker-Token": TOKEN},
    )
    with urllib.request.urlopen(req, timeout=5) as resp:
        return resp.status, resp.read().decode(errors="replace")


def checker_get(path):
    req = urllib.request.Request(BASE + path, headers={"X-AWD-Checker-Token": TOKEN})
    with urllib.request.urlopen(req, timeout=5) as resp:
        return resp.status, resp.read().decode(errors="replace")


suffix = str(int(time.time()))
status, payload = get_json("/health")
expect(status == 200 and payload.get("status") == "ok", "health endpoint is not healthy")
status, _ = request("/new", {"title": "check-" + suffix, "content": "supply check"})
expect(status in (200, 302), "ticket creation failed")
status, body = request("/")
expect(status == 200 and "check-" + suffix in body, "created ticket was not listed")
flag = f"flag{{local_supply_ticket_{suffix}}}"
status, _ = put("/api/flag", flag)
expect(status == 200, "put_flag failed")
status, body = checker_get("/api/flag")
expect(status == 200 and flag in body, "get_flag did not return the flag written by put_flag")
print("ok awd-supply-ticket")
