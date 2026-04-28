import json
import sys
import time
import urllib.request


BASE = sys.argv[1].rstrip("/") if len(sys.argv) > 1 else "http://127.0.0.1:18083"


def post(path, payload, key=None):
    headers = {"Content-Type": "application/json"}
    if key:
        headers["X-Device-Key"] = key
    req = urllib.request.Request(
        BASE + path,
        data=json.dumps(payload).encode(),
        headers=headers,
    )
    with urllib.request.urlopen(req, timeout=5) as resp:
        return resp.status, resp.read().decode(errors="replace")


def get(path):
    with urllib.request.urlopen(BASE + path, timeout=5) as resp:
        return resp.status, resp.read().decode(errors="replace")


def checker_get(path):
    req = urllib.request.Request(BASE + path, headers={"X-AWD-Checker-Token": "demo-checker-token"})
    with urllib.request.urlopen(req, timeout=5) as resp:
        return resp.status, resp.read().decode(errors="replace")


def put(path, body):
    req = urllib.request.Request(
        BASE + path,
        data=body.encode(),
        method="PUT",
        headers={"X-AWD-Checker-Token": "demo-checker-token"},
    )
    with urllib.request.urlopen(req, timeout=5) as resp:
        return resp.status, resp.read().decode(errors="replace")


suffix = str(int(time.time()))
status, _ = post(
    "/api/telemetry",
    {"topic": "team1/device-a/status", "payload": "online-" + suffix},
    key="device-demo-key",
)
assert status == 200
status, body = get("/api/topic?allow=team1/&topic=team1/device-a")
assert status == 200 and "online-" + suffix in body
status, body = get("/firmware?name=sensor-v1.bin")
assert status == 200 and "demo firmware" in body
status, _ = put("/api/flag", "flag{local_demo}")
assert status == 200
status, body = checker_get("/api/flag")
assert status == 200 and "flag{" in body
print("ok")
