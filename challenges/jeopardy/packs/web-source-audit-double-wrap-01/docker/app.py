import base64
import os

from flask import Flask, Response

app = Flask(__name__)

DEV_PLACEHOLDER_FLAG = "flag{dev_local_only_source_audit}"

_NOISE_VARIABLES = [
    ("_0xe02d", "NOT_THE_FLAG_00"),
    ("_0x4e93", "NOT_THE_FLAG_01"),
    ("_0x9586", "KEEP_LOOKING"),
    ("_0x00fa", "ALMOST_THERE"),
    ("_0xde02", "NOT_THE_FLAG_02"),
    ("_0x2035", "NOT_THE_FLAG_03"),
    ("_0xcdf8", "NOT_THE_FLAG_04"),
]


def _b64(text: str) -> str:
    return base64.b64encode(text.encode("utf-8")).decode("utf-8")


def encode_twice(text: str) -> str:
    return _b64(_b64(text))


def current_flag() -> str:
    return os.getenv("CTF_FLAG") or DEV_PLACEHOLDER_FLAG


def _script_variables() -> list[tuple[str, str]]:
    vars_out = [(name, _b64(value)) for name, value in _NOISE_VARIABLES]
    vars_out.append(("_0x4a3f", encode_twice(current_flag())))
    return vars_out


@app.get("/")
def index() -> Response:
    variables = _script_variables()
    variable_lines = "\n".join(f'var {name}="{value}";' for name, value in variables)
    hint_comment = "Y2hlY2sgc3VzcGljaW91cyB2YXJpYWJsZXMgY2FyZWZ1bGx5"

    html = f"""<!DOCTYPE html>
<html>
<!-- flag: FAKE_FLAG_IGNORE_THIS -->
<!-- flag: FAKE_FLAG_NOT_HERE -->
<!-- hint: read every suspicious variable carefully -->
<head>
  <meta charset="utf-8" />
  <title>Web-01</title>
</head>
<body>
  <section>
    <h1>Web-01 源码审计：双层伪装</h1>
    <p>欢迎来到源码审计训练场，请先从页面表面开始观察。</p>
  </section>
  <script>
{variable_lines}
// {hint_comment}
  </script>
</body>
</html>
"""
    return Response(html, mimetype="text/html")


if __name__ == "__main__":
    port = int(os.getenv("PORT", "8080"))
    app.run(host="0.0.0.0", port=port)
