## Review Target
- Repository: `ctf`
- Worktree: `/home/azhi/workspace/projects/ctf/.worktrees/feat/awd-review-pagination`
- Branch: `feat/awd-review-pagination`
- Diff source: working tree against current branch base
- Files reviewed:
  - `code/backend/internal/dto/contest.go`
  - `code/backend/internal/module/contest/api/http/contest_query_handler.go`
  - `code/backend/internal/module/contest/api/http/handler.go`
  - `code/backend/internal/module/contest/application/queries/contest_list_query.go`
  - `code/backend/internal/module/contest/application/queries/contest_result.go`
  - `code/backend/internal/module/contest/infrastructure/contest_repository.go`
  - `code/backend/internal/module/contest/ports/contest.go`
  - `code/backend/internal/app/full_router_state_matrix_integration_test.go`
  - `code/frontend/src/api/contracts.ts`
  - `code/frontend/src/api/contest.ts`
  - `code/frontend/src/api/admin/contests.ts`
  - `code/frontend/src/composables/usePagination.ts`
  - `code/frontend/src/features/contest-detail/model/useContestListPage.ts`
  - `code/frontend/src/features/scoreboard/model/useScoreboardView.ts`
  - `code/frontend/src/features/scoreboard/model/useScoreboardContestDirectoryPage.ts`
  - `code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts`
  - `code/frontend/src/features/platform-contests/model/useContestOperationsHubPage.ts`
  - `code/frontend/src/views/contests/ContestList.vue`
  - `code/frontend/src/views/scoreboard/ScoreboardView.vue`
  - `code/frontend/src/views/scoreboard/ScoreboardDetail.vue`
  - `code/frontend/src/views/platform/ContestOperationsHub.vue`
  - `code/frontend/src/components/platform/contest/ContestOperationsHubWorkspacePanel.vue`
  - related contest / scoreboard / contest-ops tests

## Classification Check
- Agree with non-trivial classification.
- Reason: this slice changes backend pagination/filter/sort contract, adds response summary fields, replaces frontend local pagination with server pagination across multiple route owners, and adjusts shared pagination state semantics.

## Gate Verdict
- Pass.

## Findings
- No未解决的 material correctness findings。
- 最后一轮修复补上了管理员竞赛目录页的 summary owner：`ContestManage` 不再用当前页 `list.filter(...).length` 伪造“报名中 / 进行中”总量，而是通过 `useContestListState` 把后端 `summary` 作为正式分页响应语义往上游传递，再在 `ContestOrchestrationPage` 消费。这样管理员端与学生端在分页后的状态指标语义保持一致。

## Material Findings
- None.

## Senior Implementation Assessment
- The chosen direction fixes the right boundary: contest list queries now own `statuses` / `mode` / `sort` / summary counts in the backend, instead of leaving `/scoreboard` and `contest-ops` on `page_size=100` plus local slicing.
- `/contests` now requests only student-visible statuses from the backend, so displayed rows, total count, and page controls stay aligned with the real contract instead of hiding draft rows after fetch.
- `/scoreboard/:contestId` keeps page ownership local to the route owner and refreshes the current page on manual refresh and websocket updates, which closes the previous hard-coded first-page behavior.

## Required Re-validation
- `go test ./internal/module/contest/application/queries ./internal/module/contest/api/http ./internal/app -run 'TestContestService|TestUpdateContestSkipsReadinessAuditPayloadWhenCommandFailsBeforeGate|TestFullRouter_AdminContestListSupportsModeStatusesSortAndSummary'`
- `pnpm typecheck`
- `pnpm test:run src/composables/__tests__/usePagination.test.ts src/views/contests/__tests__/ContestList.test.ts src/views/scoreboard/__tests__/ScoreboardView.test.ts src/views/platform/__tests__/ContestOperationsHub.test.ts`
- `pnpm test:run src/api/__tests__/contest.test.ts src/api/__tests__/admin.test.ts`
- `pnpm test:run src/views/platform/__tests__/ContestManage.test.ts src/features/platform-contests/model/platformContestsModelBoundary.test.ts`

## Residual Risk
- `ContestPageResp.summary` is additive and currently only consumed by the new paginated owners. Older callers that ignore it remain compatible, but any future list owner that needs cross-status hero metrics must explicitly use the summary instead of recomputing from the current page.
- `ScoreboardDetail` still does not have a backend-provided global scoreboard aggregate, so the page now labels `topScore` and `solvedCount` as current-page metrics to stay honest under pagination.
- 管理员竞赛目录页仍有一个诚实性约束：`AWD 模式` 指标目前只能根据当前页列表计算，因为后端 summary 还没有提供 mode 维度聚合；当前实现已把卡片文案改成“当前页已接入运维链路的赛事”，避免把当前页计数冒充全量总数。

## Touched Known-Debt Status
- The touched surfaces previously carried the exact debt called out by the user:
  - `/contests` had backend pagination but no UI paging
  - `/scoreboard` relied on `page=1&page_size=100` and local slicing
  - `/scoreboard/:contestId` hard-coded `page=1&page_size=100`
  - `/platform/contest-ops/contests` relied on `page=1&page_size=100` and current-page-derived hero counts
- This slice closes that touched debt by moving all four surfaces onto real server pagination with default `20`, removing local contest-directory slicing, and adding backend summary support where the page-level metrics needed whole-result counts.

## Workflow Completion Check
- No repository-local workflow completion script matching `check-workflow-complete.sh` or equivalent workflow-complete pattern was found during repo scan, so completion evidence is based on the archived review plus executed verification commands above.
