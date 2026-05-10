# 防守题解

## 可修改位置

- 平台防守入口对应的业务代码在 `/workspace/src/challenge_app.py`
- 题目包内源文件位置是 `docker/workspace/src/challenge_app.py`
- 受保护的 `runtime` 与 `check` 代码不要动

## 主要漏洞点

### 1. 公网预览器可以访问 Docker 内网服务

当前实现只拦截了少量本地地址字面量，但没有限制：

- 内网服务名，例如 `relay-worker`
- 私有地址段
- 只允许固定域名 / 固定上游

推荐修法：

- 先解析 hostname，再做地址判断
- 显式拒绝 Docker 服务名、私网地址、回环地址
- 最稳妥的是只允许固定 allowlist 域名

### 2. replay 节点把 archive 节点的原始结果直接外带

`relay-worker` 当前会把 `relay-store` 的调试结果原样透传。即使保留内部 replay，也不应该把整个 archive payload 回显给上游。

推荐修法：

- replay 节点只返回必要字段
- 对内部调试路由增加服务间鉴权
- 不把 archive 原文直接回传给公网预览结果

### 3. archive 调试结果直接包含 Flag

`relay-store` 的 `/internal/archive` 不应该把当前 Flag 当作调试字段直接输出。

推荐修法：

- 删除 `flag` 字段
- 或改成固定占位值 / 统计信息
- 如果必须保留调试接口，至少加服务间鉴权

## 推荐修改函数 / 角色

- `preview()`
- `internal_replay()`
- `internal_archive()`

## 保活约束

- `/health` 正常返回 200
- `/api/flag` 继续支持 checker 的 PUT / GET
- `/feeds/demo` 仍然可用
- 预览入口仍可抓取合法外部地址

## 交付判断

1. 公网预览入口不能再访问内部 Docker 服务
2. replay / archive 链路不再把 Flag 直接回显
3. `/health`、`/api/flag`、`/feeds/demo` 仍保持可用
