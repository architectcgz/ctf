import os
import secrets
import threading

DEFAULT_FLAG = os.environ.get("FLAG", "flag{awd_tcp_length_gate}")
CHECKER_TOKEN = os.environ.get("CHECKER_TOKEN", "demo-checker-token")
state_lock = threading.Lock()
stored_flag = DEFAULT_FLAG


def set_flag(value):
    global stored_flag
    with state_lock:
        stored_flag = value


def get_flag():
    with state_lock:
        return stored_flag


def authenticate_checker_token(token):
    token = token.strip()
    if not token:
        return False, b"ERR checker token required\n"
    if not secrets.compare_digest(token, CHECKER_TOKEN):
        return False, b"ERR checker auth failed\n"
    return True, b""


def handle_set_flag(writer, token, value):
    ok, message = authenticate_checker_token(token)
    if not ok:
        writer.write(message)
        return
    value = value.strip()
    if not value:
        writer.write(b"ERR empty flag\n")
        return
    set_flag(value)
    writer.write(b"OK\n")


def handle_get_flag(writer, token):
    ok, message = authenticate_checker_token(token)
    if not ok:
        writer.write(message)
        return
    writer.write((get_flag() + "\n").encode("utf-8"))
