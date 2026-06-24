# Platform Contest Manage Page Owner Cleanup 计划

## Objective

- 把 `ContestManageRoutePage.vue` 收口成纯 route 薄壳。
- 在 `features/platform/contests/ui` 新增 `PlatformContestManagePage.vue`，统一承接 contest manage 的 page model、page shell 与附属 overlay owner。

## Non-goals

- 不改 `useContestManagePage.ts` 的 query 同步、创建/编辑保存、公告抽屉、AWD readiness override 或列表刷新逻辑。
- 不继续拆 `ContestOrchestrationPage.vue`、`ContestManageOverviewPanel.vue` 或 `ContestManageCreatePanel.vue`。
- 不顺手处理 `ContestEditRoutePage.vue`、`ContestOperationsHubRoutePage.vue` 或 `ContestAnnouncementsRoutePage.vue`。

## Source Inputs

- `code/frontend/src/pages/platform/contests/ContestManageRoutePage.vue`
- `code/frontend/src/features/platform/contests/model/useContestManagePage.ts`
- `code/frontend/src/features/platform/contests/ui/ContestOrchestrationPage.vue`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/features/platform/instance-management/ui/PlatformInstanceManagementPage.vue`
- `code/frontend/src/pages/platform/InstanceManageRoutePage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspacePage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestManage.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- 当前最小且能继续压 `platform/contests` route-level owner 的切口，是新增 feature 内部 `PlatformContestManagePage.vue`，而不是继续让 `ContestManageRoutePage.vue` 直接拿 page model 和 overlay。
- 这轮把 route page 和 feature page 的 owner 边界拉直后，后续如果继续清 `ContestEdit` / `ContestAnnouncements` / `ContestOperationsHub`，也能复用同一模式，不必每页都重新证明边界。

## Task Slices

### Slice 1: 新增 feature page owner

- 目标：在 `features/platform/contests/ui` 新增 `PlatformContestManagePage.vue`，承接 `useContestManagePage()`、`ContestOrchestrationPage` 和三组 overlay 组合。
- 风险：dialog / drawer 的 props 与事件桥接必须保持现有 contract，不引入新的中间状态 owner。

### Slice 2: 收口 route page 与 public API

- 目标：让 `ContestManageRoutePage.vue` 只渲染 `PlatformContestManagePage`，并通过 `features/platform/contests` public API 读取。
- 风险：如果 route page 仍深导入 feature internal module，公共出口与真实 owner 还是会继续漂移。

### Slice 3: 护栏与 backlog 同步

- 目标：更新 `ContestManage.test.ts` 的 source-boundary 断言和 backlog 进展，确保后续不会把 page model 或 overlay 组合写回 route page。
- 风险：如果只改实现不改断言，route page 很容易再次回到“薄壳 + 半个 page controller”混合状态。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision platform-contest-manage-page-owner-cleanup`
- `cd code/frontend && npm run test:run -- src/pages/platform/contests/__tests__/ContestManage.test.ts`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-contest-manage-page-owner-cleanup.md docs/plan/archive/impl-plan/2026-06/2026-06-01-platform-contest-manage-page-owner-cleanup-plan.md docs/reviews/frontend/2026-06-01-platform-contest-manage-page-owner-cleanup-review.md code/frontend/src/features/platform/contests/ui/PlatformContestManagePage.vue code/frontend/src/features/platform/contests/ui/index.ts code/frontend/src/pages/platform/contests/ContestManageRoutePage.vue code/frontend/src/pages/platform/contests/__tests__/ContestManage.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review Focus

- `PlatformContestManagePage.vue` 是否真的成为 contest manage 的单一 page-level owner，而不是继续把 page model / overlay owner 散在 route page。
- `ContestManageRoutePage.vue` 是否已经退回真正的 feature public API 薄壳。
- source-boundary 护栏是否足够防止 route page 再次深耦合 `useContestManagePage()` 或附属 overlay。
