# 解法

结构体里 `name[24]` 后面就是 `approved`。payload 用 24 字节填充，再接 `p32(0x1337)` 即可。
