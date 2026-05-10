# 防守题解

## 可修改位置

- 平台防守入口对应的业务代码在 `/workspace/src/challenge_app.py`
- 题目包内源文件位置是 `docker/workspace/src/challenge_app.py`
- 受保护的 `runtime` 与 `check` 代码不要动

## 主要漏洞点

### 1. 公网网关不该信任客户端的 forwarded 头

当前实现把客户端提供的 `X-Forwarded-For` / `X-Forwarded-Host` 原样转发给内部 admin 服务，这等于把“我是谁”交给攻击者自己填写。

推荐修法：

- 在网关层直接丢弃用户传入的 forwarded 头
- 需要透传时，由网关自己重写可信值
- 内部服务不要把普通请求头当成可信代理认证

### 2. admin 导出只靠 forwarded 头做鉴权

`admin-app` 对 `/internal/export` 的放行条件过弱，只要 forwarded 头看起来像内网就会放行。

推荐修法：

- 使用服务间 secret、mTLS 或固定反向代理注入头
- 不接受来自公网请求的“自报家门”
- 即使是内部管理路由，也应缩小导出内容范围

### 3. audit 数据不该直接包含 Flag

`audit-db` 的导出结果直接拼进了当前 Flag，这让任何误放行都会变成直接失陷。

推荐修法：

- 删除 `flag` 字段
- 导出只保留租户状态、工单号等必要字段
- 对内部数据查询增加最小权限检查

## 推荐修改函数 / 角色

- `proxy()`
- `internal_export()`
- `internal_tenant()`

## 保活约束

- `/health` 正常返回 200
- `/api/flag` 继续支持 checker 的 PUT / GET
- `/reports/demo` 仍然可用
- `/proxy` 仍能访问允许的内部只读路径

## 交付判断

1. 客户端伪造 forwarded 头后，公网请求不能再拿到导出数据
2. admin / audit 链路不再直接回显 Flag
3. `/health`、`/api/flag`、`/reports/demo` 仍保持可用
