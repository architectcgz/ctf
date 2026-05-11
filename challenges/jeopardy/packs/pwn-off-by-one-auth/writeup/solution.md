# 解法

`name[16]` 后面紧跟一个 `unsigned char admin`。发送 `16` 个填充字节再接 `\x01`，就能用单字节越界把 admin 改成 1。
