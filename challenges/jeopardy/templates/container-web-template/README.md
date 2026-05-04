# 容器 Web 题模板说明

这个模板适合：

- 单容器 Web 题
- 最小化登录、上传、信息泄露、参数处理类题目

复制后你需要做的事情：

1. 改 `challenge.yml`
2. 改 `statement.md`
3. 在 `docker/src/` 下实现真实题目逻辑
4. 确认默认 flag 只用于本地调试
5. 执行 `docker build` 和真实访问验证

镜像入口固定为 `docker/Dockerfile`，不要把 Dockerfile 移到 `docker/src/` 或其他子目录。

## 最低验证命令

```bash
docker build -t test-<your-slug> docker
docker run --rm -p 18080:80 -e FLAG='flag{local_test}' test-<your-slug>
curl http://127.0.0.1:18080/
```

## 自查清单

- `docker build` 能通过
- 按题面路径能访问到服务
- 至少一条真实利用路径能拿到 flag
- 不依赖宿主机外部数据库或其他服务
