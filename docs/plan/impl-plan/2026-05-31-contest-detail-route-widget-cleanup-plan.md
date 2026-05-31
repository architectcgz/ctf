# 竞赛详情路由页 widget 收口计划

## Objective

- 把 `ContestDetailRoutePage.vue` 从厚 route page 收口成标准 `pages` 组合入口。
- 把竞赛详情页的模板和样式壳迁到 `widgets/contest-detail-workspace`。
- 保持 `features/contest-detail/model/useContestDetailRoutePage.ts` 继续作为 route/query、页签、派生状态和页面 workflow owner。

## Non-goals

- 不改竞赛详情 API、AWD 工作台、队伍流程和公告实时桥的 owner。
- 不在本轮重拆 `features/contest-detail/ui/*` 内部组件边界。
- 不处理 `ContestListRoutePage.vue`。

## Source Inputs

- `TODO/frontend-sliced-architecture.md`
- `code/frontend/src/pages/contests/ContestDetailRoutePage.vue`
- `code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts`
- `code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts`
- `code/frontend/src/widgets/notification-detail-workspace/NotificationDetailWorkspace.vue`

## Task Slices

### Slice 1: 新增竞赛详情 widget

- 目标：把竞赛详情页 loading / 空态 / 页签 / 主工作区 / 队伍弹窗外壳迁到 `ContestDetailWorkspace.vue`。
- 变更面：
  - `code/frontend/src/widgets/contest-detail-workspace/ContestDetailWorkspace.vue`
  - `code/frontend/src/widgets/contest-detail-workspace/index.ts`
- 风险：
  - props 透传不全会影响 AWD 切页、公告刷新、Flag 提交和队伍弹窗交互。
- 验证：
  - 竞赛详情页测试。

### Slice 2: route page 收口为组合入口

- 目标：让 `ContestDetailRoutePage.vue` 只负责 `useContestDetailRoutePage()` 和 widget 组合。
- 变更面：
  - `code/frontend/src/pages/contests/ContestDetailRoutePage.vue`
- 风险：
  - route page 如果残留大模板、样式或局部状态，结构收口不完整。
- 验证：
  - 源码级断言 route page 不直接持有主工作区模板和样式壳。

### Slice 3: 同步测试与台账

- 目标：把模板和样式断言切到 widget owner，并在迁移台账里把竞赛详情从当前优先项中移出。
- 变更面：
  - `code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts`
  - `TODO/frontend-sliced-architecture.md`
- 风险：
  - 测试如果继续盯 route page 旧模板，会产生假失败。
- 验证：
  - `ContestDetail.test.ts`

## Validation Plan

- `npm run test:run -- src/pages/contests/__tests__/ContestDetail.test.ts src/__tests__/routePageArchitectureBoundary.test.ts src/__tests__/architectureBoundaries.test.ts`

## Review Focus

- route page 是否已经回到纯组合入口。
- widget 是否只承担页面壳和区块编排，没有拿走 route model owner。
- 台账是否同步反映竞赛详情已完成当前这轮 route page -> widget 收口。

## Rollback / Recovery

- 如果 widget 抽取导致竞赛详情交互回退，可先把 route page 恢复到原模板，再按更小切片拆分页签壳和内容区。
