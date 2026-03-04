import os

from flask import Flask

app = Flask(__name__)


@app.get("/")
def index():
    flag_value = os.getenv("CTF_FLAG", "")
    if flag_value:
        return f"CTF_FLAG={flag_value}\n"
    return "CTF_FLAG is not set (platform should inject it at container run time)\n", 500


if __name__ == "__main__":
    port = int(os.getenv("PORT", "8080"))
    app.run(host="0.0.0.0", port=port)

