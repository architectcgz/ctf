# platform contest announcements data split 计划

## Objective

- 在 `platform/contests` 内新增 `useContestAnnouncementsData`，承接竞赛公告页的页面级数据 owner。
- 让 `useContestAnnouncementsPage` 只保留 route、格式化和公告 management 编排。

## Non-goals

- 不改 `ContestAnnouncementsWorkspacePanel` UI。
- 不改 `useContestAnnouncementManagement` 的发布、删除和表单逻辑。
- 不改公告页 route contract。

## Source Inputs

- `code/frontend/src/features/platform/contests/model/useContestAnnouncementsPage.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestAnnouncements.test.ts`
- `code/frontend/src/features/contest-announcements/model/useContestAnnouncementManagement.ts`
- `code/frontend/src/features/platform/overview/model/usePlatformOverviewData.ts`

## Plan Review Result

- 这页适合做 `page + data` 拆分。
- data owner 负责竞赛详情请求、页面 loading/error 和初次加载公告。
- page model 保留 route、时间格式化和对公告 management 的动作调用。

## Task Slices

### Slice 1: 新建 useContestAnnouncementsData

- 目标：收口竞赛详情请求、页面 loading/error 和初次加载公告。
- 风险：
  - 如果把发布/删除动作一起搬走，会和 `useContestAnnouncementManagement` 再次重叠 owner。

### Slice 2: useContestAnnouncementsPage 改为消费 data owner

- 目标：保留 route、格式化和 management 编排。
- 风险：
  - 如果 page 继续直接依赖 `getContest`，就没有真正收口页面级数据 owner。

### Slice 3: 更新源码级和行为测试

- 目标：给新 data owner 补直测，并更新平台端源码断言。
- 风险：
  - 不补失败态或初次加载公告测试，页面级异步 owner 还会回流进 page model。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision platform-contest-announcements-data-split`
- `npm run test:run -- src/features/platform/contests/model/useContestAnnouncementsData.test.ts src/pages/platform/contests/__tests__/ContestAnnouncements.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `npm run typecheck`
- `git diff --check`
- `bash scripts/check-reuse-first.sh`

## Review Focus

- `useContestAnnouncementsData` 是否只承接页面级数据加载 owner。
- `useContestAnnouncementsPage` 是否只剩 route、格式化和 management 编排。
- `useContestAnnouncementManagement` 是否保持发布/删除 owner 不变。

## Rollback / Recovery

- 如果 `useContestAnnouncementsData` 的接口不顺手，可以调整返回结构，但页面级数据加载 owner 仍必须留在新 composable。
