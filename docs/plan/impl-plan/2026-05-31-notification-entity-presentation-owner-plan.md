# 通知实体展示 owner 收口计划

## Objective

- 新建 `entities/notification`，收口通知对象的稳定展示规则。
- 让通知列表 / 详情链路改为消费实体层 owner，而不是在 widget / feature 里重复维护通知展示映射。

## Non-goals

- 不修改通知列表分页、筛选、已读同步和详情读取流程。
- 不修改管理员发布通知 feature。
- 不把 route、toast、mark-as-read、related link 导航搬进实体层。

## Source Inputs

- `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`
- `code/frontend/src/widgets/notification-detail-workspace/NotificationDetailWorkspace.vue`
- `code/frontend/src/features/notifications/model/useNotificationListPage.ts`
- `code/frontend/src/features/notifications/model/useNotificationDetailPage.ts`
- `code/frontend/src/entities/challenge/model/presentation.ts`
- `TODO/frontend-sliced-architecture.md`

## Task Slices

### Slice 1: 建立通知实体展示 owner

- 目标：在 `entities/notification` 中定义通知类型标签、accent 颜色、已读状态标签等稳定展示规则。
- 变更面：
  - `code/frontend/src/entities/notification/**`
- 风险：
  - 如果把 related link、route target、分页筛选带进实体层，会重新混淆 workflow owner 与展示 owner。

### Slice 2: 改通知列表 / 详情消费面

- 目标：让通知两个 workspace 改为从实体层取展示规则，并把 feature model 里的重复映射移除。
- 变更面：
  - `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`
  - `code/frontend/src/widgets/notification-detail-workspace/NotificationDetailWorkspace.vue`
  - `code/frontend/src/features/notifications/model/useNotificationListPage.ts`
  - `code/frontend/src/features/notifications/model/useNotificationDetailPage.ts`
- 风险：
  - 如果 widget 仍然保留旧的本地映射，就会形成“双 owner”。

### Slice 3: 锁住边界与迁移台账

- 目标：让测试和台账明确通知展示规则已经迁入实体层。
- 变更面：
  - `code/frontend/src/pages/notifications/__tests__/NotificationList.test.ts`
  - `code/frontend/src/pages/notifications/__tests__/NotificationDetail.test.ts`
  - `TODO/frontend-sliced-architecture.md`
- 风险：
  - 如果测试只验证 UI 文案，不验证 owner，会让后续回流难以及时暴露。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision notification-entity-presentation-owner`
- `npm run test:run -- src/pages/notifications/__tests__/NotificationList.test.ts src/pages/notifications/__tests__/NotificationDetail.test.ts src/entities/notification/model/presentation.test.ts`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-reuse-first.sh`

## Review Focus

- 通知实体层是否只承接稳定对象展示规则，而没有吸入 workflow / route owner。
- 通知列表 / 详情是否已经消除重复的 `type -> 文案/颜色/状态` 映射。
- 原有通知页面行为是否保持不变。

## Rollback / Recovery

- 如果实体层抽取后让 workspace/feature contract 变得更难读，可保留实体 model 函数但回退 UI 层抽取，前提是重复映射已被移除且 owner 仍然唯一。
