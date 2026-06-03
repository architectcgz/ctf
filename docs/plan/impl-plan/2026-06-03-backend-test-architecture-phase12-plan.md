# 2026-06-03 backend test architecture phase 12 plan

## Objective

- 把 `TestFullRouter_AdminOpsAndNotificationStateMatrix` 的 HTTP 场景断言 owner 迁到 `code/backend/tests/system/http/fullrouteradmin`。
- 让 `internal/app/full_router_admin_state_matrix_test.go` 基本只保留 page size 小回归与 glue code。
- 继续沿用前几轮已验证的“先迁断言 owner，后拆 helper owner”节奏。

## Non-goals

- 本轮不迁移 `TestFullRouter_AdminImagesCapsOversizedPageSize`。
- 本轮不抽 `performFullRouterMultipartRequest`、`receiveFullRouterWSMessageByType` 到共享 testutil。
- 本轮不改生产逻辑或 runtime 行为。

## Inputs

- `code/backend/internal/app/full_router_admin_state_matrix_test.go`
- `code/backend/tests/system/http/fullrouteradmin/admin_flow.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `.harness/reuse-decisions/backend-test-architecture-phase12.md`

## Current problem

- `full_router_admin_state_matrix_test.go` 仍把 image / user admin / dashboard / audit / cheat detection / notification / websocket 的长 HTTP 场景断言留在 `internal/app`。
- 同文件只剩一个小的 page size 回归适合继续贴近 `package app`。
- 如果把小回归也一起迁走，收益不高，反而会让 scope 掺进 helper owner 判断。

## Working design

### Target structure

- 继续复用 `code/backend/tests/system/http/fullrouteradmin/`：
  - 新增 `VerifyAdminOpsAndNotificationStateMatrix`
- 新 driver 承接：
  - 普通 request
  - multipart request
  - router 本体
  - dashboard / audit log seed callback
- `internal/app/full_router_admin_state_matrix_test.go` 继续负责：
  - 构造 `newFullRouterTestEnv`
  - 登录拿 token / headers
  - 提供 cache / DB seed callback
  - 保留 page size 小回归

### Boundary choice

- 只迁 `AdminOpsAndNotificationStateMatrix`。
- multipart / websocket helper 暂时作为 package-local helper 或 callback 复用，不在本轮改 owner。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 phase 12 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-test-architecture-phase12`

### Slice 2：迁移 admin ops / notification 场景断言 owner

- Goal：把 `TestFullRouter_AdminOpsAndNotificationStateMatrix` 迁到 `tests/system/http/fullrouteradmin`。
- Touched files：
  - `code/backend/internal/app/full_router_admin_state_matrix_test.go`
  - `code/backend/tests/system/http/fullrouteradmin/admin_flow.go`
  - `code/backend/tests/README.md`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestFullRouter_AdminOpsAndNotificationStateMatrix' -count=1`

### Slice 3：回归与 workflow gate

- Goal：确认 practice/router 相关现有行为和 workflow 约束未被破坏。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_|TestRouter' -count=1`
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- admin ops / notification 场景测试 owner
- `full_router_admin_state_matrix_test.go` 进一步变薄

## Data / API / compatibility impact

- 无生产数据、API 或业务语义变化。
- 风险主要在：
  - multipart import 场景迁出后请求构造漂移
  - WebSocket ticket / 连接验证迁出后 message assert 漏掉
  - dashboard / audit / cheat detection 的 seed callback 接线出错

## Review fit check

- Owner 清晰：admin ops / notification 的 HTTP 场景断言迁到 `tests/system/http/fullrouteradmin`。
- Reuse 清晰：继续复用 full router fixture、multipart helper 语义和现有 DB/cache seed。
- 结构收敛：这刀不碰 page size 小回归，也不顺手抽共享 helper。

## Rollback / recovery

- 纯测试断言 owner 调整，可直接回退到当前 `internal/app/full_router_admin_state_matrix_test.go` 版本。
