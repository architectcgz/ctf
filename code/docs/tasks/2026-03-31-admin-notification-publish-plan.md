# Admin Notification Publish Implementation Plan

> 状态更新（2026-03-31）：
> 该方案已执行完成。管理员现在可以在 `/notifications` 页面通过抽屉表单发布通知，后端新增批次投递模型与管理端发布接口，通知收件箱读路径保持不变。

**Goal:** Allow admins to publish notifications from `/notifications` with audience targeting for all users, roles, classes, and specific users, while preserving the existing per-user inbox model.

**Architecture:** Introduce a publish-batch layer in the backend so admin publishing and user inbox delivery are modeled separately. Keep `notifications` as the delivery table, add a batch table plus audience metadata, and surface an admin-only publishing drawer on the existing notifications page using existing user/class query APIs.

**Tech Stack:** Go, Gin, GORM, PostgreSQL migrations, Vue 3, Pinia, Vitest

---

## 实际落地结果

- 后端新增 `notification_batches` 批次表，并在 `notifications` 上增加 `batch_id` 关联，采用“批次元信息 + 每用户投递记录”模型。
- 管理端新增 `POST /api/v1/admin/notifications`，仅 `admin` 可调用。
- 请求契约收敛为：
  - `audience_rules.mode = "union"`
  - `audience_rules.rules[].type in {all, role, class, user}`
  - 非 `all` 规则统一使用 `values`
- 前端 `/notifications` 页面新增管理员专用“发布通知”入口，使用抽屉表单完成发布。
- 当前 UI 支持四种单一目标范围：
  - 所有用户
  - 按角色
  - 指定班级
  - 指定用户
- 后端保留 union 规则模型，便于后续扩展多规则组合；当前前端未开放多规则混合配置。

## 变更文件

- 后端：
  - `code/backend/migrations/000042_create_notification_batches_table.up.sql`
  - `code/backend/migrations/000042_create_notification_batches_table.down.sql`
  - `code/backend/internal/model/notification_batch.go`
  - `code/backend/internal/model/notification.go`
  - `code/backend/internal/dto/notification.go`
  - `code/backend/internal/module/ops/ports/notification.go`
  - `code/backend/internal/module/ops/infrastructure/notification_repository.go`
  - `code/backend/internal/module/ops/application/commands/notification_service.go`
  - `code/backend/internal/module/ops/api/http/notification_handler.go`
  - `code/backend/internal/app/router_routes.go`
- 前端：
  - `code/frontend/src/api/contracts.ts`
  - `code/frontend/src/api/admin.ts`
  - `code/frontend/src/components/notifications/AdminNotificationPublishDrawer.vue`
  - `code/frontend/src/composables/useAdminNotificationPublisher.ts`
  - `code/frontend/src/views/notifications/NotificationList.vue`

## 验证结果

- 后端：
  - `cd code/backend && go test ./internal/module/ops/... ./internal/app/...`
- 前端：
  - `cd code/frontend && npm test -- --run src/views/notifications/__tests__/NotificationList.test.ts src/composables/__tests__/useAdminNotificationPublisher.test.ts`
  - `cd code/frontend && npm test -- --run src/api/__tests__/admin.test.ts`
  - `cd code/frontend && npm run typecheck`
  - `cd code/frontend && npm run build`

## 文档影响说明

- 本文档从“执行计划”补充为“执行完成记录”，保留原始任务拆分以便追溯。
- 同步更新前端任务总览中的通知能力描述，避免仍把通知链路理解为仅列表/已读/WebSocket。

---

### Task 1: Add failing backend tests for publish-batch domain and admin API

**Files:**
- Modify: `code/backend/internal/module/ops/application/commands/notification_service_test.go`
- Modify: `code/backend/internal/module/ops/api/http/notification_http_integration_test.go`
- Modify: `code/backend/internal/app/full_router_state_matrix_integration_test.go`

- [x] **Step 1: Write failing command-service tests**
  Cover batch creation, audience union dedupe, recipient counting, and websocket fan-out for targeted recipients.
- [x] **Step 2: Run focused backend tests to verify RED**
  Run during implementation with targeted notification suites.
- [x] **Step 3: Write failing router/integration test**
  Cover `POST /api/v1/admin/notifications`, admin-only access, and validation failures.
- [x] **Step 4: Run router-focused tests to verify RED**
  Covered by final route and integration verification.

### Task 2: Add backend schema and repositories for publish batches

