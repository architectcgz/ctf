import json
import os
import urllib.error
import urllib.request

from flask import Flask, Response, abort, render_template_string, request

app = Flask(__name__)


def current_flag() -> str:
    return os.getenv("FLAG", "").strip() or "flag{local_web_mcp_manifest_ssrf}"


INDEX_TPL = """
<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="utf-8" />
    <title>MCP Manifest 预览器</title>
    <style>
      body {
        font-family: "Noto Sans SC", sans-serif;
        margin: 2rem auto;
        max-width: 760px;
        line-height: 1.7;
        color: #1f2937;
      }
      .panel {
        background: #f8fafc;
        border: 1px solid #dbe4ee;
        border-radius: 14px;
        padding: 1rem 1.25rem;
      }
      input {
        width: 100%;
        padding: 0.7rem 0.9rem;
        border: 1px solid #cbd5e1;
        border-radius: 10px;
      }
      button {
        margin-top: 0.8rem;
        padding: 0.65rem 1rem;
        border: 0;
        border-radius: 10px;
        background: #0f172a;
        color: white;
      }
      code, pre {
        background: #e2e8f0;
        border-radius: 8px;
        padding: 0.15rem 0.4rem;
      }
      pre {
        overflow: auto;
        padding: 1rem;
      }
    </style>
  </head>
  <body>
    <h1>MCP Manifest 预览器</h1>
    <p>输入一个工具 manifest 地址，服务端会帮你抓取并预览内容。</p>
    <div class="panel">
      <form action="/preview" method="get">
        <label for="url">Manifest URL</label>
        <input id="url" name="url" value="https://example.com/.well-known/manifest.json" />
        <button type="submit">预览</button>
      </form>
    </div>
    <p>调试说明：已拦截 <code>127.0.0.1</code> 与 <code>localhost</code>。</p>
  </body>
</html>
"""


@app.get("/")
def index():
    return render_template_string(INDEX_TPL)


@app.get("/internal/flag")
def internal_flag():
    if request.remote_addr not in {"127.0.0.1", "::1"}:
        abort(403, "local access only")
    return {"flag": current_flag(), "scope": "debug-local"}


@app.get("/preview")
def preview():
    target_url = request.args.get("url", "").strip()
    if not target_url:
        abort(400, "missing url parameter")
    lowered = target_url.lower()
    if not lowered.startswith(("http://", "https://")):
        abort(400, "only http or https is allowed")
    if "127.0.0.1" in lowered or "localhost" in lowered:
        abort(403, "local targets are blocked")

    try:
        with urllib.request.urlopen(target_url, timeout=3) as resp:
            body = resp.read(4096).decode("utf-8", errors="replace")
            payload = {
                "url": target_url,
                "status": resp.status,
                "preview": body,
            }
            return Response(
                render_template_string(
                    "<h2>Preview Result</h2><pre>{{ body }}</pre>",
                    body=json.dumps(payload, ensure_ascii=False, indent=2),
                ),
                mimetype="text/html; charset=utf-8",
            )
    except urllib.error.HTTPError as exc:
        abort(exc.code, exc.reason)
    except urllib.error.URLError as exc:
        abort(502, f"fetch failed: {exc.reason}")


if __name__ == "__main__":
    port = int(os.getenv("PORT", "8080"))
    app.run(host="0.0.0.0", port=port)
