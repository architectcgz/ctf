import os
import socketserver
import threading


DEFAULT_FLAG = os.environ.get("FLAG", "flag{awd_tcp_length_gate}")
state_lock = threading.Lock()
stored_flag = DEFAULT_FLAG


class Handler(socketserver.StreamRequestHandler):
    def handle(self):
        self.wfile.write(b"length-gate ready\n")
        self.wfile.flush()

        for raw_line in self.rfile:
            line = raw_line.decode("utf-8", errors="replace").strip()
            if not line:
                continue

            upper_line = line.upper()
            if upper_line == "PING":
                self.wfile.write(b"PONG\n")
            elif upper_line.startswith("SET_FLAG "):
                self.handle_set_flag(line[len("SET_FLAG ") :])
            elif upper_line == "GET_FLAG":
                self.handle_get_flag()
            elif upper_line.startswith("CHECK "):
                self.handle_length_gate(line[len("CHECK ") :])
            elif upper_line == "HELP":
                self.wfile.write(b"PING | SET_FLAG <flag> | GET_FLAG | CHECK <payload>\n")
            else:
                self.wfile.write(b"ERR unknown command\n")
            self.wfile.flush()

    def handle_set_flag(self, value):
        global stored_flag
        value = value.strip()
        if not value:
            self.wfile.write(b"ERR empty flag\n")
            return
        with state_lock:
            stored_flag = value
        self.wfile.write(b"OK\n")

    def handle_get_flag(self):
        with state_lock:
            value = stored_flag
        self.wfile.write((value + "\n").encode("utf-8"))

    def handle_length_gate(self, payload):
        length = len(payload.encode("utf-8"))
        if length > 40 and "magic" in payload.lower():
            with state_lock:
                value = stored_flag
            self.wfile.write((f"accepted length={length}\n{value}\n").encode("utf-8"))
            return
        self.wfile.write(
            (f"rejected length={length}\nneed a longer payload with magic\n").encode("utf-8")
        )


class Server(socketserver.ThreadingTCPServer):
    allow_reuse_address = True
    daemon_threads = True


if __name__ == "__main__":
    port = int(os.environ.get("PORT", "8080"))
    with Server(("0.0.0.0", port), Handler) as server:
        server.serve_forever()
