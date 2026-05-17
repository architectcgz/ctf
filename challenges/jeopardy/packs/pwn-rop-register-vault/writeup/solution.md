# 解法

漏洞点是经典的栈溢出：`read(0, buf, 256)` 会覆盖 `64` 字节缓冲区、saved `rbp` 和返回地址，所以偏移是 `72`。

这题的关键在于目标函数 `reveal(a, b, c)` 会同时检查三个参数，必须按 x86_64 SysV ABI 准备寄存器：

- `rdi = 0x1337c0decafef00d`
- `rsi = 0x4142434445464748`
- `rdx = 0xdeadbeef10203040`

附件里保留了 4 段 gadget：

- `gadget_ret`
- `gadget_pop_rdi`
- `gadget_pop_rsi`
- `gadget_pop_rdx`

最终 payload 结构是：

```text
padding(72)
+ ret
+ pop_rdi + arg1
+ pop_rsi + arg2
+ pop_rdx + arg3
+ reveal
```

连上远程服务发送这段 ROP 链后，`reveal` 会从环境变量 `FLAG` 读取并打印动态 Flag。
