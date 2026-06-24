# Platform Contest Operations Hub Page Owner Cleanup 计划

## Objective

- 把 `ContestOperationsHubRoutePage.vue` 收口成纯 route 薄壳。
- 在 `features/platform/contests/ui` 新增 `PlatformContestOperationsHubPage.vue`，统一承接 operations hub 的 page model 与 page shell owner。

## Non-goals

- 不改 `useContestOperationsHubPage.ts` 的请求、分页、preferred contest、stale request 处理或 route target contract。
- 不继续拆 `ContestOperationsHubHeroPanel.vue` 或 `ContestOperationsHubWorkspacePanel.vue`。
- 不顺手处理 `ContestAnnouncementsRoutePage.vue`、`ContestEditRoutePage.vue` 或 `ContestOperationsRoutePage.vue`。

## Source Inputs

- `code/frontend/src/pages/platform/contests/ContestOperationsHubRoutePage.vue`
- `code/frontend/src/features/platform/contests/model/useContestOperationsHubPage.ts`
- `code/frontend/src/features/platform/contests/ui/ContestOperationsHubHeroPanel.vue`
- `code/frontend/src/features/platform/contests/ui/ContestOperationsHubWorkspacePanel.vue`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/features/platform/instance-management/ui/PlatformInstanceManagementPage.vue`
- `code/frontend/src/features/platform/contests/ui/PlatformContestManagePage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestOperationsHub.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- 当前更适合的下一刀是 `ContestOperationsHubRoutePage.vue`，因为它没有 props 或 overlay，只是 route page 仍直接持有 page model 和两块 panel 组合，收口面清晰。
- 这轮完成后，`platform/contests` 的 route-level page owner 模式会进一步统一，后续再推进公告页或编辑页时不需要重新试探这一层边界。

## Task Slices

### Slice 1: 新增 operations hub feature page

- 目标：新增 `PlatformContestOperationsHubPage.vue`，承接 `useContestOperationsHubPage()`、hero panel 和 workspace panel 的组合。
- 风险：刷新与翻页事件桥接必须保持现有行为，不引入新的局部状态 owner。

### Slice 2: 收口 route page 与 public API

- 目标：让 `ContestOperationsHubRoutePage.vue` 只渲染 `PlatformContestOperationsHubPage`，并通过 `@/features/platform/contests` 公共出口读取。
- 风险：如果 route page 仍深导入具体 panel 或 page model，owner 收口是不完整的。

### Slice 3: 护栏与 backlog 同步

- 目标：更新 `ContestOperationsHub.test.ts` 的 source-boundary 断言和 backlog 进展，防止 route page 再次直接持有 page model。
- 风险：如果不补断言，这条 route page 很容易回到“薄壳 + 半个 page controller”状态。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision platform-contest-operations-hub-page-owner-cleanup`
- `cd code/frontend && npm run test:run -- src/pages/platform/contests/__tests__/ContestOperationsHub.test.ts`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-contest-operations-hub-page-owner-cleanup.md docs/plan/archive/impl-plan/2026-06/2026-06-01-platform-contest-operations-hub-page-owner-cleanup-plan.md docs/reviews/frontend/2026-06-01-platform-contest-operations-hub-page-owner-cleanup-review.md code/frontend/src/features/platform/contests/ui/PlatformContestOperationsHubPage.vue code/frontend/src/features/platform/contests/ui/index.ts code/frontend/src/pages/platform/contests/ContestOperationsHubRoutePage.vue code/frontend/src/pages/platform/contests/__tests__/ContestOperationsHub.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review Focus

- `PlatformContestOperationsHubPage.vue` 是否真正成为 operations hub 的 page-level owner，而不是只把 route page 内容机械搬家。
- `ContestOperationsHubRoutePage.vue` 是否已经退回纯 feature public API 薄壳。
- source-boundary 护栏是否足够阻止 route page 再次直接拿 `useContestOperationsHubPage()` 或具体 panel。
