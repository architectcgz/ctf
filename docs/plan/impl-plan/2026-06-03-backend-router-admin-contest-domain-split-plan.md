# 2026-06-03 backend router admin contest domain split plan

## Objective

- 把 `code/backend/internal/app/router_admin_contest_routes.go` 继续从单文件大 registrar 拆成稳定的 admin contest 子域 registrar。
- 让 contest core、challenge roster、participation、AWD 各自落到独立文件，收口当前 admin contest surface 的 owner 混杂。
- 保持现有 URL、权限、中间件顺序、handler owner 和测试语义不变。

## Non-goals

- 不在本轮修改 `registerAdminRoutes`、`registerUserRoutes`、`registerTeacherRoutes` 或 authoring registrars。
- 不在本轮调整 admin contest / AWD 的权限模型、audit 语义或 handler 实现。
- 不在本轮修复既有 runtime / practice 集成测试里的环境性失败。

## Inputs

- `code/backend/internal/app/router_admin_contest_routes.go`
- `code/backend/internal/app/router_admin_contest_routes_test.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `docs/reviews/backend/2026-06-03-backend-router-admin-contest-awd-split-review.md`

## Current problem

- `router_admin_contest_routes.go` 当前仍同时承载：
  - contest create/list/get/update/freeze/unfreeze/export
  - contest challenge roster
  - registration / announcement / scoreboard 这类 participation ops
  - AWD 全部管理链路
- 虽然它已经从主 `router_routes.go` 移出，但从 owner 看，这几块并不是同一类行为，继续让它们堆在一个 registrar 文件里，后续 admin contest 改动仍会回到一把大文件。

## Working design

### Target structure

- 保留 `registerAdminContestRoutes(adminOnly, deps)` 作为 admin contest 总入口。
- 在 `internal/app` 内新增：
  - `router_admin_contest_core_routes.go`
  - `router_admin_contest_challenge_routes.go`
  - `router_admin_contest_participation_routes.go`
  - `router_admin_contest_awd_routes.go`
- `router_admin_contest_routes.go` 只保留：
  - `adminContestRouteDeps`
  - `registerAdminContestRoutes`
  - 组装 `contestByID`
  - 委托到四个子 registrar

### Boundary choice

- `contest core` 文件负责：
  - contest create/list/get/update
  - freeze / unfreeze
  - export
- `contest challenge` 文件负责：
  - `GET/POST/PUT/DELETE /contests/:id/challenges`
- `contest participation` 文件负责：
  - registrations review
  - announcements
  - live scoreboard
- `contest awd` 文件负责：
  - 现有 `registerAdminContestAWDRoutes` 全部 AWD 路由

### Grill check

- 不把 challenge roster 并进 core，因为 contest lifecycle 和 challenge roster 的改动频率、handler owner、review 关注点已经不同。
- 不把 scoreboard/live 放进 core；它更接近 participation / run-time ops，而不是 contest definition。
- AWD 继续整块独立，不再和普通 contest 管理落在同一文件。

## Task slices

### Slice 1：计划与 reuse-first 门禁

- Goal：补 implementation plan 与 reuse decision。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-router-admin-contest-domain-split`

### Slice 2：先补结构测试

- Goal：用结构测试约束 admin contest core / challenge / participation / awd 必须迁到独立 registrar 文件。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestAdminContestCoreRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminContestChallengeRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminContestParticipationRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminContestAWDRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`

### Slice 3：拆出 admin contest 子 registrars

- Goal：把 `router_admin_contest_routes.go` 退化为 delegator，并把 core / challenge / participation / awd 落到独立文件。
- Touched files：
  - `code/backend/internal/app/router_admin_contest_routes.go`
  - `code/backend/internal/app/router_admin_contest_core_routes.go`
  - `code/backend/internal/app/router_admin_contest_challenge_routes.go`
  - `code/backend/internal/app/router_admin_contest_participation_routes.go`
  - `code/backend/internal/app/router_admin_contest_awd_routes.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestAdminContestCoreRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminContestChallengeRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminContestParticipationRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminContestAWDRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`

### Slice 4：最小充分回归

- Goal：确认 admin contest / AWD 路径存在性、practice wiring 和访问矩阵不漂移。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes|TestFullRouter_AccessControlMatrix|TestAdminContestCoreRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminContestChallengeRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminContestParticipationRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminContestAWDRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`
  - `bash scripts/check-consistency.sh`

### Slice 5：review 归档

- Goal：记录本轮 self-review 与独立 reviewer gate 状态。
- Touched files：
  - `docs/reviews/backend/2026-06-03-backend-router-admin-contest-domain-split-review.md`

## Expected change surface

- `internal/app` admin contest registrar 继续拆分
- 局部 deps struct 复用
- 结构测试
- review 记录

## Data / API / compatibility impact

- 无数据结构变更。
- 无 API 路径、method、权限变更。
- 风险主要在：
  - `contestByID.Use(ParseInt64Param("id"))` 的 group 作用域漂移
  - challenge / participation 路由误落到 core 或 AWD 文件
  - admin AWD orchestration 路由遗漏或 audit 顺序漂移

## Review fit check

- Owner 清晰：本轮完成后 admin contest 总入口不再承载四类子域的具体注册细节。
- Reuse 清晰：沿用前几刀已经建立的 registrar 拆分模式。
- 结构收敛：一旦触碰 `router_admin_contest_routes.go`，本轮直接完成该 surface 的 owner 收口，而不是只把 AWD 或 scoreboard 再单独挪一次。

## Rollback / recovery

- 纯代码组织调整，可直接回退新增 registrar 文件与分发调用。
