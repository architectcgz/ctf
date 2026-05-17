# 解法

栈溢出点在主函数的 `read`，覆盖返回地址后不能只跳一个无参 `win`，而是要调用带 3 个参数的 `reveal`。

这题关键是按 x86_64 SysV ABI 准备寄存器：

- 第 1 个参数进 `rdi`
- 第 2 个参数进 `rsi`
- 第 3 个参数进 `rdx`

二进制里专门保留了 `pop rdi; ret`、`pop rsi; ret`、`pop rdx; ret` 和一个单独的 `ret`。栈偏移是 `72` 字节，所以 payload 结构就是：

```text
padding(72)
+ ret
+ pop_rdi + arg1
+ pop_rsi + arg2
+ pop_rdx + arg3
+ reveal
```

把三段常量按小端写进去后，本地执行即可打印 flag。
