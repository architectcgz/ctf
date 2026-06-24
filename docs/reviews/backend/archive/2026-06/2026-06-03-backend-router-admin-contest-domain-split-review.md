## Review Target

- Repository: `ctf`
- Branch: `main`
- Diff source: working tree
- Files reviewed:
  - `code/backend/internal/app/router_admin_contest_routes.go`
  - `code/backend/internal/app/router_admin_contest_core_routes.go`
  - `code/backend/internal/app/router_admin_contest_challenge_routes.go`
  - `code/backend/internal/app/router_admin_contest_participation_routes.go`
  - `code/backend/internal/app/router_admin_contest_awd_routes.go`
  - `code/backend/internal/app/router_admin_contest_routes_test.go`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-03-backend-router-admin-contest-domain-split-plan.md`
  - `.harness/reuse-decisions/backend-router-admin-contest-domain-split.md`

## Classification Check

- Agree with non-trivial classification.
- Reason: this slice closes the remaining oversized admin contest router surface and touches route registration, practice wiring, tests, and harness evidence.

## Gate Verdict

- Conditional pass for implementation self-check.
- Independent review gate not met.

## Findings

- No material correctness findings in the current diff.

## Material Findings

- None.

## Senior Implementation Assessment

- `router_admin_contest_routes.go` 已退化为 delegator，contest core、challenge roster、participation、AWD 四类 owner 不再混在一个文件里。
- `contestByID` 仍在总入口组装并统一挂 `ParseInt64Param("id")`，没有因为拆分把相同 middleware 重复散落回多个文件。
- AWD 路由完整迁到 `router_admin_contest_awd_routes.go`，同时保留了 `awdReadinessAudit` 的现有使用位置，没有引入新的全局 helper 或 route DSL。
- challenge roster 与 participation ops 分开后，后续改动 contest challenge 编排或 registration / announcement 逻辑时，不需要再回到 AWD 或 contest core 文件。

## Required Re-validation

- `cd code/backend && go test ./internal/app -run 'TestAdminContestCoreRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminContestChallengeRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminContestParticipationRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminContestAWDRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminContestRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`
- `cd code/backend && go test ./internal/app -run 'TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes|TestFullRouter_AccessControlMatrix|TestAdminContestCoreRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminContestChallengeRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminContestParticipationRoutesAreExtractedIntoDedicatedRegistrarFile|TestAdminContestAWDRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`
- `bash scripts/check-consistency.sh`

## Residual Risk

- 当前会话只做了同上下文 self-review；由于用户没有显式要求 delegation / sub-agents，独立 reviewer gate 仍未满足。
- `go test ./internal/app -count=1` 和 `TestFullRouter_AuthorizedSmokeMatrix` 的既有 runtime / practice 环境失败没有在本轮处理，本轮也没有用它们作为完成门槛。

## Touched Known-Debt Status

- 本轮 touched debt 是 `router_admin_contest_routes.go` 同时承载 contest core、challenge roster、participation 和 AWD 四类 owner。
- 该 debt 已在本轮 touched surface 内收口到四个独立 registrar 文件。

## Independent Review Gate Status

- 当前会话的工具策略只允许在“用户明确要求 delegation / sub-agents”时启用独立 reviewer。
- 本次用户没有显式要求 delegation，因此本文档仍是同上下文 self-review，不能替代独立 review。

## Workflow Completion Check

- 未发现仓库内额外的 workflow completion script。