**Files:**
- Create: `code/backend/migrations/000042_create_notification_batches_table.up.sql`
- Create: `code/backend/migrations/000042_create_notification_batches_table.down.sql`
- Create: `code/backend/internal/model/notification_batch.go`
- Modify: `code/backend/internal/model/notification.go`
- Modify: `code/backend/internal/module/ops/ports/notification.go`
- Modify: `code/backend/internal/module/ops/infrastructure/notification_repository.go`

- [x] **Step 1: Add batch schema and supporting indexes**
  Include `batch_id`, batch metadata, audience JSON, and delivery-table indexes for inbox queries and batch joins.
- [x] **Step 2: Add repository contracts and failing compile/test checks**
  Run: `cd code/backend && go test ./internal/module/ops/...`
- [x] **Step 3: Implement repository persistence for batches and delivery records**
- [x] **Step 4: Re-run focused backend tests to verify GREEN**
  Run: `cd code/backend && go test ./internal/module/ops/...`

### Task 3: Implement backend publish service and admin route

**Files:**
- Modify: `code/backend/internal/dto/notification.go`
- Modify: `code/backend/internal/module/ops/application/commands/notification_service.go`
- Modify: `code/backend/internal/module/ops/api/http/notification_handler.go`
- Modify: `code/backend/internal/app/composition/ops_module.go`
- Modify: `code/backend/internal/app/router_routes.go`
- Modify: `code/backend/internal/app/router.go`

- [x] **Step 1: Add publish DTOs and handler contract**
  Model composable audience rules for `all`, `role`, `class`, and `user`.
- [x] **Step 2: Implement audience resolution with identity repository reuse**
  Resolve recipients from `users` by role/class/id, union them, and filter duplicates.
- [x] **Step 3: Implement admin publish endpoint and audit wiring**
  Add `POST /api/v1/admin/notifications` under admin-only routes.
- [x] **Step 4: Run focused backend notification tests**
  Run: `cd code/backend && go test ./internal/module/ops/... ./internal/app/... -run Notification`

### Task 4: Add failing frontend tests for admin publishing UI

**Files:**
- Modify: `code/frontend/src/views/notifications/__tests__/NotificationList.test.ts`
- Create: `code/frontend/src/composables/__tests__/useAdminNotificationPublisher.test.ts`

- [x] **Step 1: Write failing notification-page test**
  Cover admin-only publish trigger visibility and non-admin absence.
- [x] **Step 2: Write failing publish-flow test**
  Cover audience mode switching, class/user loading, payload assembly, and submit success path through the publish composable.
- [x] **Step 3: Run focused frontend tests to verify RED**
  Run during implementation with targeted notification-page and composable suites.

### Task 5: Implement frontend admin publish flow on `/notifications`

**Files:**
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/api/admin.ts`
- Create: `code/frontend/src/components/notifications/AdminNotificationPublishDrawer.vue`
- Create: `code/frontend/src/composables/useAdminNotificationPublisher.ts`
- Modify: `code/frontend/src/views/notifications/NotificationList.vue`

- [x] **Step 1: Add admin publish API types and request client**
- [x] **Step 2: Implement publish composable**
  Encapsulate drawer state, audience selections, remote option loading, and submit action.
- [x] **Step 3: Implement admin publish drawer component**
  Keep explicit props/emits and reuse existing page visual language.
- [x] **Step 4: Compose drawer into notifications page**
  Show admin-only trigger while keeping the user inbox experience unchanged.
- [x] **Step 5: Re-run focused frontend tests to verify GREEN**
  Run: `cd code/frontend && npm test -- --run src/views/notifications/__tests__/NotificationList.test.ts src/composables/__tests__/useAdminNotificationPublisher.test.ts`

### Task 6: Validate end-to-end behavior and sync docs

**Files:**
- Modify: `code/docs/tasks/b33-notification-merge-plan.md` (only if existing notification task notes should reflect the new publish model)

- [x] **Step 1: Run backend verification suite**
  Run: `cd code/backend && go test ./internal/module/ops/... ./internal/app/...`
- [x] **Step 2: Run frontend verification suite**
  Run:
  - `cd code/frontend && npm test -- --run src/views/notifications/__tests__/NotificationList.test.ts src/composables/__tests__/useAdminNotificationPublisher.test.ts`
  - `cd code/frontend && npm test -- --run src/api/__tests__/admin.test.ts`
  - `cd code/frontend && npm run typecheck`
  - `cd code/frontend && npm run build`
- [x] **Step 3: Review documentation impact**
  Update only the smallest relevant task/design note if the publish model changes documented behavior.
- [x] **Step 4: Request code review and fix findings**
- [x] **Step 5: Record final verification evidence before claiming completion**
