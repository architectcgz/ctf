# Admin Teaching API Owner Alignment 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-26-admin-teaching-api-owner-alignment-implementation-plan.md`
  - files reviewed：
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
- Classification check：同意本轮属于前端 P1 级结构收口，且边界限定在 admin teaching API owner 和直接依赖它的 platform/admin feature。
- Gate verdict：Pass

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- `api/teaching/*` 继续保留底层 HTTP contract owner；新增的 `api/admin/teaching.ts` 只承担 admin 语义 wrapper，不引入新的请求实现或分页逻辑分叉。
- `platform-class-management`、`platform-student-management`、`platform-instance-management` 与 `admin-notification-publisher` 已经从 `@/api/teaching` 摘下来，admin feature 不再直接引用 teacher 语义 API 入口。
- 这轮没有改 route view、页面模板、query contract 或 destroy flow owner，风险被限制在 import owner 与命名边界。
- backlog 第 2 条原文已按当前代码事实纠偏：`platform-users` 这只超大桶已经拆完，当前残留债务应继续归到“admin / teacher 结构耦合”。

## Required re-validation

- `npm run test:run -- src/views/platform/__tests__/ClassManage.test.ts src/views/platform/__tests__/StudentManage.test.ts src/views/platform/__tests__/InstanceManage.test.ts src/features/admin-notification-publisher/model/useAdminNotificationPublisher.test.ts`
- `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`
- `git diff --check -- code/frontend/src/api/admin/index.ts code/frontend/src/api/admin/teaching.ts code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.ts code/frontend/src/views/platform/__tests__/ClassManage.test.ts code/frontend/src/views/platform/__tests__/StudentManage.test.ts code/frontend/src/views/platform/__tests__/InstanceManage.test.ts code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.test.ts .harness/reuse-decisions/admin-teaching-api-owner-alignment.md docs/plan/impl-plan/2026-05-26-admin-teaching-api-owner-alignment-implementation-plan.md docs/reviews/frontend/2026-05-26-admin-teaching-api-owner-alignment-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Residual risk

- `TeacherClassItem`、`TeacherInstanceItem`、`getTeacherWriteupSubmissions` 这类历史命名仍然存在；这轮只把 admin feature 从 teacher API 入口上摘下来，还没处理 contract 层语义命名。
- `PlatformClassStudents.vue`、`PlatformStudentAnalysis.vue`、`PlatformStudentReviewArchive.vue` 这些 admin 页面仍在复用中立 workspace / widget，与 teacher 语义的更深层耦合还需要单独切片。

## Touched known-debt status

- 本轮 touched 的已知结构债是 admin feature 直接依赖 `@/api/teaching` 和 teacher 命名 API owner。
- 该债务在 touched surface 上已完成收口：platform class / student / instance feature 与 admin 通知发布都改为通过 `api/admin/teaching.ts` 取教学目录能力；剩余未收口部分转移到共享页面与 contract 命名层。
