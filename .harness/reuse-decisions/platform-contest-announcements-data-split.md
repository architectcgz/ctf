# Reuse Decision

## Change type
frontend refactor / platform contest announcements data split

## Existing code searched
- `code/frontend/src/features/platform/contests/model/useContestAnnouncementsPage.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestAnnouncements.test.ts`
- `code/frontend/src/features/contest-announcements/model/useContestAnnouncementManagement.ts`
- `code/frontend/src/features/platform/overview/model/usePlatformOverviewData.ts`

## Similar implementations found
- `useContestAnnouncementsPage.ts` 当前同时持有竞赛详情请求、页面 loading/error 和公告工作区编排。
- `useContestAnnouncementManagement.ts` 已经承接了公告列表加载、发布、删除和表单状态。
- `usePlatformOverviewData.ts` 已经证明平台页可以把页面数据 owner 抽到独立 data composable。

## Decision
refactor_existing

## Reason
当前最小正确切片是把平台竞赛公告页拆成：

- `useContestAnnouncementsData`：承接竞赛详情请求、页面 loading/error，以及初次加载公告的页面级协同。
- `useContestAnnouncementsPage`：保留返回工作台 route、格式化函数和对公告 management 的调用编排。

这样可以：

- 去掉 `useContestAnnouncementsPage` 里的 API 请求 owner
- 保持公告 management 继续只负责公告列表、表单和发布/删除动作

本轮不做：

- 不改公告管理 UI
- 不改 `useContestAnnouncementManagement` 的发布/删除行为
- 不引入 shared contest page owner

## Files to modify
- `.harness/reuse-decisions/platform-contest-announcements-data-split.md`
- `docs/plan/impl-plan/2026-05-31-platform-contest-announcements-data-split-plan.md`
- `code/frontend/src/features/platform/contests/model/useContestAnnouncementsData.ts`
- `code/frontend/src/features/platform/contests/model/useContestAnnouncementsData.test.ts`
- `code/frontend/src/features/platform/contests/model/useContestAnnouncementsPage.ts`
- `code/frontend/src/features/platform/contests/model/index.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestAnnouncements.test.ts`

## After implementation
- 平台竞赛公告页的页面级数据加载 owner 会集中到 `useContestAnnouncementsData`。
- `useContestAnnouncementsPage` 只保留 route、格式化和 management 编排。
