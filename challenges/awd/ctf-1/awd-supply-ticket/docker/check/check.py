import sys
import time
import urllib.parse
import urllib.request


BASE = sys.argv[1].rstrip("/") if len(sys.argv) > 1 else "http://127.0.0.1:18081"


def request(path, data=None):
    body = None
    headers = {}
    if data is not None:
        body = urllib.parse.urlencode(data).encode()
        headers["Content-Type"] = "application/x-www-form-urlencoded"
    req = urllib.request.Request(BASE + path, data=body, headers=headers)
    with urllib.request.urlopen(req, timeout=5) as resp:
        return resp.status, resp.read().decode(errors="replace")


def put(path, body):
    req = urllib.request.Request(BASE + path, data=body.encode(), method="PUT")
    with urllib.request.urlopen(req, timeout=5) as resp:
        return resp.status, resp.read().decode(errors="replace")


suffix = str(int(time.time()))
status, _ = request("/health")
assert status == 200
status, _ = request("/new", {"title": "check-" + suffix, "content": "supply check"})
assert status in (200, 302)
status, body = request("/")
assert status == 200 and "check-" + suffix in body
status, _ = put("/api/flag", "flag{local_demo}")
assert status == 200
status, body = request("/api/flag")
assert status == 200 and "flag{" in body
print("ok")
