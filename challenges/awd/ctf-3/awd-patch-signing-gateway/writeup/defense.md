# 防守题解

## 可修改位置

- 平台防守入口对应的业务代码在 `/workspace/src/challenge_app.py`
- 题目包内源文件位置是 `docker/workspace/src/challenge_app.py`
- 受保护的 `runtime` 与 `check` 代码不要动

## 主要漏洞点

### 1. signer 角色不该由公网客户端指定

当前实现把 `signer_role` 直接变成内部信任头 `X-Signer-Role`，等于让攻击者自己声明“我是 release-manager”。

推荐修法：

- 网关层不要接受客户端传入的 signer 角色
- 需要岗位区分时，改为服务端固定映射或独立鉴权
- signer 只信任服务间凭证，不信任普通业务头

### 2. 高权限 bundle 不该直接走 keyset 导出

`signer-core` 只凭一个头就切换到高权限链路，而且把 `key-vault` 的结果原样回给公网入口。

推荐修法：

- 高权限导出必须重新鉴权
- 对外只返回 public manifest，不直接暴露 keyset
- `key-vault` 的结果不要直接透传给公网

## 推荐修改函数 / 角色

- `bundle()`
- `internal_bundle()`
- `keyset()`

## 保活约束

- `/health` 正常返回 200
- `/api/flag` 继续支持 checker 的 PUT / GET
- `/channels/demo` 仍然可用
- `/bundle` 仍可返回合法 public manifest

## 交付判断

1. 公网客户端不能再把自己升级成 `release-manager`
2. 高权限签名链路不再直接回显 Flag
3. `/health`、`/api/flag`、`/channels/demo` 仍保持可用
