import sys
import time
import urllib.parse
import urllib.request


BASE = sys.argv[1].rstrip("/") if len(sys.argv) > 1 else "http://127.0.0.1:18082"


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


boundary = "----ctfcheck"
name = f"check-{int(time.time())}.txt"
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
    assert resp.status in (200, 302)
status, body = get("/")
assert status == 200 and name in body
status, body = get("/preview?" + urllib.parse.urlencode({"path": name}))
assert status == 200 and "drive check" in body
status, _ = put("/api/flag", "flag{local_demo}")
assert status == 200
status, body = checker_get("/api/flag")
assert status == 200 and "flag{" in body
print("ok")
