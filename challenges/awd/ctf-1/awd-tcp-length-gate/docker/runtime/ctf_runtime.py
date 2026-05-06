import os
import threading

DEFAULT_FLAG = os.environ.get("FLAG", "flag{awd_tcp_length_gate}")
state_lock = threading.Lock()
stored_flag = DEFAULT_FLAG


def set_flag(value):
    global stored_flag
    with state_lock:
        stored_flag = value


def get_flag():
    with state_lock:
        return stored_flag


def handle_set_flag(writer, value):
    value = value.strip()
    if not value:
        writer.write(b"ERR empty flag\n")
        return
    set_flag(value)
    writer.write(b"OK\n")


def handle_get_flag(writer):
    writer.write((get_flag() + "\n").encode("utf-8"))
