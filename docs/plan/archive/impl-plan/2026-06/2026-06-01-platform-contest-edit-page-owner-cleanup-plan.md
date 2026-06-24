# Platform Contest Edit Page Owner Cleanup 计划

## Objective

- 把 `ContestEditRoutePage.vue` 收口成纯 route 薄壳。
- 在 `features/platform/contests/ui` 新增 `PlatformContestEditPage.vue`，统一承接竞赛编辑页的 page model 与 page shell owner。

## Non-goals

- 不改 `useContestEditPage.ts` 内部的 AWD workbench 数据加载、save redirect、query-tab、toast 或 breadcrumb owner。
- 不改 `ContestEditWorkspacePanel`、`ContestEditTopbarPanel`、`ContestWorkbenchStageTabs` 的现有 contract。
- 不继续处理 `contest-awd-config` 等更深层 route owner。

## Source Inputs

- `code/frontend/src/pages/platform/contests/ContestEditRoutePage.vue`
- `code/frontend/src/features/platform/contests/model/useContestEditPage.ts`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/features/platform/contests/ui/PlatformContestManagePage.vue`
- `code/frontend/src/features/platform/contests/ui/PlatformContestAnnouncementsPage.vue`
- `code/frontend/src/features/platform/contests/ui/PlatformContestOperationsPage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestEdit.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- 这页虽然比 `ContestOperationsRoutePage.vue` 更重，但当前 route 层超出的 owner 仍是标准 page shell 组合面，不需要进一步拆 model 才能收口。
- 保留 `contestId` route prop 透传，能让这轮只处理 route shell 与 feature page owner，不扩大到 `useContestEditPage.ts` 内部已有的 query / redirect / workbench owner。

## Task Slices

### Slice 1: 新增 edit feature page

- 目标：新增 `PlatformContestEditPage.vue`，承接 `useContestEditPage()`、redirect shell、topbar、stage tabs、workspace panel 与 loading shell。
- 风险：save 事件桥接、AWD workbench refresh 透传和 redirect 行为必须保持现有编辑页行为。

### Slice 2: 收口 route page 与 public API

- 目标：让 `ContestEditRoutePage.vue` 只接受 `contestId` prop 并渲染 `PlatformContestEditPage`，通过 `@/features/platform/contests` 公共出口读取。
- 风险：如果 route page 仍直接引用 page model 或 page shell 组件，owner 收口不完整。

### Slice 3: 护栏与 backlog 同步

- 目标：更新 `ContestEdit.test.ts` 的 source-boundary 断言和 backlog 进展，防止 route page 再次直接持有 page model 与编辑工作台壳。
- 风险：如果只改实现不改断言，这页很容易回到“薄壳 + page model + page shell”混合状态。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision platform-contest-edit-page-owner-cleanup`
- `cd code/frontend && npm run test:run -- src/pages/platform/contests/__tests__/ContestEdit.test.ts`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-contest-edit-page-owner-cleanup.md docs/plan/archive/impl-plan/2026-06/2026-06-01-platform-contest-edit-page-owner-cleanup-plan.md docs/reviews/frontend/2026-06-01-platform-contest-edit-page-owner-cleanup-review.md code/frontend/src/features/platform/contests/ui/PlatformContestEditPage.vue code/frontend/src/features/platform/contests/ui/index.ts code/frontend/src/pages/platform/contests/ContestEditRoutePage.vue code/frontend/src/pages/platform/contests/__tests__/ContestEdit.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review Focus

- `PlatformContestEditPage.vue` 是否真正成为 edit page-level owner，而不是 route page 内容的机械搬家。
- `ContestEditRoutePage.vue` 是否已经退回 “route prop -> feature page” 的薄壳。
- source-boundary 护栏是否足够防止 route page 再次直接拿 `useContestEditPage()`、topbar、stage tabs 或 workspace panel。
