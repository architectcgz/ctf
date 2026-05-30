# Reuse Decision

## Change type
frontend architecture / feature-owned shell migration

## Existing code searched
- code/frontend/src/pages/contests/ContestDetailRoutePage.vue
- code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts
- code/frontend/src/pages/contests/__tests__/contestDetailUiStrategy.test.ts
- code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/components/contests/ContestOverviewPanel.vue
- code/frontend/src/components/contests/ContestAnnouncementsPanel.vue
- code/frontend/src/components/contests/ContestAnnouncementsWorkspaceSection.vue
- code/frontend/src/components/contests/ContestChallengeWorkspacePanel.vue
- code/frontend/src/components/contests/ContestTeamPanel.vue
- code/frontend/src/components/contests/ContestTeamWorkspaceSection.vue
- code/frontend/src/components/contests/ContestTeamDialogs.vue
- code/frontend/src/features/contest-detail/**
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `features/auth`、`features/profile`、`features/instance-list`、`features/notifications`、`features/scoreboard` 已证明：只服务单一 route page / capability 的历史 shell，应直接回到 owning feature 的 `ui/`。
- `ContestDetailRoutePage.vue` 当前已经把 route/query/data workflow 收口在 `features/contest-detail/model`，这组 `components/contests/*` 壳只是它的单 consumer UI surface。
- `features/contest-detail` 当前已有明确 page model public API，但还缺少对应的 `ui/` owner；这轮可以直接补齐，不需要新建 widget 或中间桥接层。

## Decision
refactor_existing

## Reason
这轮不是新增 contest detail 能力，而是把一组已经明确只服务 `ContestDetailRoutePage.vue` 的历史壳组件，从 `components/contests/*` 收回 `features/contest-detail/ui/*`：

- `ContestOverviewPanel.vue`
- `ContestAnnouncementsPanel.vue`
- `ContestAnnouncementsWorkspaceSection.vue`
- `ContestChallengeWorkspacePanel.vue`
- `ContestTeamPanel.vue`
- `ContestTeamWorkspaceSection.vue`
- `ContestTeamDialogs.vue`

这样可以让：

- `ContestDetailRoutePage.vue` 只通过 `@/features/contest-detail` 组合它的 page model 和 feature-owned UI
- `components/contests/*` 不再继续作为 contest detail 的历史 owner
- raw-source 测试与类型声明一起对齐到新的 feature public API

不做：

- 不在这轮处理 `components/contests/awd/*`
- 不新增 `widgets/*` 过渡层
- 不改 `ContestDetailRoutePage.vue` 的页面行为、route owner 或 AWD workspace owner

## Files to modify
- .harness/reuse-decisions/contest-detail-feature-owned-shell-migration-batch1.md
- docs/plan/impl-plan/2026-05-30-contest-detail-feature-owned-shell-migration-batch1-plan.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/components.d.ts
- code/frontend/src/components/contests/ContestOverviewPanel.vue
- code/frontend/src/components/contests/ContestAnnouncementsPanel.vue
- code/frontend/src/components/contests/ContestAnnouncementsWorkspaceSection.vue
- code/frontend/src/components/contests/ContestChallengeWorkspacePanel.vue
- code/frontend/src/components/contests/ContestTeamPanel.vue
- code/frontend/src/components/contests/ContestTeamWorkspaceSection.vue
- code/frontend/src/components/contests/ContestTeamDialogs.vue
- code/frontend/src/features/contest-detail/index.ts
- code/frontend/src/features/contest-detail/ui/index.ts
- code/frontend/src/features/contest-detail/ui/ContestOverviewPanel.vue
- code/frontend/src/features/contest-detail/ui/ContestAnnouncementsPanel.vue
- code/frontend/src/features/contest-detail/ui/ContestAnnouncementsWorkspaceSection.vue
- code/frontend/src/features/contest-detail/ui/ContestChallengeWorkspacePanel.vue
- code/frontend/src/features/contest-detail/ui/ContestTeamPanel.vue
- code/frontend/src/features/contest-detail/ui/ContestTeamWorkspaceSection.vue
- code/frontend/src/features/contest-detail/ui/ContestTeamDialogs.vue
- code/frontend/src/pages/contests/ContestDetailRoutePage.vue
- code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts
- code/frontend/src/pages/contests/__tests__/contestDetailUiStrategy.test.ts
- code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts

## After implementation
- `features/contest-detail/ui` 会成为 contest detail 这组 student workspace shell 的唯一 owner。
- `ContestDetailRoutePage.vue` 将不再直接 import `@/components/contests/*`。
- `components/contests/*` 这一批已迁走文件会从历史目录中删除，不保留桥接壳。
