# Reuse Decision

## Change type
refactor_existing

## Existing code searched
- code/frontend/src/features/platform/contests/model/*.ts
- code/frontend/src/features/platform/contests/ui/*.vue
- code/frontend/src/pages/platform/contests/*.vue
- code/frontend/src/api/admin/contests.ts
- code/frontend/src/api/admin/index.ts
- code/frontend/src/api/teaching/awd-reviews.ts
- code/frontend/src/features/contest-awd-config/model/*.ts
- code/frontend/src/features/contest-projector/model/*.ts

## Similar implementations found
- 无 — 这是按能力域拆分已有 feature 的纯目录重组，文件内容不变。
- 已有 5 份 6/1 plan 定义了 route-page → feature-page 的收口模式，
  本次拆分是该模式在 feature 内部的延续。
- `api/admin/teaching.ts` / `api/admin/authoring.ts`
  已经采用“public owner wrapper + shared implementation reuse”的模式，
  本次 `contest-reviews.ts` 继续沿用该 owner 组织方式。

## Decision
refactor_existing

## Reason
这轮不是新功能，而是沿着既有 page owner split 继续把 contest domain 的 runtime owner 收口到底：

- `platform/contests` 按能力域拆为三个 feature
- contest-manage (CRUD + 编排)
- contest-announcements (公告)
- contest-operations (运维)
- `api/admin/contests.ts` 再按 use case 下沉成 5 个 owning modules，
  让 platform contest runtime consumer 直接依赖具体 transport owner。
- 旧 `features/platform/contests/index.ts` 删除；
  旧 `api/admin/contests.ts` 仅保留 test/mock 兼容重导出。

优先复用已有 platform wrapper / shared implementation 组织方式，
避免在 AWD review、AWD config、projector round loader 上再造一层新的中间 facade。

## Files to modify
- code/frontend/src/features/platform/contest-manage/** (git mv from contests)
- code/frontend/src/features/platform/contest-announcements/** (git mv from contests)
- code/frontend/src/features/platform/contest-operations/** (git mv from contests)
- code/frontend/src/features/platform/contests/index.ts (删除兼容重导出)
- code/frontend/src/pages/platform/contests/__tests__/*.test.ts (import 路径更新)
- code/frontend/src/api/admin/contest-support.ts
- code/frontend/src/api/admin/contest-manage.ts
- code/frontend/src/api/admin/contest-announcements.ts
- code/frontend/src/api/admin/contest-operations.ts
- code/frontend/src/api/admin/contest-awd-admin.ts
- code/frontend/src/api/admin/contest-reviews.ts
- code/frontend/src/api/admin/contests.ts
- code/frontend/src/api/admin/index.ts
- code/frontend/src/api/__tests__/admin.test.ts
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
- code/frontend/src/features/contest-awd-config/model/useContestAwdConfigDataLoader.ts
- code/frontend/src/features/contest-projector/model/useProjectorRoundSnapshotLoader.ts
