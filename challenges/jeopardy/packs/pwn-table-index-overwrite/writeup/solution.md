# 解法

`slots[4]` 后面紧接着 `target`，因此写 `idx=4` 就会越界改到 target。把它写成 win 地址后，程序末尾调用 target 时就会打印 flag。
