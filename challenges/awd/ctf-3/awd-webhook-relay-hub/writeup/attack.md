# Attack Writeup

## 漏洞点

这题的公网入口 `relay-web` 提供 `/preview?url=`，但它只拦截原始字符串里的：

- `127.0.0.1`
- `localhost`
- `0.0.0.0`

它没有限制：

- Docker 内网服务名
- 内部 replay 节点
- replay 节点再去访问 archive 节点的链路

## 利用思路

1. 访问 `/feeds/demo`，可以看到示例 feed 里直接暴露了内部 replay 地址：

```text
http://relay-worker:9091/internal/replay?feed=demo
```

2. 把这个 URL 交给公网预览器，让 `relay-web` 代我们访问：

```text
/preview?url=http://relay-worker:9091/internal/replay?feed=demo
```

3. `relay-worker` 会继续访问 `relay-store` 的 `/internal/archive`，并把返回结果原样回给我们。
4. `relay-store` 的调试归档结果里直接包含当前动态 Flag。

## 关键边界

- `/api/flag` 是 checker 私有通道，不应作为选手攻击面。
- 攻击面应聚焦 `docker/workspace/src/challenge_app.py` 中三段业务角色的协作逻辑。
- 这题需要 `relay-web -> relay-worker -> relay-store` 的真实拓扑边界；压成单容器后，内部链路与防守重点都会失真。
