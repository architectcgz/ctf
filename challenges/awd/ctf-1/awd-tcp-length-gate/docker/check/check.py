import socket
import sys


def recv_until(sock, marker):
    data = b""
    while marker not in data:
        chunk = sock.recv(4096)
        if not chunk:
            break
        data += chunk
    return data.decode("utf-8", errors="replace")


def main():
    if len(sys.argv) != 3:
        print("usage: check.py <host> <port>")
        return 2
    host = sys.argv[1]
    port = int(sys.argv[2])
    flag = "flag{local_tcp_check}"

    with socket.create_connection((host, port), timeout=3) as sock:
        recv_until(sock, b"ready\n")
        sock.sendall(b"PING\n")
        assert "PONG" in recv_until(sock, b"\n")
        sock.sendall(f"SET_FLAG {flag}\n".encode("utf-8"))
        assert "OK" in recv_until(sock, b"\n")
        sock.sendall(b"GET_FLAG\n")
        assert flag in recv_until(sock, b"\n")

    print("ok")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
