附件是一份 64 位 ELF，容器实例提供同一份二进制的远程 TCP 服务。

程序的栈溢出点很明显，但真正打印动态 Flag 的函数不是无参 `win()`，而是会校验三段寄存器参数。

## 目标

1. 分析附件 `challenge.bin`，确认溢出偏移和目标函数逻辑。
2. 构造一条能同时控制 `rdi`、`rsi`、`rdx` 的 ROP 链。
3. 连接实例服务并恢复动态 Flag。

## 附件

- `challenge.bin`

## 访问方式

- 容器服务端口：`8080/tcp`
- 本地测试示例：

```bash
docker build -t test-pwn-rop-register-vault docker
docker run --rm -p 18080:8080 -e FLAG='flag{local_test}' test-pwn-rop-register-vault
nc 127.0.0.1 18080
```
