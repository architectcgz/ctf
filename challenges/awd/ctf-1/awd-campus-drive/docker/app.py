from flask import Flask

from challenge_app import register_challenge_routes
from ctf_runtime import ensure_flag_file, register_runtime_routes

APP = Flask(__name__)


@APP.before_request
def before_runtime_request():
    ensure_flag_file()


register_runtime_routes(APP)
register_challenge_routes(APP)


if __name__ == "__main__":
    APP.run(host="0.0.0.0", port=8080)
