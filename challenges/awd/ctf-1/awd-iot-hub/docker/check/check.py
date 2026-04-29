import json
import sys
import time
import urllib.request


BASE = sys.argv[1].rstrip("/") if len(sys.argv) > 1 else "http://127.0.0.1:18083"
TOKEN = "demo-checker-token"


def fail(message):
    raise SystemExit(f"checker failed: {message}")


def expect(condition, message):
    if not condition:
        fail(message)


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


suffix = str(int(time.time()))
status, payload = get_json("/health")
expect(status == 200 and payload.get("status") == "ok", "health endpoint is not healthy")
status, _ = post(
    "/api/telemetry",
    {"topic": "team1/device-a/status", "payload": "online-" + suffix},
    key="device-demo-key",
)
expect(status == 200, "telemetry write failed")
status, body = get_json("/api/topic?allow=team1/&topic=team1/device-a")
expect(
    status == 200 and any(item.get("payload") == "online-" + suffix for item in body),
    "telemetry read did not return the written payload",
)
status, body = get("/firmware?name=sensor-v1.bin")
expect(status == 200 and "demo firmware" in body, "firmware read failed")
flag = f"flag{{local_iot_hub_{suffix}}}"
status, _ = put("/api/flag", flag)
expect(status == 200, "put_flag failed")
status, body = checker_get("/api/flag")
expect(status == 200 and flag in body, "get_flag did not return the flag written by put_flag")
print("ok awd-iot-hub")
