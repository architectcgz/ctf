# 2026-06-03 backend internal app router test split plan

## Objective

- 把 `code/backend/internal/app/router_test.go` 中混杂的路由行为测试与 composition 守卫测试拆成多个按职责聚焦的测试文件。
- 保留原文件作为 `internal/app` router test shared helper owner。
- 继续降低 `internal/app` 中 1000+ 行测试文件的 owner 混杂和 review 成本。

## Non-goals

- 本轮不改 `router.go`、`router_routes.go` 或 `composition/*` 的生产逻辑。
- 本轮不重写 `newAppTestDependencies`、TLS 证书 helper、反射断言 helper。
- 本轮不处理已知的 `full_router` 既有失败。

## Inputs

- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/router_session_routes_test.go`
- `code/backend/internal/app/router_admin_contest_routes_test.go`
- `.harness/reuse-decisions/backend-internal-app-router-test-split.md`

## Current problem

- `router_test.go` 当前有 1302 行。
- 文件里同时承载：
  - 路由注册 / handler 归属 / runtime dial failure / owner guard 行为测试
  - assessment AWD query param / 错误消息守卫
  - composition module struct、typed deps、源码 marker、跨模块依赖守卫
- 这些测试虽然都围绕 `internal/app`，但 owner 和回归信号已经明显不同，继续堆在一个文件里会拉高 review 和定位成本。

## Working design

### Target structure

- 原文件保留：
  - `newAppTestDependencies`
  - `assertHasRoute`、`assertRouteHandlerContains`、`assertRouteMissing`
  - `assertFieldType`、`assertNoField`、`assertFunctionParamType`
  - TLS helper
- 新增 3 个按职责归类的测试文件：
  - `router_route_wiring_test.go`
  - `router_composition_structure_test.go`
  - `router_composition_typed_deps_test.go`

### Boundary choice

- `router_route_wiring_test.go` 负责真实 `NewRouter` 行为、guard、query param 和 dial failure 这组更接近外部行为的测试。
- `router_composition_structure_test.go` 负责 `BuildRoot`、module contract compile、公开字段与 builder/source marker 守卫。
- `router_composition_typed_deps_test.go` 负责 runtime/auth/ops/contest/challenge/practice/assessment/identity/teaching query 的 typed deps 与 cross-module guard。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-internal-app-router-test-split`

### Slice 2：拆分 router 行为与 composition 守卫测试

- Goal：把 `router_test.go` 的顶层测试按职责迁到 3 个新文件，原文件只保留 shared helper owner。
- Touched files：
  - `code/backend/internal/app/router_test.go`
  - `code/backend/internal/app/router_route_wiring_test.go`
  - `code/backend/internal/app/router_composition_structure_test.go`
  - `code/backend/internal/app/router_composition_typed_deps_test.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestNewRouter|TestChallengeOwnerGuard|TestTeacherAWDReview|TestBuildRoot|TestComposition|TestRuntimeModule|TestAuthModule|TestOpsModule|TestBuildContestModule|TestContest|TestChallengeModule|TestBuildChallengeModule|TestPracticeModule|TestBuildPracticeModule|TestAssessmentModule|TestBuildAssessmentModule|TestIdentityModule|TestTeachingQueryModule|TestRouterRateLimitStrategy' -count=1`

### Slice 3：一致性收口

- Goal：确认 harness 文档与测试文件组织没有引入新的漂移。
- Validation：
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- `internal/app` router 测试文件组织
- router/composition 测试 owner 划分
- harness reuse decision 与 implementation plan

## Data / API / compatibility impact

- 无生产数据、API 或业务逻辑变更。
- 风险主要在：
  - 拆分时遗漏原有断言
  - 测试迁移后缺少必要 import 或 helper owner 不清晰
  - 把 helper 也一起搬散，导致后续 router 测试继续复制

## Review fit check

- Owner 清晰：原文件收口为 shared helper owner，新文件分别承载 route wiring、composition structure、typed deps 三个面。
- Reuse 清晰：继续复用现有 router test helper，不新造额外 testkit。
- 结构收敛：既然本轮已经触碰 `router_test.go` 这个 oversized surface，就直接把职责混杂拆开，而不是只搬走一小部分测试继续保留同一问题。

## Rollback / recovery

- 纯测试与文档组织调整，可直接回退到拆分前的单文件版本。
