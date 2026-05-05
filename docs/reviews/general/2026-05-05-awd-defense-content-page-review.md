# AWD Defense Content Page Review

- Review target:
  - Repository: `ctf`
  - Worktree: `/home/azhi/workspace/projects/ctf/.worktrees/feat/awd-defense-content-page`
  - Review date: `2026-05-05`
  - Scope:
    - backend read-only AWD defense file access
    - student defense content page route and page owner
    - defense button navigation from AWD battle workspace
  - Files reviewed:
    - `code/backend/internal/module/runtime/api/http/handler.go`
    - `code/backend/internal/module/runtime/api/http/handler_test.go`
    - `code/backend/internal/module/runtime/runtime/adapters.go`
    - `code/backend/internal/module/runtime/runtime/adapters_test.go`
    - `code/backend/internal/module/runtime/runtime/module.go`
    - `code/frontend/src/api/contest.ts`
    - `code/frontend/src/api/__tests__/contest.test.ts`
    - `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
    - `code/frontend/src/components/contests/awd/AWDDefenseFileWorkbench.vue`
    - `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
    - `code/frontend/src/components/contests/awd/__tests__/AWDDefenseFileWorkbench.test.ts`
    - `code/frontend/src/features/contest-awd-workspace/model/useContestAwdDefenseWorkbenchPage.ts`
    - `code/frontend/src/features/contest-awd-workspace/model/useContestAwdDefenseWorkbenchPage.test.ts`
    - `code/frontend/src/router/routes/studentRoutes.ts`
    - `code/frontend/src/views/contests/ContestAWDDefenseWorkbench.vue`
    - `code/frontend/src/views/contests/__tests__/ContestAWDDefenseWorkbench.test.ts`
    - `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`
    - `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
    - `docs/architecture/features/awd-defense-content-page-design.md`
    - `docs/plan/impl-plan/2026-05-05-awd-defense-content-page-implementation-plan.md`

- Validation executed:
  - Implementation validation:
    - `cd code/backend && go test ./internal/module/runtime/api/http ./internal/module/runtime/runtime -count=1`
    - `cd code/backend && go test ./internal/app/composition -run AWDDefenseWorkbench -count=1`
    - `cd code/frontend && npm run test:run -- src/api/__tests__/contest.test.ts src/features/contest-awd-workspace/model/useContestAwdDefenseWorkbenchPage.test.ts src/components/contests/awd/__tests__/AWDDefenseFileWorkbench.test.ts src/views/contests/__tests__/ContestAWDDefenseWorkbench.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/views/contests/__tests__/ContestDetail.test.ts src/components/layout/__tests__/AppLayout.test.ts`
    - `cd code/frontend && npm run typecheck`
  - Follow-up review validation:
    - `cd code/frontend && npm run test:run -- src/features/contest-awd-workspace/model/useContestAwdDefenseWorkbenchPage.test.ts src/components/contests/awd/__tests__/AWDDefenseFileWorkbench.test.ts`
    - `cd code/frontend && npm run typecheck`

- Review process:
  - Initial independent review blocked the change on two major frontend issues:
    - no parent/root navigation after entering a subdirectory
    - old file responses could write back after service/page switch
  - Fixes were applied:
    - `AWDDefenseFileWorkbench.vue` added `根目录 / 上一级` actions and regression test coverage
    - `useContestAwdDefenseWorkbenchPage.ts` now invalidates old `directoryRequestSeq` / `fileRequestSeq` at `loadPage()` start and added route-switch stale-response coverage
  - Follow-up independent review found no material findings

## Findings

No material findings blocking completion.

## Residual Risks

1. Review reran only the directly affected frontend regression tests plus `typecheck`; the broader frontend targeted suite was already run during implementation, but not re-run by the reviewer itself.
2. `根目录 / 上一级` buttons can still be clicked during initial page loading and may trigger extra read-only requests; current sequence guards and backend permission boundaries keep this low risk.
3. Backend read-only safety currently relies on rooted path resolution and sensitive path filtering. This review did not add a runtime end-to-end check for symbolic links or other special files inside the mounted root.
4. [ContestAWDWorkspacePanel.vue](/home/azhi/workspace/projects/ctf/.worktrees/feat/awd-defense-content-page/code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue) 仍是 `1163` 行的超大组件，且已经在 `docs/reviews/frontend/ctf-frontend-audit-20260422.md` 的 `TD-1` backlog 中作为待拆对象存在。本次切片已经把新的防守内容页 owner 抽到 [ContestAWDDefenseWorkbench.vue](/home/azhi/workspace/projects/ctf/.worktrees/feat/awd-defense-content-page/code/frontend/src/views/contests/ContestAWDDefenseWorkbench.vue) 和 [useContestAwdDefenseWorkbenchPage.ts](/home/azhi/workspace/projects/ctf/.worktrees/feat/awd-defense-content-page/code/frontend/src/features/contest-awd-workspace/model/useContestAwdDefenseWorkbenchPage.ts)，避免继续把路由和文件浏览状态堆回父组件，但这不等于父组件拆分债已经消失。更严格的 review 记录应显式保留这一条非阻塞结构风险。
