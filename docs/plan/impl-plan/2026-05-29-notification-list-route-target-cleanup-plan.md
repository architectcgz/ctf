> 状态：Current
> 事实源：notification list page model、notification route view、前端架构 allowlist
> 替代：无

# Notification List Route Target Cleanup Plan

## 目标

- 把 `useNotificationListPage.ts` 从单次 `router.push()` 收口为纯 route target helper。
- 保持通知列表数据 owner 不变，同时再清掉 1 条 `featureRouterImportAllowlist`。

## 非目标

- 不处理通知详情页 route owner。
- 不重写通知列表样式、分页或批量已读逻辑。

## 输入依据

- `code/frontend/src/features/notifications/model/useNotificationListPage.ts`
- `code/frontend/src/views/notifications/NotificationList.vue`
- `code/frontend/src/views/notifications/__tests__/NotificationList.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- `useNotificationListPage.ts` 当前对 router 的唯一依赖是详情跳转。
- `NotificationList.vue` 是完整 route view 模板，可以直接消费 `RouterLink`，因此没必要继续保留 `router.push()`。
- 这一条适合按“route target contract + RouterLink”处理，而不是新造一个 route-aware wrapper。

## 设计边界

### `useNotificationListPage()` 本轮负责

- 通知列表数据读取、分页、分类、批量已读、发布抽屉和刷新提示 owner
- 生成通知详情 route target helper

### `NotificationList.vue` 本轮负责

- 用 `RouterLink` 消费 route target helper
- 保持 route view 不直接 `useRouter`

## 任务切片

### Slice 1：page model 去 router 化

- 目标：
  - 移除 `useNotificationListPage.ts` 对 `vue-router` 的依赖
  - 改成纯 route target helper
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/notifications/__tests__/NotificationList.test.ts`
- Review focus：
  - `useNotificationListPage.ts` 不再 import `vue-router`

### Slice 2：route view 与测试切到 RouterLink

- 目标：
  - 通知列表项改为 `RouterLink`
  - 测试继续覆盖行点击跳转与 route view 边界
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/notifications/__tests__/NotificationList.test.ts`
- Review focus：
  - route view 仍只组合，不直接拿 router

### Slice 3：allowlist / review / backlog 收尾

- 目标：
  - 更新 allowlist、review 和 backlog
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - `featureRouterImportAllowlist` 是否净减少 1 条

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/notifications/__tests__/NotificationList.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/notification-list-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-notification-list-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-notification-list-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/notifications/model/useNotificationListPage.ts code/frontend/src/views/notifications/NotificationList.vue code/frontend/src/views/notifications/__tests__/NotificationList.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `useNotificationDetailPage.ts` 仍保留 route param owner，这轮不一并处理。
