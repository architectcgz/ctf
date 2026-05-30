# Reuse Decision

## Change type
frontend architecture / feature-owned shell migration

## Existing code searched
- code/frontend/src/pages/auth/*.vue
- code/frontend/src/pages/profile/*.vue
- code/frontend/src/pages/instances/*.vue
- code/frontend/src/pages/notifications/*.vue
- code/frontend/src/pages/scoreboard/*.vue
- code/frontend/src/components/auth/AuthEntryShell.vue
- code/frontend/src/components/profile/*.vue
- code/frontend/src/components/instance/InstanceListWorkspaceShell.vue
- code/frontend/src/components/notifications/NotificationCategoryFilter.vue
- code/frontend/src/components/scoreboard/ScoreboardWorkspaceShell.vue
- code/frontend/src/features/auth/**
- code/frontend/src/features/profile/**
- code/frontend/src/features/skill-profile/**
- code/frontend/src/features/instance-list/**
- code/frontend/src/features/notifications/**
- code/frontend/src/features/scoreboard/**
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `features/teacher/**`、`features/teaching/**`、`features/platform/**` 已经证明，单 capability 的 page shell 和 workspace shell 应直接落到 owning feature 的 `ui/`，而不是继续挂在历史 `components/*`。
- `pages/**` 当前已经是唯一运行时 route entry，因此这轮只需要把 route page 组合用到的单能力 shell 从 `components/*` 收回 feature owner，不需要再改页面层职责。
- `features/scoreboard/ui/ScoreboardRealtimeBridge.vue` 已经说明 `scoreboard` feature 允许持有本地 `ui/` public API；`auth`、`profile`、`instance-list`、`notifications` 也具备同类收口条件。

## Decision
refactor_existing

## Reason
这轮不是新增页面，也不是继续清理 `views/**`，而是收口一批已经明确只服务单一 capability 的历史业务壳组件：

- `code/frontend/src/components/auth/AuthEntryShell.vue` -> `code/frontend/src/features/auth/ui/AuthEntryShell.vue`
- `code/frontend/src/components/profile/SecuritySettingsWorkspaceShell.vue` -> `code/frontend/src/features/profile/ui/SecuritySettingsWorkspaceShell.vue`
- `code/frontend/src/components/profile/UserProfileWorkspaceShell.vue` -> `code/frontend/src/features/profile/ui/UserProfileWorkspaceShell.vue`
- `code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue` -> `code/frontend/src/features/skill-profile/ui/SkillProfileWorkspaceShell.vue`
- `code/frontend/src/components/instance/InstanceListWorkspaceShell.vue` -> `code/frontend/src/features/instance-list/ui/InstanceListWorkspaceShell.vue`
- `code/frontend/src/components/notifications/NotificationCategoryFilter.vue` -> `code/frontend/src/features/notifications/ui/NotificationCategoryFilter.vue`
- `code/frontend/src/components/scoreboard/ScoreboardWorkspaceShell.vue` -> `code/frontend/src/features/scoreboard/ui/ScoreboardWorkspaceShell.vue`

同时同步 route page、邻近测试、`features/*/index.ts` public API 与 backlog 当前事实，让 `components/auth|profile|instance|notifications|scoreboard` 不再继续作为单能力页面壳 owner。

不做：

- 不在这轮处理 `components/contests/*`
- 不在这轮处理 `components/challenge/*`
- 不扩大到 `components/contests/awd/*`
- 不改 `pages/**` 的路由层职责，只改 shell 落位与 feature public API

## Files to modify
- .harness/reuse-decisions/feature-owned-shell-migration-batch1.md
- docs/plan/impl-plan/2026-05-30-feature-owned-shell-migration-batch1-plan.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/components/auth/AuthEntryShell.vue
- code/frontend/src/components/profile/SecuritySettingsWorkspaceShell.vue
- code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue
- code/frontend/src/components/profile/UserProfileWorkspaceShell.vue
- code/frontend/src/components/instance/InstanceListWorkspaceShell.vue
- code/frontend/src/components/notifications/NotificationCategoryFilter.vue
- code/frontend/src/components/scoreboard/ScoreboardWorkspaceShell.vue
- code/frontend/src/features/auth/index.ts
- code/frontend/src/features/auth/ui/AuthEntryShell.vue
- code/frontend/src/features/profile/index.ts
- code/frontend/src/features/profile/ui/SecuritySettingsWorkspaceShell.vue
- code/frontend/src/features/profile/ui/UserProfileWorkspaceShell.vue
- code/frontend/src/features/skill-profile/index.ts
- code/frontend/src/features/skill-profile/ui/SkillProfileWorkspaceShell.vue
- code/frontend/src/features/instance-list/index.ts
- code/frontend/src/features/instance-list/ui/InstanceListWorkspaceShell.vue
- code/frontend/src/features/notifications/index.ts
- code/frontend/src/features/notifications/ui/NotificationCategoryFilter.vue
- code/frontend/src/features/scoreboard/index.ts
- code/frontend/src/features/scoreboard/ui/ScoreboardWorkspaceShell.vue
- code/frontend/src/pages/auth/LoginRoutePage.vue
- code/frontend/src/pages/auth/RegisterRoutePage.vue
- code/frontend/src/pages/profile/SecuritySettingsRoutePage.vue
- code/frontend/src/pages/profile/SkillProfileRoutePage.vue
- code/frontend/src/pages/profile/UserProfileRoutePage.vue
- code/frontend/src/pages/instances/InstanceListRoutePage.vue
- code/frontend/src/pages/notifications/NotificationListRoutePage.vue
- code/frontend/src/pages/scoreboard/ScoreboardViewRoutePage.vue
- code/frontend/src/pages/auth/__tests__/LoginRoutePage.test.ts
- code/frontend/src/pages/profile/__tests__/SecuritySettings.test.ts
- code/frontend/src/pages/profile/__tests__/SkillProfile.test.ts
- code/frontend/src/pages/profile/__tests__/UserProfile.test.ts
- code/frontend/src/pages/instances/__tests__/InstanceList.test.ts
- code/frontend/src/pages/notifications/__tests__/NotificationList.test.ts
- code/frontend/src/pages/scoreboard/__tests__/ScoreboardView.test.ts
- code/frontend/src/pages/__tests__/journalEyebrowStyles.test.ts
- code/frontend/src/pages/__tests__/journalUserDirectoryButtonVariants.test.ts
- code/frontend/src/pages/__tests__/journalUserDirectoryStyles.test.ts
- code/frontend/src/pages/__tests__/journalUserShellStyles.test.ts
- code/frontend/src/pages/__tests__/pageTabsStyles.test.ts
- code/frontend/src/pages/__tests__/profileJournalButtonStyles.test.ts
- code/frontend/src/pages/__tests__/profileJournalNoteStyles.test.ts
- code/frontend/src/pages/__tests__/profileJournalUtilityStyles.test.ts
- code/frontend/src/pages/__tests__/studentDirectoryTypographyBoundary.test.ts
- code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts
- code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts

## After implementation
- `components/auth`、`components/profile`、`components/instance`、`components/notifications`、`components/scoreboard` 将不再承载单 capability 的页面壳 owner。
- 相关 route page 会改为从 owning feature 的 public API 或 feature-local `ui/` 引用 shell。
- backlog 中这批“单 capability workspace shell”条目会与当前代码事实一致。
