# 2026-06-03 backend router authoring split plan

## Objective

- 把 `code/backend/internal/app/router_routes.go` 中的 `registerTeacherAuthoringRoutes` 从单个大函数拆成稳定的 authoring registrar 组合。
- 让普通题 authoring、authoring 资源资产、AWD authoring 各自落到独立文件，收口当前 authoring surface 的 owner 混杂。
- 保持现有 URL、权限、中间件顺序、handler owner 和测试语义不变。

## Non-goals

- 不在本轮修改 `registerAdminRoutes`、`registerUserRoutes` 或其它已经拆好的 registrar。
- 不在本轮调整 `/api/v1/authoring/*` 的权限模型、ownerGuard 逻辑或 handler 实现。
- 不在本轮修复既有 runtime / practice 集成测试里的环境性失败。

## Inputs

- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `docs/reviews/backend/archive/2026-06/2026-06-03-backend-router-user-teacher-split-review.md`

## Current problem

- `registerTeacherAuthoringRoutes` 当前同时承载：
  - 普通题导入、题目 CRUD、publish/self-check、writeup、topology、package export、flag
  - 镜像与环境模板这类 authoring 共享资源
  - AWD 导入与 AWD 题目 CRUD
- 这些路由虽然都挂在 `/authoring/*` 下，但业务 owner 已经不是一类东西，继续堆在同一个 registrar 里只会让后续 authoring 改动重新回到一把大函数。

## Working design

### Target structure

- 保留 `registerTeacherAuthoringRoutes(adminAuthoring, deps)` 作为 authoring 总入口。
- 在 `internal/app` 内新增：
  - `router_authoring_challenge_routes.go`
  - `router_authoring_asset_routes.go`
  - `router_authoring_awd_routes.go`
- `registerTeacherAuthoringRoutes` 只做三次委托：
  - `registerAuthoringChallengeRoutes(adminAuthoring, deps)`
  - `registerAuthoringAssetRoutes(adminAuthoring, deps)`
  - `registerAuthoringAWDRoutes(adminAuthoring, deps)`

### Boundary choice

- `authoring challenge` 文件负责：
  - `challenge-imports/*`
  - `challenges/*`
  - challenge publish/self-check/writeup/topology/package export/flag
- `authoring asset` 文件负责：
  - `images/*`
  - `environment-templates/*`
- `authoring awd` 文件负责：
  - `awd-challenge-imports/*`
  - `awd-challenges/*`

### Grill check

- 不按“HTTP method”或“是否带 ownerGuard”切分，因为那会把同一个业务 owner 打散。
- 不把 environment template 塞回 challenge file；它和 image 一样都属于 authoring 可复用资产，更适合形成资源 owner。
- AWD challenge 独立成组，是因为它和普通题目的 handler、import 流程、payload 契约都已经分叉。

## Task slices

### Slice 1：计划与 reuse-first 门禁

- Goal：补 implementation plan 与 reuse decision。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-router-authoring-split`

### Slice 2：先补结构测试

- Goal：用结构测试约束 `registerTeacherAuthoringRoutes` 中的不同 owner 路由必须迁到独立 registrar 文件。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestAuthoringChallengeRoutesAreExtractedIntoDedicatedRegistrarFile|TestAuthoringAssetRoutesAreExtractedIntoDedicatedRegistrarFile|TestAuthoringAWDRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`

### Slice 3：拆出 authoring registrars

- Goal：把 `registerTeacherAuthoringRoutes` 拆成 challenge / asset / awd 三个独立 registrar 文件，并收窄 deps。
- Touched files：
  - `code/backend/internal/app/router_routes.go`
  - `code/backend/internal/app/router_authoring_challenge_routes.go`
  - `code/backend/internal/app/router_authoring_asset_routes.go`
  - `code/backend/internal/app/router_authoring_awd_routes.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestAuthoringChallengeRoutesAreExtractedIntoDedicatedRegistrarFile|TestAuthoringAssetRoutesAreExtractedIntoDedicatedRegistrarFile|TestAuthoringAWDRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`

### Slice 4：最小充分回归

- Goal：确认 authoring 路径存在性、owner guard 和 access matrix 不漂移。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestNewRouterRegistersStudentChallengeRoutes|TestFullRouter_AccessControlMatrix|TestAuthoringChallengeRoutesAreExtractedIntoDedicatedRegistrarFile|TestAuthoringAssetRoutesAreExtractedIntoDedicatedRegistrarFile|TestAuthoringAWDRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`
  - `bash scripts/check-consistency.sh`

### Slice 5：review 归档

- Goal：记录本轮 self-review 与独立 reviewer gate 状态。
- Touched files：
  - `docs/reviews/backend/archive/2026-06/2026-06-03-backend-router-authoring-split-review.md`

## Expected change surface

- `internal/app` authoring route registrar 拆分
- 局部 deps struct
- 结构测试
- review 记录

## Data / API / compatibility impact

- 无数据结构变更。
- 无 API 路径、method、权限变更。
- 风险主要在：
  - `ownerGuard` 挂载顺序漂移
  - `challenge-imports` / `awd-challenge-imports` 错分到错误 owner
  - `images` / `environment-templates` 这类共享资源从 authoring 总入口迁出时漏注册

## Review fit check

- Owner 清晰：本轮完成后 `/authoring` 总入口不再承载三个子域的具体注册细节。
- Reuse 清晰：沿用前三刀已经建立的 registrar 拆分模式。
- 结构收敛：一旦触碰 `registerTeacherAuthoringRoutes`，本轮直接完成该 surface 的 owner 收口，而不是只拆一段 challenge CRUD。

## Rollback / recovery

- 纯代码组织调整，可直接回退新增 registrar 文件与分发调用。
