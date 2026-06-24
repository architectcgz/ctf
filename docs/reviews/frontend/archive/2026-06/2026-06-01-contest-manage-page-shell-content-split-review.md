# Contest Manage Page Shell Content Split Review

## Review Target

- Repository: `ctf`
- Branch: `main`
- Diff source: working tree changes for `contest-manage-page-shell-content-split`
- Files reviewed:
  - `.harness/reuse-decisions/contest-manage-page-shell-content-split.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-01-contest-manage-page-shell-content-split-plan.md`
  - `docs/reviews/frontend/2026-06-01-contest-manage-page-shell-content-split-review.md`
  - `code/frontend/src/features/platform/contests/ui/ContestOrchestrationPage.vue`
  - `code/frontend/src/features/platform/contests/ui/ContestManageOverviewPanel.vue`
  - `code/frontend/src/features/platform/contests/ui/ContestManageCreatePanel.vue`
  - `code/frontend/src/features/platform/contests/ui/contestOrchestrationPage.css`
  - `code/frontend/src/features/platform/contests/ui/contestOrchestrationPage.types.ts`
  - `code/frontend/src/features/platform/contests/ui/contestAdminUiStrategy.test.ts`
  - `code/frontend/src/pages/platform/contests/__tests__/ContestManage.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Classification Check

- 结论：同意 `non-trivial frontend refactor` 分类。
- 原因：这次涉及 page shell 职责重排、两个内部 panel owner、新增页面级 css owner 和相邻 source-boundary 护栏更新。

## Gate Verdict

- `pass with minor issues`
- 说明：当前归档的是显式自审结果；由于本回合没有用户授权 delegation，未执行独立 reviewer gate。

## Findings

- 无 blocker / major / minor finding。

## Material Findings

- 无。

## Senior Implementation Assessment

- 当前切口比回去继续动 `ContestManageRoutePage.vue` 更合适：route page 虽然还在组合 page model 与附属 dialog，但真正过宽的是 `ContestOrchestrationPage.vue` 同时承接 overview/create 两块内容装配。
- 把 overview/create panel 拆开，同时保留 `ContestOrchestrationPage.vue` 作为 shell / active panel / 事件桥接 owner，能直接压缩 `platform/contests` 当前最明显的 oversized feature surface，而不会把 route-level owner 和 feature 内部 page shell 两条线混在一刀里。

## Required Re-validation

- `bash scripts/check-task-intake.sh --reuse-decision contest-manage-page-shell-content-split`
- `cd code/frontend && npm run test:run -- src/pages/platform/contests/__tests__/ContestManage.test.ts src/features/platform/contests/ui/contestAdminUiStrategy.test.ts`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-manage-page-shell-content-split.md docs/plan/archive/impl-plan/2026-06/2026-06-01-contest-manage-page-shell-content-split-plan.md docs/reviews/frontend/2026-06-01-contest-manage-page-shell-content-split-review.md code/frontend/src/features/platform/contests/ui/ContestOrchestrationPage.vue code/frontend/src/features/platform/contests/ui/ContestManageOverviewPanel.vue code/frontend/src/features/platform/contests/ui/ContestManageCreatePanel.vue code/frontend/src/features/platform/contests/ui/contestOrchestrationPage.css code/frontend/src/features/platform/contests/ui/contestOrchestrationPage.types.ts code/frontend/src/features/platform/contests/ui/contestAdminUiStrategy.test.ts code/frontend/src/pages/platform/contests/__tests__/ContestManage.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual Risk

- 这轮只收 `ContestOrchestrationPage.vue` 的内部 content assembly，不继续改 route page owner，也不动 create/edit dialog 和 announcement drawer；如果后续继续压 `platform/contests` 的 route-level owner，再单开一刀对齐 `StudentAnalysisWorkspacePage.vue` 这一类 feature page 组合模式更合适。
- 独立 reviewer gate 仍未执行；如果需要严格满足 pipeline，这一条需要在用户明确授权 delegation 后补上。

## Touched Known-debt Status

- 已触达并收口已知结构债：`ContestOrchestrationPage.vue` 这个 backlog 中已记录的 oversized feature owner surface，本轮已从单一 page-sized SFC 收口成 shell + 两个内部 panel owner，不再维持原状。
