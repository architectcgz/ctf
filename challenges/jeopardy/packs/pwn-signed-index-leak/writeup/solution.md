# 解法

`items` 前面正好是 `secret[64]`，每个槽位 8 字节，所以 `items[-8]` 就会指向 secret 起始地址，直接打印出 flag。
