import os
import socketserver
import sys
from pathlib import Path

WORKSPACE_SRC = Path("/workspace/src")
workspace_src = str(WORKSPACE_SRC)
sys.path = [workspace_src] + [path for path in sys.path if path != workspace_src]

from challenge_app import handle_sync_gateway
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
        self.wfile.write(b"sync-gateway ready\n")
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
            elif upper_line.startswith("PUSH "):
                handle_sync_gateway(self.wfile, line)
            elif upper_line == "STATUS":
                handle_sync_gateway(self.wfile, line)
            elif upper_line.startswith("EXPORT "):
                handle_sync_gateway(self.wfile, line)
            elif upper_line == "HELP":
                self.wfile.write(b"PING | PUSH <device> <payload> | STATUS | EXPORT <key> | HELP\n")
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
