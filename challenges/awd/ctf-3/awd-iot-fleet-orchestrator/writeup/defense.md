# 防守题解

## 可修改位置

- 平台防守入口对应的业务代码在 `/workspace/src/challenge_app.py`
- 题目包内源文件位置是 `docker/workspace/src/challenge_app.py`
- 受保护的 `runtime` 与 `check` 代码不要动

## 主要漏洞点

### 1. fleet-agent 不该保留固定支持 key

`fleet-agent` 把 `fleet-support-2026` 当成全局后门 key，任何知道这个值的人都能越过设备级认证。

推荐修法：

- 删除默认 key fallback
- 只接受每台设备自己的 key 或签名
- 对高风险操作再做一次单独授权

### 2. pull-config 不该把完整配置直接回给公网

`fleet-agent` 当前会把 `config-vault` 的结果原样回给 `fleet-web`，一旦认证失守就会把敏感配置一起带出。

推荐修法：

- `pull-config` 只返回必要摘要
- 把敏感字段留在内网或改成脱敏信息
- agent 与 vault 之间增加服务间鉴权

## 推荐修改函数 / 角色

- `dispatch()`
- `internal_dispatch()`
- `internal_device()`

## 保活约束

- `/health` 正常返回 200
- `/api/flag` 继续支持 checker 的 PUT / GET
- `/devices/demo` 仍然可用
- `/dispatch` 仍可执行合法设备操作

## 交付判断

1. 默认支持 key 不能再调用高权限内部操作
2. 配置读取链路不再直接回显 Flag
3. `/health`、`/api/flag`、`/devices/demo` 仍保持可用
