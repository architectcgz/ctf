# Reuse Decision

## Change type
api / composition / feature / page

## Existing code searched
- `code/frontend/src/api/teaching.ts`
- `code/frontend/src/api/teaching/classes.ts`
- `code/frontend/src/api/teaching/instances.ts`
- `code/frontend/src/api/teacher/classes.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/frontend/src/api/admin/index.ts`
- `code/frontend/src/api/admin/platform.ts`
- `code/frontend/src/features/platform-user-management/model/usePlatformUsers.ts`
- `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/student-directory/model/useStudentDirectoryQuery.ts`
- `code/frontend/src/features/teacher-student-management/model/useStudentManagementPage.ts`
- `code/frontend/src/features/teacher-instances/model/useInstances.ts`
- `code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.ts`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `api/teacher/*.ts` 已经证明仓库接受“语义命名的薄 wrapper 重新导出底层 teaching 实现”的 API owner 模式。
- `platform-user-management` 已经从旧 `platform-users` 超大桶里独立出来，说明当前前端结构收口策略是：先按 owner 切 feature，再逐步把上层引用从旧语义迁走。
- `student-directory` 已经作为中立 query owner 同时服务 teacher / platform 场景，说明这轮不需要再新造一层 page owner，只需要把 platform 入口从 teacher 语义 API 上摘下来。

## Decision
extend_existing

## Reason
- 这次不是新增接口，也不是重写 teacher / platform 页面，而是把 admin 教学目录能力的 API owner 收口到更准确的位置。
- `api/teaching/*` 仍是实际 HTTP contract owner，可以继续保留；但 `/platform/*` feature 直接依赖 `@/api/teaching`、`getTeacherInstances`、`TeacherClassItem`、`TeacherInstanceItem` 这类 teacher 语义，会继续放大 admin / teacher 结构耦合。
- 最小正确方案是扩展现有 `api/admin/*` 入口，新增 admin 侧 teaching directory wrapper，并把 platform class / student / instance feature 与相关测试切到新的 admin owner；不需要再引入新的 route view、页面组件或重复 query composable。

## Files to modify
- `.harness/reuse-decisions/admin-teaching-api-owner-alignment.md`
- `docs/plan/impl-plan/2026-05-26-admin-teaching-api-owner-alignment-implementation-plan.md`
- `code/frontend/src/api/admin/index.ts`
- `code/frontend/src/api/admin/teaching.ts`
- `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.ts`
- `code/frontend/src/views/platform/__tests__/ClassManage.test.ts`
- `code/frontend/src/views/platform/__tests__/StudentManage.test.ts`
- `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- `code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-26-admin-teaching-api-owner-alignment-review.md`

## After implementation
- 如果这层 owner 收口稳定，可以把 “teacher / admin 语义通过 `api/<role>/...` 薄 wrapper 隔离，而不是让 feature 直接引用 `api/teaching`” 记入本地 `.harness/reuse-index/`，作为后续 AWD review、student analysis、review archive 等 admin/teacher 共享能力的复用线索。
