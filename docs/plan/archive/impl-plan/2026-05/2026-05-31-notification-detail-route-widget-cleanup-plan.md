# 通知详情路由页 widget 收口计划

## Objective

- 把 `NotificationDetailRoutePage.vue` 从厚 route page 收口成标准 `pages` 入口。
- 把通知详情的大模板和样式壳下沉到 `widgets` 层。
- 保持 `features/notifications/model/useNotificationDetailPage.ts` 继续作为通知详情读取、已读同步和探针交互 owner。

## Non-goals

- 不改通知详情的 API / store owner。
- 不在本轮引入 `entities/notification`。
- 不处理 `NotificationListRoutePage.vue`。

## Source Inputs

- `TODO/frontend-sliced-architecture.md`
- `code/frontend/src/pages/notifications/NotificationDetailRoutePage.vue`
- `code/frontend/src/features/notifications/model/useNotificationDetailPage.ts`
- `code/frontend/src/pages/notifications/__tests__/NotificationDetail.test.ts`
- `code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue`

## Task Slices

### Slice 1: 新增通知详情 widget

- 目标：把通知详情页面模板和样式迁到 `widgets/notification-detail-workspace/NotificationDetailWorkspace.vue`。
- 变更面：
  - `code/frontend/src/widgets/notification-detail-workspace/NotificationDetailWorkspace.vue`
  - `code/frontend/src/widgets/notification-detail-workspace/index.ts`
- 风险：
  - widget 透传 props 不完整会导致关联入口、加载态或探针交互回退。
- 验证：
  - 通知详情页测试。

### Slice 2: route page 收口为组合入口

- 目标：让 `NotificationDetailRoutePage.vue` 只负责 props -> feature model -> widget 组合。
- 变更面：
  - `code/frontend/src/pages/notifications/NotificationDetailRoutePage.vue`
- 风险：
  - route page 如果残留大量模板或样式，收口目标会失效。
- 验证：
  - 源码级断言 route page 不直接碰 API / router / watch，并且改为组合 widget。

### Slice 3: 更新测试与迁移台账

- 目标：同步把源码断言改到 widget owner，并在迁移台账里记录通知详情页已完成这轮瘦身。
- 变更面：
  - `code/frontend/src/pages/notifications/__tests__/NotificationDetail.test.ts`
  - `TODO/frontend-sliced-architecture.md`
- 风险：
  - 测试如果仍盯 route page 的旧模板片段，会造成假失败。
- 验证：
  - `NotificationDetail.test.ts`

## Validation Plan

- `npm run test:run -- src/pages/notifications/__tests__/NotificationDetail.test.ts src/__tests__/routePageArchitectureBoundary.test.ts src/__tests__/architectureBoundaries.test.ts`

## Review Focus

- route page 是否真正变成组合入口。
- widget 是否只承担展示和区块编排，没有反向拿走 feature owner。
- 台账是否同步反映通知详情页已从厚 route page 列表中移出。

## Rollback / Recovery

- 如果 widget 抽取导致通知详情交互回退，可先回退 route page 到原始模板，同时保留台账和计划单独重新切片。
