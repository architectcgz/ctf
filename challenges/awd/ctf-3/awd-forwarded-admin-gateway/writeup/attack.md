# Attack Writeup

## 漏洞点

这题的公网入口 `edge-gateway` 会把用户请求里的 `X-Forwarded-For` 和 `X-Forwarded-Host` 原样带给 `admin-app`。而 `admin-app` 错把这些头部当成“来自可信代理”的证据。

## 利用思路

1. 访问 `/reports/demo`，确认公网确实可以通过 `/proxy` 访问内部管理路径。
2. 构造请求：

```text
GET /proxy?path=/internal/export?tenant=blue-lab
X-Forwarded-For: 127.0.0.1
```

3. `edge-gateway` 会把这个伪造头部透传给 `admin-app`。
4. `admin-app` 误以为请求来自可信代理，于是放行 `/internal/export`，再去 `audit-db` 读取租户数据。
5. `audit-db` 返回的数据里直接包含当前动态 Flag。

## 关键边界

- `/api/flag` 是 checker 私有通道，不应作为选手攻击面。
- 攻击面应聚焦 `docker/workspace/src/challenge_app.py` 中网关与 admin 的头部信任关系。
- 这题需要 `edge-gateway -> admin-app -> audit-db` 的真实拓扑边界；压成单容器后，内部管理面与审计面的边界会失真。
