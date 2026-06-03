# 2026-06-03 backend internal app full router integration test split plan

## Objective

- 把 `code/backend/internal/app/full_router_integration_test.go` 中混合的顶层集成测试按领域拆到多个 `_test.go` 文件。
- 保留原文件作为 `fullRouterTestEnv`、route helper 和 build wiring 测试的共享 owner。
- 继续降低 `internal/app` 1000+ 行测试文件的 owner 混杂程度。

## Non-goals

- 本轮不改业务 handler、路由、权限、runtime 逻辑或测试语义。
- 本轮不重写 `newFullRouterTestEnv`、`seedFullRouterData`、access helper 或 smoke helper。
- 本轮不继续处理 `practice_flow_integration_test.go`、`router_test.go`。

## Inputs

- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `.harness/reuse-decisions/backend-internal-app-full-router-integration-test-split.md`
- `docs/plan/impl-plan/2026-06-03-backend-internal-app-full-router-state-matrix-test-split-plan.md`

## Current problem

- `full_router_integration_test.go` 当前有 1445 行。
- 顶层测试同时覆盖 access control、smoke matrix、teacher authoring、admin AWD control、publish request lifecycle 和 router build 多类责任。
- 其中共享 env / helper 已经相对稳定，真正的结构问题在于顶层 `Test...` 和 helper owner 混在一个文件里。

## Working design

### Target structure

- 原文件保留：
  - `TestRouterBuildUsesCompositionModules`
  - `isAcceptableSmokeStatus` 及其以下所有 route / env / seed helper
  - `fullRouterTestEnv` 结构体
- 新增 3 个按领域归类的测试文件：
  - `full_router_access_integration_test.go`
  - `full_router_admin_integration_test.go`
  - `full_router_teacher_authoring_integration_test.go`

### Boundary choice

- 不新增新的 helper owner。
- 只移动顶层测试函数，不修改测试逻辑。
- 保持 `package app`，避免引入新的测试装配成本。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-internal-app-full-router-integration-test-split`

### Slice 2：按领域拆分顶层测试

- Goal：把 `full_router_integration_test.go` 的顶层测试按 access / admin / teacher-authoring 分到独立文件，原文件只保留 build test 与共享 helper。
- Touched files：
  - `code/backend/internal/app/full_router_integration_test.go`
  - `code/backend/internal/app/full_router_access_integration_test.go`
  - `code/backend/internal/app/full_router_admin_integration_test.go`
  - `code/backend/internal/app/full_router_teacher_authoring_integration_test.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestFullRouter_AccessControlMatrix|TestFullRouter_AuthorizedSmokeMatrix|TestFullRouter_ListInstancesMatchesContract|TestFullRouter_AdminCanToggleAWDControlsAndSeeOrchestrationState|TestFullRouter_TeacherCanBrowseArchivedAndDraftChallengesButOnlyManageOwnChallenges|TestFullRouter_CreateChallengeStoresCreator|TestFullRouter_ChallengeSelfCheckRunsPrecheckAndRuntime|TestFullRouter_AdminChallengePublishRequestLifecycle|TestRouterBuildUsesCompositionModules' -count=1`

### Slice 3：一致性收口

- Goal：确认 harness 文档和测试拆分后的文件组织没有引入新的漂移。
- Validation：
  - `bash scripts/check-consistency.sh`

## Expected change surface

- `internal/app` 集成测试文件组织
- harness reuse decision 与 implementation plan

## Data / API / compatibility impact

- 无生产数据、API、权限或业务语义变更。
- 风险主要在：
  - 复制顶层测试时漏 import
  - 原文件删除区间不准导致测试重复定义或缺失
  - helper 所需 import 被误删

## Review fit check

- Owner 清晰：原文件收口为 env / helper owner，新文件各自只承载单一测试领域。
- Reuse 清晰：继续复用现有 full router test env、route helper 和 seeded fixtures，不新造抽象。
- 结构收敛：既然本轮已经触碰 `full_router_integration_test.go` 这个 oversized surface，就直接把顶层测试 owner 拆开，不把同一位置的结构债继续留到 follow-up。

## Rollback / recovery

- 纯测试代码与文档组织调整，可直接回退到拆分前的单文件版本。
