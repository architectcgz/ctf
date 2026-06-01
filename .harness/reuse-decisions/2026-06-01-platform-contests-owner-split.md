# Reuse Decision

## Change type
refactor_existing

## Existing code searched
- code/frontend/src/features/platform/contests/model/*.ts
- code/frontend/src/features/platform/contests/ui/*.vue
- code/frontend/src/pages/platform/contests/*.vue

## Similar implementations found
- 无 — 这是按能力域拆分已有 feature 的纯目录重组，文件内容不变。
- 已有 5 份 6/1 plan 定义了 route-page → feature-page 的收口模式，
  本次拆分是该模式在 feature 内部的延续。

## Decision
refactor_existing

## Reason
47 文件/140KB 的 platform/contests 按能力域拆为三个 feature：
- contest-manage (CRUD + 编排)
- contest-announcements (公告)
- contest-operations (运维)

所有文件通过 git mv 移动，保持历史；内容仅改 import 路径。
旧 contests/index.ts 保留向后兼容重导出。

## Files to modify
- code/frontend/src/features/platform/contest-manage/** (git mv from contests)
- code/frontend/src/features/platform/contest-announcements/** (git mv from contests)
- code/frontend/src/features/platform/contest-operations/** (git mv from contests)
- code/frontend/src/features/platform/contests/index.ts (兼容重导出)
- code/frontend/src/pages/platform/contests/__tests__/*.test.ts (import 路径更新)
- code/frontend/src/features/platform/contest-manage/model/contestFormSupport.ts
- code/frontend/src/features/platform/contest-manage/model/useAwdStartOverrideFlow.ts
- code/frontend/src/features/platform/contest-manage/model/useContestDialogState.ts
- code/frontend/src/features/platform/contest-manage/model/useContestEditPage.ts
- code/frontend/src/features/platform/contest-manage/model/useContestListState.ts
- code/frontend/src/features/platform/contest-manage/model/useContestManagePage.ts
- code/frontend/src/features/platform/contest-manage/model/useContestManagePanelRoute.ts
- code/frontend/src/features/platform/contest-manage/model/useContestSaveFlow.ts
- code/frontend/src/features/platform/contest-manage/model/usePlatformContests.ts
- code/frontend/src/features/platform/contest-manage/ui/PlatformContestFormActions.vue
- code/frontend/src/features/platform/contest-manage/ui/PlatformContestFormDialog.vue
- code/frontend/src/features/platform/contest-manage/ui/PlatformContestFormPanel.vue
- code/frontend/src/features/platform/contest-manage/ui/PlatformContestFormSectionShell.vue
- code/frontend/src/features/platform/contest-manage/ui/PlatformContestTable.vue
- code/frontend/src/features/platform/contest-manage/ui/PlatformContestTable.test.ts
- code/frontend/src/features/platform/contest-announcements/model/useContestAnnouncementsData.ts
- code/frontend/src/features/platform/contest-announcements/model/useContestAnnouncementsPage.ts
- code/frontend/src/features/platform/contest-announcements/model/useContestAnnouncementsData.test.ts
- code/frontend/src/features/platform/contest-operations/model/useContestOperationsData.ts
- code/frontend/src/features/platform/contest-operations/model/useContestOperationsPage.ts
- code/frontend/src/features/platform/contest-operations/model/useContestOperationsHubPage.ts
- code/frontend/src/features/platform/contest-operations/model/useContestOperationsData.test.ts
