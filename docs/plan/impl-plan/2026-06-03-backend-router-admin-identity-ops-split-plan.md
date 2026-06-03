# 2026-06-03 backend router admin identity / ops split plan

## Objective

- 继续收口 `code/backend/internal/app/router_routes.go` 中 `registerAdminRoutes` 的剩余 owner，把 admin `ops` 与 `identity/session` 路由拆到独立 registrar 文件。
- 让 `registerAdminRoutes` 退化成 admin 入口级分发，不再承载具体业务子域的注册细节。
- 保持现有 API 路径、权限、中间件顺序、session revoke 行为和 handler owner 不变。

## Non-goals

- 不在本轮继续拆 `registerUserRoutes` 或 `registerTeacherAuthoringRoutes`。
- 不在本轮把 session 管理逻辑迁移进 auth / identity module handler。
- 不在本轮改动 `tokenService` contract 或 session API 响应格式。

## Inputs

- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/router_admin_contest_routes.go`
- `code/backend/internal/app/router_session_routes_test.go`
- `code/backend/internal/app/router_test.go`
- `docs/reviews/backend/2026-06-03-backend-router-admin-contest-awd-split-review.md`

## Current problem

- 经过上一刀后，`registerAdminRoutes` 仍直接承载：
  - admin ops 路由：audit logs / dashboard / cheat detection / notifications
  - admin identity 路由：user CRUD / import
  - admin session 管理：list / revoke one / revoke all
- 这些路由虽然比 contest/AWD 少，但依然把不同 owner 混在一个函数里；继续在这里增量加改会复发同样的结构问题。

## Working design

### Target structure for this slice

- 在 `internal/app` 内新增：
  - `router_admin_ops_routes.go`
  - `router_admin_identity_routes.go`
- `registerAdminRoutes` 只负责：
  - 本地 audit helper
  - 委托 `registerAdminOpsRoutes`
  - 委托 `registerAdminIdentityRoutes`
  - 委托现有 `registerAdminContestRoutes`

### Dependency boundary

- `ops` registrar 只依赖审计与 `OpsModule`。
- `identity` registrar 只依赖审计、`identityHandler` 和 `tokenService`。
- session revoke 的匿名 handler 若仍保留在 registrar 文件内，至少收口到 identity registrar 文件，不再留在总入口。

## Task slices

### Slice 1：计划与 reuse-first 门禁

- Goal：补 implementation plan 与 reuse decision，明确这刀只处理 admin ops + identity/session。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-router-admin-identity-ops-split`

### Slice 2：先补结构测试

- Goal：先写测试约束 `registerAdminRoutes` 只做分发，ops 与 identity/session 已迁到独立 registrar 文件。
- Touched files：
  - 新增 registrar 结构测试文件或扩展现有 router 结构测试
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestAdminOpsRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminIdentityRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`

### Slice 3：拆出 registrar 并收窄 deps

- Goal：新增 admin ops / identity registrar 文件，并把剩余路由从 `router_routes.go` 中迁出。
- Touched files：
  - `code/backend/internal/app/router_routes.go`
  - `code/backend/internal/app/router_admin_ops_routes.go`
  - `code/backend/internal/app/router_admin_identity_routes.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestAdminOpsRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminIdentityRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminSessionRoutes' -count=1`

### Slice 4：最小充分回归

- Goal：确认拆分未影响已存在的 route matrix。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes|TestAdminSessionRoutes|TestFullRouter_AccessControlMatrix' -count=1`
  - `bash scripts/check-consistency.sh`

### Slice 5：review 归档

- Goal：记录本轮 self-review 结论与独立 reviewer gate 状态。
- Touched files：
  - `docs/reviews/backend/2026-06-03-backend-router-admin-identity-ops-split-review.md`

## Expected change surface

- `internal/app` admin registrar 进一步拆分
- admin ops / identity/session 局部 deps struct
- registrar 结构测试
- review 记录

## Data / API / compatibility impact

- 无数据结构变更。
- 无 API 路径与权限语义变更。
- 风险主要在 `users/:id` 参数解析顺序、session handler 逻辑迁移时的行为保持，以及 notifications 审计挂载不漂移。

## Review fit check

- Owner 清晰：本轮直接收口 `registerAdminRoutes` 剩余 owner 混杂问题。
- Reuse 清晰：沿用上轮已经建立的 registrar 组织方式，不新增第二套模式。
- 结构收敛：若本轮完成，`registerAdminRoutes` 将退化为单纯分发函数，admin 入口的 oversized owner 基本收口。

## Rollback / recovery

- 纯代码组织调整，可直接回退 registrar 文件和分发调用。
