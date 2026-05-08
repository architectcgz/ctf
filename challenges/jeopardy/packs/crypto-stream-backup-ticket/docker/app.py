import os

from flask import Flask, abort, render_template_string, request

app = Flask(__name__)

KNOWN_PLAINTEXT = b"backup-ticket:readonly"
SECRET_CODE = b"stream-2026-backup"
KEYSTREAM = bytes(
    [
        0x44, 0x11, 0x26, 0x72, 0x3A, 0x0F, 0x90, 0x23, 0x51, 0x19, 0x83,
        0x17, 0x67, 0x44, 0xA1, 0x38, 0xD2, 0x1C, 0x09, 0x70, 0xB4, 0x2A,
    ]
)


def xor_bytes(data: bytes) -> bytes:
    return bytes(left ^ right for left, right in zip(data, KEYSTREAM))


def current_flag() -> str:
    return os.getenv("FLAG", "").strip() or "flag{local_crypto_stream_backup_ticket}"


INDEX_TPL = """
<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="utf-8" />
    <title>Stream Backup Ticket</title>
    <style>
      body {
        font-family: "Noto Sans SC", sans-serif;
        max-width: 860px;
        margin: 2rem auto;
        line-height: 1.7;
        color: #1f2937;
      }
      .card {
        background: #f8fafc;
        border: 1px solid #dbe4ee;
        border-radius: 14px;
        padding: 1rem 1.25rem;
        margin-bottom: 1rem;
      }
      code, pre {
        background: #e2e8f0;
        border-radius: 8px;
        padding: 0.15rem 0.35rem;
      }
      pre {
        overflow: auto;
        padding: 1rem;
      }
    </style>
  </head>
  <body>
    <h1>Stream Backup Ticket</h1>
    <div class="card">
      <p>已知明文：</p>
      <pre>{{ known_plaintext }}</pre>
      <p>对应密文（hex）：</p>
      <pre>{{ known_cipher }}</pre>
    </div>
    <div class="card">
      <p>未知恢复码密文（hex）：</p>
      <pre>{{ secret_cipher }}</pre>
      <p>拿到恢复码后访问 <code>/redeem?code=...</code>。</p>
    </div>
  </body>
</html>
"""


@app.get("/")
def index():
    return render_template_string(
        INDEX_TPL,
        known_plaintext=KNOWN_PLAINTEXT.decode("utf-8"),
        known_cipher=xor_bytes(KNOWN_PLAINTEXT).hex(),
        secret_cipher=xor_bytes(SECRET_CODE).hex(),
    )


@app.get("/redeem")
def redeem():
    code = request.args.get("code", "").strip()
    if code != SECRET_CODE.decode("utf-8"):
        abort(403, "invalid recovery code")
    return {"flag": current_flag(), "mode": "stream-recovery"}


if __name__ == "__main__":
    port = int(os.getenv("PORT", "8080"))
    app.run(host="0.0.0.0", port=port)
