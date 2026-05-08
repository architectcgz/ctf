# 防守题解

## 可修改位置

- 平台防守入口对应的业务代码在 `/workspace/src/challenge_app.py`
- 题目包内源文件位置是 `docker/workspace/src/challenge_app.py`
- checker 与 runtime 代码不要动

## 主要漏洞点

### 1. 默认支持密钥是硬编码常量

`SUPPORT_KEY` 直接写成了固定值 `sync-support-2026`，任何人拿到协议文档或做过一次攻击都能重复利用。

推荐修法：

- 不要使用公开常量
- 改成环境变量或你自己设定的高强度随机值
- 更稳妥的方式是直接移除公开 `EXPORT` 能力

### 2. 导出命令直接返回完整 Flag

即使保留导出命令，也不应该把当前动态 Flag 作为公开业务快照返回。

推荐修法：

- 导出结果改成固定状态、计数或脱敏摘要
- 把真正的敏感值继续留给受保护的 checker 通道

## 推荐修改函数

- `handle_sync_gateway()`

## 保活约束

- `PING` 正常返回 `PONG`
- `SET_FLAG` / `GET_FLAG` 继续只接受正确 checker token
- `PUSH` 和 `STATUS` 继续可用

## 交付判断

1. 默认调试密钥不能再直接导出 Flag
2. `EXPORT` 即使保留，也不再返回完整 Flag
3. `PING`、`PUSH`、`STATUS` 与 checker 通道仍可用
