# 竞赛列表路由页 widget 收口计划

## Objective

- 把 `ContestListRoutePage.vue` 从厚 route page 收口成标准 `pages` 组合入口。
- 把竞赛列表页的摘要卡、筛选区、目录、空态和分页壳迁到 `widgets/contest-list-workspace`。
- 保持 `features/contest-detail/model/useContestListPage.ts` 继续作为查询、筛选、分页和路由目标 owner。

## Non-goals

- 不改竞赛列表 API、分页实现或 route target owner。
- 不在本轮重拆 `useContestListPage.ts` 的 summary / filter 逻辑。
- 不处理 `ScoreboardDetailRoutePage.vue`。

## Source Inputs

- `TODO/frontend-sliced-architecture.md`
- `code/frontend/src/pages/contests/ContestListRoutePage.vue`
- `code/frontend/src/features/contest-detail/model/useContestListPage.ts`
- `code/frontend/src/pages/contests/__tests__/ContestList.test.ts`
- `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`

## Task Slices

### Slice 1: 新增竞赛列表 widget

- 目标：把竞赛列表页头、概况卡、筛选区、目录、空态和分页壳层迁到 `ContestListWorkspace.vue`。
- 变更面：
  - `code/frontend/src/widgets/contest-list-workspace/ContestListWorkspace.vue`
  - `code/frontend/src/widgets/contest-list-workspace/index.ts`
- 风险：
  - props 透传不全会影响筛选、分页或详情跳转。
- 验证：
  - 竞赛列表页测试。

### Slice 2: route page 收口为组合入口

- 目标：让 `ContestListRoutePage.vue` 只负责 `useContestListPage()` 和 widget 组合。
- 变更面：
  - `code/frontend/src/pages/contests/ContestListRoutePage.vue`
- 风险：
  - route page 如果残留摘要卡、筛选和目录模板，结构收口不完整。
- 验证：
  - 源码级断言 route page 不直接持有摘要卡和列表模板。

### Slice 3: 同步测试与台账

- 目标：把模板和样式断言切到 widget owner，并在迁移台账里把竞赛列表从当前优先项中移出。
- 变更面：
  - `code/frontend/src/pages/contests/__tests__/ContestList.test.ts`
  - `TODO/frontend-sliced-architecture.md`
- 风险：
  - 测试如果继续盯 route page 旧模板，会产生假失败。
- 验证：
  - `ContestList.test.ts`

## Validation Plan

- `npm run test:run -- src/pages/contests/__tests__/ContestList.test.ts src/__tests__/routePageArchitectureBoundary.test.ts src/__tests__/architectureBoundaries.test.ts`

## Review Focus

- route page 是否已经回到纯组合入口。
- widget 是否只承担页面壳和交互编排，没有拿走 feature model 的 owner。
- 台账是否同步反映竞赛列表已完成当前这轮 route page -> widget 收口。

## Rollback / Recovery

- 如果 widget 抽取导致筛选或分页回退，可先把 route page 恢复到原模板，再按更小切片拆摘要卡和目录区。
