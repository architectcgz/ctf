import sys
from pathlib import Path
import os

from flask import Flask

WORKSPACE_SRC = Path("/workspace/src")
if str(WORKSPACE_SRC) not in sys.path:
    sys.path.insert(0, str(WORKSPACE_SRC))

from challenge_app import register_challenge_routes
from ctf_runtime import ensure_flag_file, register_runtime_routes

APP = Flask(__name__)
ROLE = os.environ.get("ROLE", "relay-web")
APP_PORT = int(os.environ.get("APP_PORT", "8080"))


@APP.before_request
def before_runtime_request():
    ensure_flag_file()


register_runtime_routes(APP)
register_challenge_routes(APP)


if __name__ == "__main__":
    APP.run(host="0.0.0.0", port=APP_PORT)
