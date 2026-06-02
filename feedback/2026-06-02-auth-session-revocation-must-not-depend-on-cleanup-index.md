# auth 会话撤销不能依赖清理索引

## 问题描述

这次“改密后用户不会登出”的修复第一版，把用户级会话撤销建立在 Redis `user_sessions` 反向索引上。这样在本地 happy path 看起来成立，但只要遇到两类真实场景就会失效：

- 发布前已经存在的历史 session 没有被补索引。
- 索引维护或清理链路失败时，handler 仍可能把结果当成成功返回。

最终表现是：当前浏览器会退出，但其他设备的旧 session 可能继续有效，安全语义被错误地建立在“清理是否刚好成功”上。

## 原因分析

问题本质不是“少删了几个 key”，而是把安全语义放在了一个可选的清理辅助结构上。

- 反向索引适合做批量清理和存储回收，不适合单独承担认证正确性。
- 安全相关的“用户级会话已失效”应当能在主鉴权链路里被直接判定。
- 一旦主语义和清理辅助结构混在一起，实现者很容易把 best-effort cleanup 误写成 hard guarantee。

## 解决方案

本次修复把语义收口为：

1. `tokenService.RevokeAllUserSessions` 先递增用户级 `session version`。
2. `tokenService.CreateSession` 读取当前 version 并写入 session record。
3. `tokenService.GetSession` 校验 session record 的 version 是否与用户当前 version 一致；不一致则直接判定 session 失效。
4. 反向索引仅保留为撤销后的物理清理优化，不再承担安全正确性。
5. 改密 handler 只有在用户级撤销语义建立成功后才返回成功。

## 收获

- 对认证 / 授权 / 会话这类安全路径，`cleanup succeeded` 不能等价于 `security guarantee established`。
- 如果一个修复要覆盖历史数据、已有会话或跨发布窗口，必须先问一句：没有新辅助索引的旧对象还能不能被主链路识别为失效。
- review 里看到“best-effort”“仅记录日志”“后续可能遗漏”这类表述时，要直接检查它是不是被错误地放在了安全语义 owner 上。

## 沉淀状态

- 状态：仅项目保留
- Owner：`feedback/2026-06-02-auth-session-revocation-must-not-depend-on-cleanup-index.md`
- 链接：
  - `docs/reviews/security/2026-06-01-local-review-password-change-session-revocation.md`
  - `code/backend/internal/module/auth/infrastructure/token_service.go`
