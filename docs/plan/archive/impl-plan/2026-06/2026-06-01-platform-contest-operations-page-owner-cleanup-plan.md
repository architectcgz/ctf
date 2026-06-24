# Platform Contest Operations Page Owner Cleanup 计划

## Objective

- 把 `ContestOperationsRoutePage.vue` 收口成纯 route 薄壳。
- 在 `features/platform/contests/ui` 新增 `PlatformContestOperationsPage.vue`，统一承接单场运维页的 page model 与 page shell owner。

## Non-goals

- 不改 `useContestOperationsPage.ts` 的赛事加载、breadcrumb、toast 或 runtime/readiness 判定。
- 不改 `AWDOperationsPanel`、`AWDServiceAlertBanner` 与 `contest-awd-admin` 相关 feature owner。
- 不继续处理 `ContestEditRoutePage.vue`。

## Source Inputs

- `code/frontend/src/pages/platform/contests/ContestOperationsRoutePage.vue`
- `code/frontend/src/features/platform/contests/model/useContestOperationsPage.ts`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/features/platform/contests/ui/PlatformContestManagePage.vue`
- `code/frontend/src/features/platform/contests/ui/PlatformContestOperationsHubPage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestOperations.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- 这页比 `ContestEditRoutePage.vue` 更适合作为下一刀，因为它只有一个 route prop，没有 query-tab、save redirect 或 AWD workbench 切面，owner 收口边界更清楚。
- 保留 `contestId` route prop 透传，能让这轮只处理 route shell 与 feature page owner，不扩大到 route param 输入边界本身。

## Task Slices

### Slice 1: 新增 operations feature page

- 目标：新增 `PlatformContestOperationsPage.vue`，承接 `useContestOperationsPage()`、loading shell、`AWDOperationsPanel` 与服务告警插槽组合。
- 风险：`AWDOperationsPanel` 的固定 props 和插槽桥接必须保持现有运维页行为。

### Slice 2: 收口 route page 与 public API

- 目标：让 `ContestOperationsRoutePage.vue` 只接受 `contestId` prop 并渲染 `PlatformContestOperationsPage`，通过 `@/features/platform/contests` 公共出口读取。
- 风险：如果 route page 仍直接引用 page model 或运维面板，owner 收口不完整。

### Slice 3: 护栏与 backlog 同步

- 目标：更新 `ContestOperations.test.ts` 的 source-boundary 断言和 backlog 进展，防止 route page 再次直接持有 page model 与运维布局。
- 风险：如果只改实现不改断言，这条路由页很容易回到“薄壳 + page model + layout”混合状态。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision platform-contest-operations-page-owner-cleanup`
- `cd code/frontend && npm run test:run -- src/pages/platform/contests/__tests__/ContestOperations.test.ts`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-contest-operations-page-owner-cleanup.md docs/plan/archive/impl-plan/2026-06/2026-06-01-platform-contest-operations-page-owner-cleanup-plan.md docs/reviews/frontend/2026-06-01-platform-contest-operations-page-owner-cleanup-review.md code/frontend/src/features/platform/contests/ui/PlatformContestOperationsPage.vue code/frontend/src/features/platform/contests/ui/index.ts code/frontend/src/pages/platform/contests/ContestOperationsRoutePage.vue code/frontend/src/pages/platform/contests/__tests__/ContestOperations.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review Focus

- `PlatformContestOperationsPage.vue` 是否真正成为 operations 的 page-level owner，而不是 route page 内容的机械搬家。
- `ContestOperationsRoutePage.vue` 是否已经退回 “route prop -> feature page” 的薄壳。
- source-boundary 护栏是否足够防止 route page 再次直接拿 `useContestOperationsPage()` 或运维布局。
