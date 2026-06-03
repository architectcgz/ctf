## Review Target

- Repository: `ctf`
- Branch: `main`
- Diff source: working tree
- Files reviewed:
  - `code/backend/internal/app/router_routes.go`
  - `code/backend/internal/app/router_admin_contest_routes.go`
  - `code/backend/internal/app/router_admin_contest_routes_test.go`
  - `docs/plan/impl-plan/2026-06-03-backend-router-admin-contest-awd-split-plan.md`
  - `.harness/reuse-decisions/backend-router-admin-contest-awd-split.md`

## Classification Check

- Agree with non-trivial classification.
- Reason: this slice changes backend route registration structure, test coverage, and harness artifacts on an already oversized owner surface.

## Gate Verdict

- Conditional pass for implementation self-check.
- Independent review gate not met.

## Findings

- No material correctness findings in the current diff.

## Material Findings

- None.

## Senior Implementation Assessment

- `router.go` 继续只做 composition root，本轮没有把 route owner 又扩散到 module runtime，符合最小切片目标。
- admin contest / AWD 路由已从主 `router_routes.go` 移到独立 registrar 文件，并且通过 `adminContestRouteDeps` 收窄了依赖面。
- 新 registrar 使用 `Group("/:id")`、`Group("/awd")`、`Group("/:sid")`、`Group("/:rid")` 收口重复参数解析，减少了后续在同一 surface 上继续复制中间件的风险。
- `registerAdminRoutes` 仍保留 admin ops / identity / session revoke，这说明本轮只解决了 touched surface 中最拥挤的一段，没有虚构“整个 router 已经完成治理”。

## Required Re-validation

- `cd code/backend && go test ./internal/app -run TestAdminContestRoutesAreExtractedIntoDedicatedRegistrarFile -count=1`
- `cd code/backend && go test ./internal/app -run 'TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes|TestAdminSessionRoutes' -count=1`
- `cd code/backend && go test ./internal/app -run TestFullRouter_AccessControlMatrix -count=1`
- `bash scripts/check-consistency.sh`

## Residual Risk

- `registerAdminRoutes` 里剩余的 admin identity / ops 仍在原文件，本轮没有继续拆；这不是 blocker，但后续继续在同一函数叠加变更时仍会遇到相同结构压力。
- `TestFullRouter_AuthorizedSmokeMatrix` 当前会因为既有 runtime / practice 集成环境问题在 `POST /api/v1/contests/:id/challenges/:cid/instances` 上返回 500，这次没有尝试在 router 拆分切片里一并修复。

## Touched Known-Debt Status

- 本轮 touched debt 是 `registerAdminRoutes` 内 contest / AWD 路由 owner 过宽、重复参数解析和注册逻辑堆积。
- 该债务已在本轮 touched surface 内收口到独立 registrar，未继续作为 residual risk 留在相同位置。

## Independent Review Gate Status

- 当前会话的工具策略只允许在“用户明确要求 delegation / sub-agents”时启用独立 reviewer。
- 本次用户没有显式要求 delegation，因此无法满足 `development-pipeline` 要求的独立 reviewer gate。
- 本文档记录的是同上下文 self-review，只能作为实现自检证据，不能替代独立 review。

## Workflow Completion Check

- 未发现仓库内额外的 workflow completion script。
