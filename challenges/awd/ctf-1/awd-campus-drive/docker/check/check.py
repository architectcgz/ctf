import json
import os
import sys
import time
import urllib.parse
import urllib.request


BASE = sys.argv[1].rstrip("/") if len(sys.argv) > 1 else "http://127.0.0.1:18082"
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


boundary = "----ctfcheck"
suffix = str(int(time.time()))
name = f"check-{suffix}.txt"
status, payload = get_json("/health")
expect(status == 200 and payload.get("status") == "ok", "health endpoint is not healthy")
payload = (
    f"--{boundary}\r\n"
    f'Content-Disposition: form-data; name="file"; filename="{name}"\r\n'
    "Content-Type: text/plain\r\n\r\n"
    "drive check\r\n"
    f"--{boundary}--\r\n"
).encode()
req = urllib.request.Request(
    BASE + "/upload",
    data=payload,
    headers={"Content-Type": f"multipart/form-data; boundary={boundary}"},
)
with urllib.request.urlopen(req, timeout=5) as resp:
    expect(resp.status in (200, 302), "file upload failed")
status, body = get("/")
expect(status == 200 and name in body, "uploaded file was not listed")
status, body = get("/preview?" + urllib.parse.urlencode({"path": name}))
expect(status == 200 and "drive check" in body, "uploaded file preview failed")
flag = f"flag{{local_campus_drive_{suffix}}}"
status, _ = put("/api/flag", flag)
expect(status == 200, "put_flag failed")
status, body = checker_get("/api/flag")
expect(status == 200 and flag in body, "get_flag did not return the flag written by put_flag")
print("ok awd-campus-drive")
