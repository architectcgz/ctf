# Reuse Decision

## Change type
frontend refactor / entity presentation owner extraction

## Existing code searched
- `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`
- `code/frontend/src/widgets/notification-detail-workspace/NotificationDetailWorkspace.vue`
- `code/frontend/src/features/notifications/model/useNotificationListPage.ts`
- `code/frontend/src/features/notifications/model/useNotificationDetailPage.ts`
- `code/frontend/src/pages/notifications/__tests__/NotificationList.test.ts`
- `code/frontend/src/pages/notifications/__tests__/NotificationDetail.test.ts`
- `code/frontend/src/entities/challenge/model/presentation.ts`
- `code/frontend/src/entities/challenge/index.ts`
- `code/frontend/src/entities/contest/model/presentation.ts`
- `code/frontend/src/entities/contest/index.ts`
- 前端 feature-sliced architecture 迁移台账

## Similar implementations found
- `entities/challenge/model/presentation.ts` 已经承担稳定业务对象的标签、颜色、状态文案和时间展示规则，适合作为通知实体层的直接参考。
- 通知列表与详情页当前都各自持有 `type -> 文案/颜色/状态` 展示规则，说明这部分 owner 还留在 widget / feature，而不是实体层。

## Decision
refactor_existing

## Reason
当前通知链路的页面壳已经迁到 `pages -> widgets -> features` 主链路，下一步最小正确改动不是继续拆 workflow，而是把通知对象的稳定展示规则收口到 `entities/notification`。

本轮最小正确改动是：

- 新建 `entities/notification`，承接通知类型文案、accent 颜色、已读状态文案这类稳定展示规则
- 让通知列表 / 详情 workspace 改为消费实体层 owner
- 让通知 feature model 停止持有展示规则，只保留筛选、分页、已读同步、详情读取等 workflow owner
- 用测试锁住通知展示规则不再回流到 feature / widget

本轮不做：

- 不改通知读取、分页、实时同步、已读 mutation 流程
- 不改通知发布 feature
- 不新增通知实体层的业务 mutation 或 route owner

## Files to modify
- `.harness/reuse-decisions/notification-entity-presentation-owner.md`
- `docs/plan/impl-plan/2026-05-31-notification-entity-presentation-owner-plan.md`
- `code/frontend/src/entities/notification/index.ts`
- `code/frontend/src/entities/notification/model/index.ts`
- `code/frontend/src/entities/notification/model/presentation.ts`
- `code/frontend/src/entities/notification/model/presentation.test.ts`
- `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`
- `code/frontend/src/widgets/notification-detail-workspace/NotificationDetailWorkspace.vue`
- `code/frontend/src/features/notifications/model/useNotificationListPage.ts`
- `code/frontend/src/features/notifications/model/useNotificationDetailPage.ts`
- `code/frontend/src/pages/notifications/__tests__/NotificationList.test.ts`
- `code/frontend/src/pages/notifications/__tests__/NotificationDetail.test.ts`
- 前端 feature-sliced architecture 迁移台账

## After implementation
- 通知类型/状态的稳定展示规则会有明确的 `entities/notification` owner。
- 通知列表与详情 workspace 不再各自维护重复的通知展示映射。
- 通知 feature model 只保留 workflow owner，不再回答“通知对象怎么展示”。
