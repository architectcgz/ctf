## Review Target

- Repository: `ctf`
- Branch: `main`
- Diff source: working tree
- Files reviewed:
  - `code/backend/internal/app/router_routes.go`
  - `code/backend/internal/app/router_authoring_challenge_routes.go`
  - `code/backend/internal/app/router_authoring_asset_routes.go`
  - `code/backend/internal/app/router_authoring_awd_routes.go`
  - `code/backend/internal/app/router_authoring_routes_test.go`
  - `docs/plan/impl-plan/2026-06-03-backend-router-authoring-split-plan.md`
  - `.harness/reuse-decisions/backend-router-authoring-split.md`

## Classification Check

- Agree with non-trivial classification.
- Reason: this slice closes the remaining oversized authoring router surface and touches route registration, owner guard grouping, tests, and harness evidence.

## Gate Verdict

- Conditional pass for implementation self-check.
- Independent review gate not met.

## Findings

- No material correctness findings in the current diff.

## Material Findings

- None.

## Senior Implementation Assessment

- `registerTeacherAuthoringRoutes` 已退化为三个委托：`registerAuthoringChallengeRoutes`、`registerAuthoringAssetRoutes`、`registerAuthoringAWDRoutes`，此前混在同一个函数里的普通题 authoring、共享资源资产、AWD authoring 已经分开。
- `challengeOwnerGuard` 仍只作用在普通题 authoring 的 owner-sensitive 路由上，没有被错误带进 image / template / AWD surface。
- image 与 environment template 被放到同一个 asset registrar，语义上都属于 authoring 共享资源，后续增量修改不需要再回到 challenge CRUD 文件。
- AWD challenge 与 AWD import 被单独收口到 `router_authoring_awd_routes.go`，保持它和普通题 import / CRUD 的契约边界分离。

## Required Re-validation

- `cd code/backend && go test ./internal/app -run 'TestAuthoringChallengeRoutesAreExtractedIntoDedicatedRegistrarFile|TestAuthoringAssetRoutesAreExtractedIntoDedicatedRegistrarFile|TestAuthoringAWDRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`
- `cd code/backend && go test ./internal/app -run 'TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes|TestFullRouter_AccessControlMatrix|TestFullRouter_AdminChallengeManagementStateMatrix|TestFullRouter_AWDChallengeAuthoringStateMatrix|TestAuthoringChallengeRoutesAreExtractedIntoDedicatedRegistrarFile|TestAuthoringAssetRoutesAreExtractedIntoDedicatedRegistrarFile|TestAuthoringAWDRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`
- `bash scripts/check-consistency.sh`

## Residual Risk

- 当前会话只做了同上下文 self-review；由于用户没有显式要求 delegation / sub-agents，独立 reviewer gate 仍未满足。
- `go test ./internal/app -count=1` 和 `TestFullRouter_AuthorizedSmokeMatrix` 的既有 runtime / practice 环境失败没有在本轮处理，本轮也没有用它们作为完成门槛。

## Touched Known-Debt Status

- 本轮 touched debt 是 `registerTeacherAuthoringRoutes` 把普通题 authoring、资源资产和 AWD authoring 混在同一个大 registrar。
- 该 debt 已在本轮 touched surface 内收口到 challenge / asset / awd 三个 registrar 文件。

## Independent Review Gate Status

- 当前会话的工具策略只允许在“用户明确要求 delegation / sub-agents”时启用独立 reviewer。
- 本次用户没有显式要求 delegation，因此本文档仍是同上下文 self-review，不能替代独立 review。

## Workflow Completion Check

- 未发现仓库内额外的 workflow completion script。
