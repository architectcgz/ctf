import os
import socketserver
import sys
from pathlib import Path

WORKSPACE_SRC = Path("/workspace/src")
if str(WORKSPACE_SRC) not in sys.path:
    sys.path.insert(0, str(WORKSPACE_SRC))

from challenge_app import handle_length_gate
from ctf_runtime import handle_get_flag, handle_set_flag


def parse_checker_command(line, command):
    parts = line.split(" ", 2 if command == "SET_FLAG" else 1)
    if command == "SET_FLAG":
        if len(parts) < 3:
            return "", ""
        return parts[1], parts[2]
    if len(parts) < 2:
        return "", ""
    return parts[1], ""


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
                token, flag_value = parse_checker_command(line, "SET_FLAG")
                handle_set_flag(self.wfile, token, flag_value)
            elif upper_line == "GET_FLAG":
                self.wfile.write(b"ERR checker token required\n")
            elif upper_line.startswith("GET_FLAG "):
                token, _ = parse_checker_command(line, "GET_FLAG")
                handle_get_flag(self.wfile, token)
            elif upper_line.startswith("CHECK "):
                handle_length_gate(self.wfile, line[len("CHECK ") :])
            elif upper_line == "HELP":
                self.wfile.write(b"PING | CHECK <payload> | HELP\n")
            else:
                self.wfile.write(b"ERR unknown command\n")
            self.wfile.flush()


class Server(socketserver.ThreadingTCPServer):
    allow_reuse_address = True
    daemon_threads = True


if __name__ == "__main__":
    port = int(os.environ.get("PORT", "8080"))
    with Server(("0.0.0.0", port), Handler) as server:
        server.serve_forever()
