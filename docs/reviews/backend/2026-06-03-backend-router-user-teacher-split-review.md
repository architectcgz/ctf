## Review Target

- Repository: `ctf`
- Branch: `main`
- Diff source: working tree
- Files reviewed:
  - `code/backend/internal/app/router_routes.go`
  - `code/backend/internal/app/router_user_self_routes.go`
  - `code/backend/internal/app/router_user_teacher_routes.go`
  - `code/backend/internal/app/router_user_teacher_routes_test.go`
  - `docs/plan/impl-plan/2026-06-03-backend-router-user-teacher-split-plan.md`
  - `.harness/reuse-decisions/backend-router-user-teacher-split.md`

## Classification Check

- Agree with non-trivial classification.
- Reason: this slice closes the remaining oversized user/teacher router surface and touches route registration, access control grouping, tests, and harness evidence.

## Gate Verdict

- Conditional pass for implementation self-check.
- Independent review gate not met.

## Findings

- No material correctness findings in the current diff.

## Material Findings

- None.

## Senior Implementation Assessment

- `registerUserRoutes` 已退化为两个委托：`registerUserSelfRoutes` 与 `registerTeacherRoutes`，原来的 oversized owner surface 已经收口。
- `user self` 文件负责 public contest browse、protected gameplay、instances、`users/me/*` 和 personal report；这个分组和运行时权限边界一致。
- `teacher` 文件承接 `teacherOrAbove` 下的教师工作区，以及 `protected /users/:id/skill-profile` 这种 teacher-protected 但不在 `/teacher/*` 前缀下的路由，避免了单纯按路径前缀切分导致的 owner 失真。
- 这轮没有继续把 teacher route 里的匿名业务逻辑下沉，因为当前 teacher surface 基本都是直接调 handler，已经比此前的总入口函数清楚得多。

## Required Re-validation

- `cd code/backend && go test ./internal/app -run 'TestUserSelfRoutesAreExtractedIntoDedicatedRegistrarFile|TestTeacherRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`
- `cd code/backend && go test ./internal/app -run 'TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes|TestFullRouter_AccessControlMatrix|TestUserSelfRoutesAreExtractedIntoDedicatedRegistrarFile|TestTeacherRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`
- `bash scripts/check-consistency.sh`

## Residual Risk

- `registerTeacherAuthoringRoutes` 仍是下一块明显的大 surface；如果继续做 router 治理，下一刀应转向 authoring，而不是再回头修改已收口的 admin/user 入口。
- `TestFullRouter_AuthorizedSmokeMatrix` 的既有 runtime/practice 环境失败没有在本轮处理，本轮也没有用它作为完成门槛。

## Touched Known-Debt Status

- 本轮 touched debt 是 `registerUserRoutes` 里学生、教师、自助画像和报表出口混在同一函数。
- 该 debt 已在本轮 touched surface 内收口到 `user self` 与 `teacher` 两个 registrar 文件。

## Independent Review Gate Status

- 当前会话的工具策略只允许在“用户明确要求 delegation / sub-agents”时启用独立 reviewer。
- 本次用户没有显式要求 delegation，因此本文档仍是同上下文 self-review，不能替代独立 review。

## Workflow Completion Check

- 未发现仓库内额外的 workflow completion script。
