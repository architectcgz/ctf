# 解法

程序本身只是一个很薄的解释器：

1. 读取 `program.blk`
2. 根据 header 里的 `entry` 和 `seed` 初始化状态
3. 每次按 block id 找到下一个记录
4. 对记录字段去掩码后，执行一轮字节校验

单条记录实际字段是：

```text
id, next, index, xor_key, add_key, rol_bits, expected
```

但除 `id` 外其余字段都按下面的掩码做了异或：

```text
mask = (id * 29 + pos * 11 + 0x5A) & 0xff
```

校验流程是：

```text
x = input[index]
x ^= xor_key
x ^= state
x = (x + add_key) & 0xff
x = rol8(x, rol_bits)
assert x == expected
state = expected ^ index ^ 0xA5
```

因为 `expected`、`add_key`、`xor_key`、`rol_bits` 和状态更新全都已知，所以可以反着算：

1. `ror8(expected, rol_bits)`
2. 减去 `add_key`
3. 再异或 `xor_key` 和上一步的 `state`

按 `next` 指针走完整条链后，就能把每个 `index` 对应的输入字节填回去，最终得到 flag。
