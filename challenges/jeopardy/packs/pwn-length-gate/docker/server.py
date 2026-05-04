import os
import socketserver


FLAG = os.environ.get("FLAG", "flag{pwn_length_gate_passed}")


class Handler(socketserver.BaseRequestHandler):
    def handle(self):
        self.request.sendall(b"length gate online\nsend one line:\n")
        data = self.request.recv(256).rstrip(b"\r\n")
        length = len(data)
        if length > 40 and b"magic" in data.lower():
            self.request.sendall(f"accepted length={length}\n{FLAG}\n".encode())
            return
        self.request.sendall(
            f"rejected length={length}\nneed a longer line with the magic word\n".encode()
        )


class Server(socketserver.ThreadingTCPServer):
    allow_reuse_address = True


if __name__ == "__main__":
    with Server(("0.0.0.0", 8080), Handler) as server:
        server.serve_forever()
