## Review Target

- Repository: `ctf`
- Branch: `main`
- Diff source: working tree
- Files reviewed:
  - `code/backend/internal/app/router_user_self_routes.go`
  - `code/backend/internal/app/router_user_contest_routes.go`
  - `code/backend/internal/app/router_user_practice_routes.go`
  - `code/backend/internal/app/router_user_self_service_routes.go`
  - `code/backend/internal/app/router_user_self_domain_routes_test.go`
  - `docs/plan/impl-plan/2026-06-03-backend-router-user-self-domain-split-plan.md`
  - `.harness/reuse-decisions/backend-router-user-self-domain-split.md`

## Classification Check

- Agree with non-trivial classification.
- Reason: this slice closes the remaining oversized user self router surface and touches route registration, runtime handler grouping, tests, and harness evidence.

## Gate Verdict

- Conditional pass for implementation self-check.
- Independent review gate not met.

## Findings

- No material correctness findings in the current diff.

## Material Findings

- None.

## Senior Implementation Assessment

- `registerUserSelfRoutes` 已退化为三个委托：`registerUserContestRoutes`、`registerUserPracticeRoutes`、`registerUserSelfServiceRoutes`，此前混在一个文件里的 contest participation、practice/runtime、自助画像与报表已经分开。
- contest 内的 challenge / AWD instance、proxy、target access 继续放在 `user contest` owner 下，没有被底层 practice / runtime handler 所属 module 打散。
- `user practice` 只承接公开题练习、全局实例生命周期和排行榜，和 contest participation surface 的路径、语义已经分离。
- `/reports/class` 的 protected + teacher guard 版本仍保留在 self service 文件里，没有因为 teacher role 约束被错误并回 teacher registrar。

## Required Re-validation

- `cd code/backend && go test ./internal/app -run 'TestUserContestRoutesAreExtractedIntoDedicatedRegistrarFile|TestUserPracticeRoutesAreExtractedIntoDedicatedRegistrarFile|TestUserSelfServiceRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`
- `cd code/backend && go test ./internal/app -run 'TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes|TestFullRouter_AccessControlMatrix|TestUserContestRoutesAreExtractedIntoDedicatedRegistrarFile|TestUserPracticeRoutesAreExtractedIntoDedicatedRegistrarFile|TestUserSelfServiceRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`
- `bash scripts/check-consistency.sh`

## Residual Risk

- 当前会话只做了同上下文 self-review；由于用户没有显式要求 delegation / sub-agents，独立 reviewer gate 仍未满足。
- `go test ./internal/app -count=1` 和 `TestFullRouter_AuthorizedSmokeMatrix` 的既有 runtime / practice 环境失败没有在本轮处理，本轮也没有用它们作为完成门槛。

## Touched Known-Debt Status

- 本轮 touched debt 是 `router_user_self_routes.go` 同时承载 contest participation、practice/runtime、自助画像与报表出口。
- 该 debt 已在本轮 touched surface 内收口到 contest / practice / self service 三个 registrar 文件。

## Independent Review Gate Status

- 当前会话的工具策略只允许在“用户明确要求 delegation / sub-agents”时启用独立 reviewer。
- 本次用户没有显式要求 delegation，因此本文档仍是同上下文 self-review，不能替代独立 review。

## Workflow Completion Check

- 未发现仓库内额外的 workflow completion script。
