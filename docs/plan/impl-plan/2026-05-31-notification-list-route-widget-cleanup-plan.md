# 通知列表路由页 widget 收口计划

## Objective

- 把 `NotificationListRoutePage.vue` 从厚 route page 收口成标准 `pages` 组合入口。
- 把通知列表页的模板和样式壳迁到 `widgets/notification-list-workspace`。
- 保持 `features/notifications/model/useNotificationListPage.ts` 继续作为分页、筛选、批量已读和发布抽屉 owner。

## Non-goals

- 不改通知列表的 API / store / pagination owner。
- 不在本轮把 `NotificationCategoryFilter` 再拆层。
- 不处理 `ContestDetailRoutePage.vue`。

## Source Inputs

- `TODO/frontend-sliced-architecture.md`
- `code/frontend/src/pages/notifications/NotificationListRoutePage.vue`
- `code/frontend/src/features/notifications/model/useNotificationListPage.ts`
- `code/frontend/src/pages/notifications/__tests__/NotificationList.test.ts`
- `code/frontend/src/widgets/notification-detail-workspace/NotificationDetailWorkspace.vue`

## Task Slices

### Slice 1: 新增通知列表 widget

- 目标：把通知列表的页头、筛选、目录、空态、分页和发布抽屉壳层迁到 `NotificationListWorkspace.vue`。
- 变更面：
  - `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`
  - `code/frontend/src/widgets/notification-list-workspace/index.ts`
- 风险：
  - props 透传不全会影响筛选、刷新、批量已读或发布抽屉交互。
- 验证：
  - 通知列表页测试。

### Slice 2: route page 收口为组合入口

- 目标：让 `NotificationListRoutePage.vue` 只负责 `useNotificationListPage()` 和 widget 组合。
- 变更面：
  - `code/frontend/src/pages/notifications/NotificationListRoutePage.vue`
- 风险：
  - route page 如果残留旧模板和样式，无法真正实现结构收口。
- 验证：
  - 源码级断言 route page 不直接碰 API、pagination 实现和列表模板片段。

### Slice 3: 同步测试与台账

- 目标：把样式和模板断言切到 widget owner，并在迁移台账里把通知列表从当前优先项中移出。
- 变更面：
  - `code/frontend/src/pages/notifications/__tests__/NotificationList.test.ts`
  - `TODO/frontend-sliced-architecture.md`
- 风险：
  - 测试如果继续盯 route page 旧源码，会造成假失败。
- 验证：
  - `NotificationList.test.ts`

## Validation Plan

- `npm run test:run -- src/pages/notifications/__tests__/NotificationList.test.ts src/__tests__/routePageArchitectureBoundary.test.ts src/__tests__/architectureBoundaries.test.ts`

## Review Focus

- route page 是否变成纯组合入口。
- widget 是否只承担展示和页面区块编排，没有拿走 feature model 的 owner。
- 台账是否同步反映通知列表已完成当前这轮 route page -> widget 收口。

## Rollback / Recovery

- 如果 widget 抽取导致通知列表交互回退，可先把 route page 恢复到原模板，再按更小切片重新拆分目录区和页头区。
