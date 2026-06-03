# 2026-06-03 backend test architecture phase 11 plan

## Objective

- 把 `TestFullRouter_AdminChallengeManagementStateMatrix` 的 HTTP 场景断言 owner 迁到 `code/backend/tests/system/http/fullrouteradmin`。
- 让 `internal/app/full_router_admin_state_matrix_test.go` 只保留 glue code、剩余 admin 场景和本地 helper。
- 继续沿用前几轮已验证的“先迁断言 owner，后拆 helper owner”节奏。

## Non-goals

- 本轮不迁移 `TestFullRouter_AdminOpsAndNotificationStateMatrix`。
- 本轮不迁移 `TestFullRouter_AdminImagesCapsOversizedPageSize`。
- 本轮不抽 `createPracticeSubmission`、challenge status 更新、instance seed 或 topology helper 到共享 testutil。
- 本轮不改生产逻辑或 runtime 行为。

## Inputs

- `code/backend/internal/app/full_router_admin_state_matrix_test.go`
- `code/backend/tests/system/http/fullrouteradmin/admin_flow.go`
- `.harness/reuse-decisions/backend-test-architecture-phase11.md`

## Current problem

- `full_router_admin_state_matrix_test.go` 仍把 challenge lifecycle、writeup、manual review、topology 等长 HTTP 场景断言放在 `internal/app`。
- 同文件还混着 admin ops/notification 场景与小型 page size 回归，不适合一刀全抽。
- 如果把 challenge 场景 owner、ops 场景 owner 和 helper owner 混做，会让本轮 scope 过宽。

## Working design

### Target structure

- 继续复用 `code/backend/tests/system/http/fullrouteradmin/`：
  - 新增 `VerifyAdminChallengeManagementStateMatrix`
- 新增 driver 只承接 HTTP 请求、断言以及少量 callback。
- `internal/app/full_router_admin_state_matrix_test.go` 继续负责：
  - 构造 `newFullRouterTestEnv`
  - 登录拿 token / headers
  - 提供 DB status 更新、practice submission seed、instance lifecycle callback
  - 保留其余 admin 场景

### Boundary choice

- 只迁 `AdminChallengeManagementStateMatrix`。
- 通过 callback/driver 复用现有 `internal/app` helper，而不是在 `tests/system/http` 复制 env 或 seed 逻辑。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 phase 11 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-test-architecture-phase11`

### Slice 2：迁移 admin challenge management 场景断言 owner

- Goal：把 `TestFullRouter_AdminChallengeManagementStateMatrix` 迁到 `tests/system/http/fullrouteradmin`。
- Touched files：
  - `code/backend/internal/app/full_router_admin_state_matrix_test.go`
  - `code/backend/tests/system/http/fullrouteradmin/admin_flow.go`
  - `code/backend/tests/README.md`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestFullRouter_AdminChallengeManagementStateMatrix' -count=1`

### Slice 3：回归与 workflow gate

- Goal：确认 practice/router 相关现有行为和 workflow 约束未被破坏。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_|TestRouter' -count=1`
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- admin challenge management 场景测试 owner
- `fullrouteradmin` 目录下新增一段可复用场景断言

## Data / API / compatibility impact

- 无生产数据、API 或业务语义变化。
- 风险主要在：
  - challenge / writeup / manual review / topology 断言迁出后发生语义漂移
  - teacher 与 other teacher 的权限分支遗漏
  - 删除 challenge 前的实例阻塞断言在 callback 迁接时断裂

## Review fit check

- Owner 清晰：admin challenge management 的 HTTP 场景断言迁到 `tests/system/http/fullrouteradmin`。
- Reuse 清晰：继续复用 full router fixture 和现有 seed helper，不在本轮复制 env。
- 结构收敛：这刀只解决 admin challenge management owner，不把 ops/notification 和 helper 抽取混进来。

## Rollback / recovery

- 纯测试断言 owner 调整，可直接回退到当前 `internal/app/full_router_admin_state_matrix_test.go` 版本。
