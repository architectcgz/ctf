# 容器 Pwn 题模板说明

这个模板适合：

- 单容器 Pwn 题
- 提供 HTTP 下载页和 TCP 利用端口的题目
- 基础栈溢出、ROP、格式化字符串、堆利用入门题

复制后你需要做的事情：

1. 改 `challenge.yml`
2. 改 `statement.md`
3. 在 `docker/src/challenge.c` 里实现真实漏洞逻辑
4. 在 `docker/site/index.html` 里改成真实下载页说明
5. 确认默认 flag 只用于本地调试
6. 执行 `docker build` 和真实利用验证

## 模板说明

- 模板默认暴露 `80/tcp` 下载题目二进制
- 模板默认暴露 `9999/tcp` 作为交互利用端口
- 模板在镜像构建时编译 `challenge.c`
- 模板会把编译出的二进制复制到下载页目录
- 模板没有预置 `attachments/challenge`，避免把演示二进制误交付为正式题目材料

如果你希望题包里也附带二进制：

1. 先本地编译或从镜像中导出正式 `challenge`
2. 放到 `attachments/`
3. 把 `challenge.yml` 的 `attachments` 改成真实路径和真实 `sha256`

## 最低验证命令

```bash
docker build -t test-<your-slug> docker
docker run --rm -p 18080:80 -p 19999:9999 -e FLAG='flag{local_test}' test-<your-slug>
curl http://127.0.0.1:18080/
nc 127.0.0.1 19999
```

## 自查清单

- `docker build` 能通过
- 下载页能打开
- 下载链接能下到真实二进制
- 远程端口能连通
- 至少一条真实利用路径能拿到 flag
- 不依赖宿主机外部数据库或其他服务
