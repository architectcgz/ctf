# 2026-06-01 auth 改密后会话撤销修复计划

## Objective

- 修复“用户修改密码后不会可靠退出所有设备”的后端会话失效缺口。
- 修复“会话撤销失败时接口仍返回成功”的错误成功语义。
- 保持前端只在后端确认改密成功后再做本地登出和跳转。

## Non-goals

- 不扩展到 WebSocket 已建立连接的主动断链。
- 不重做整套认证架构，也不引入新的前端页面流程。
- 不顺手处理工作区内这次任务之外的通知页 / 样式改动。

## Inputs

- `docs/reviews/security/2026-06-01-local-review-password-change-session-revocation.md`
- `code/backend/internal/module/auth/infrastructure/token_service.go`
- `code/backend/internal/middleware/auth.go`
- `code/backend/internal/module/auth/api/http/handler.go`
- `code/frontend/src/pages/profile/SecuritySettingsRoutePage.vue`

## Task slices

### Slice 1：补失败测试，锁定回归点

- Goal：先让测试表达两个 review finding。
- Touched files：
  - `code/backend/internal/module/auth/infrastructure/token_service_test.go`
  - `code/backend/internal/module/auth/api/http/http_integration_test.go`
  - `code/frontend/src/pages/profile/__tests__/SecuritySettings.test.ts`
- Validation：
  - `cd code/backend && go test ./internal/module/auth/infrastructure ./internal/module/auth/api/http`
  - `cd code/frontend && npm exec -- vitest run src/pages/profile/__tests__/SecuritySettings.test.ts`
- Review focus：
  - 测试是否真的覆盖“历史 session 无索引”与“撤销失败不能返回成功”两个场景。

### Slice 2：收口后端会话撤销语义

- Goal：让用户级撤销不依赖历史索引完整性，并让鉴权链路能识别改密后的失效会话。
- Touched files：
  - `code/backend/internal/module/auth/infrastructure/token_service.go`
  - `code/backend/internal/module/auth/api/http/handler.go`
- Validation：
  - `cd code/backend && go test ./internal/module/auth/...`
- Review focus：
  - 会话失效 owner 是否明确落在 `tokenService.GetSession` / `RevokeAllUserSessions`
  - 撤销失败是否不再伪装成成功

### Slice 3：对齐前端与 harness

- Goal：确认前端成功路径仍正确，失败路径不误登出，并把本次经验记录进 harness。
- Touched files：
  - `code/frontend/src/pages/profile/__tests__/SecuritySettings.test.ts`
  - `docs/reviews/security/2026-06-01-local-review-password-change-session-revocation.md`
  - `feedback/2026-06-01-password-change-session-revocation-hard-guarantee.md`
- Validation：
  - `cd code/frontend && npm exec -- vitest run src/pages/profile/__tests__/SecuritySettings.test.ts`
  - `bash scripts/check-consistency.sh`
- Review focus：
  - 路由页是否仍是登出跳转 owner
  - harness 记录是否写到正确目录并带沉淀状态

## Expected change surface

- 后端 auth token service / handler / integration tests
- 前端安全设置页测试
- review 文档与 feedback harness 记录

## Data / API / compatibility impact

- 会话记录的 Redis payload 会新增用户级失效判定所需字段或关联键。
- 历史 session 在用户改密后会被强制视为失效，这是预期兼容变化。
- 改密接口在无法完成会话撤销时将不再返回成功。

## Validation matrix

- 历史 session 无反向索引时，改密后旧 session 访问 `/auth/profile` 返回未授权。
- 撤销失败时，改密接口不返回 200 success。
- 正常改密后，前端仍提示成功并跳转登录页。
- 所有相关 Go / Vitest 用例通过。

## Review fit check

- Owner 清晰：用户级会话撤销与会话有效性判断由 `tokenService` 负责；前端只消费成功/失败结果。
- Reuse 点清晰：继续复用现有 `tokenService`、auth middleware、SecuritySettings route owner。
- 结构收敛：本次不再把“删 cookie + 删除部分 session”当成撤销语义，而是收口到服务端鉴权可判定的单点。
- 已知债触达：本次 review 指出的 touched surface 债务就是 auth 会话撤销语义，本轮必须直接收口，不留 follow-up。

## Rollback / recovery

- 若修复导致异常，可回退本次 auth token service 改动；不会涉及 schema migration。
- 若回退代码，需同时回退新增测试与 feedback / review 更新，避免文档与实现漂移。
