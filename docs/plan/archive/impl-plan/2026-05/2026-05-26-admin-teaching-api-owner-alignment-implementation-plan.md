> 状态：Current
> 事实源：`api/teaching/*`、`api/teacher/*`、`api/admin/*`、platform class/student/instance feature owner、相关 architecture review 与 backlog
> 替代：无

# Admin Teaching API Owner Alignment Implementation Plan

## 目标

- 让 admin 侧教学目录能力通过 `api/admin/*` owner 暴露，而不是继续让 `/platform/*` feature 直接依赖 `@/api/teaching` 和 teacher 语义函数名。
- 收口当前最直接的 admin / teacher 结构耦合：platform class、student、instance 管理，以及 admin 通知发布里的班级列表读取。
- 保持底层 HTTP contract 不变，只调整前端 API owner 与 feature 引用边界。

## 非目标

- 本轮不改后端接口、URL、请求参数、分页 contract 或响应结构。
- 本轮不重命名 `api/contracts.ts` 里的 `TeacherClassItem`、`TeacherInstanceItem` 等历史类型名。
- 本轮不改 `PlatformClassStudents.vue`、`PlatformStudentAnalysis.vue`、`ReviewArchiveWorkspace` 这类更深的 admin / teacher 共享页面 owner。
- 本轮不处理 `request.ts` 的全局错误导航 owner。

## 输入依据

- `code/frontend/src/api/teaching/classes.ts`
- `code/frontend/src/api/teaching/instances.ts`
- `code/frontend/src/api/teacher/classes.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/frontend/src/api/admin/index.ts`
- `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.ts`
- `code/frontend/src/views/platform/__tests__/ClassManage.test.ts`
- `code/frontend/src/views/platform/__tests__/StudentManage.test.ts`
- `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- `code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.test.ts`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 旧的 `platform-users` 超大桶已经拆掉，当前 backlog 第 2 条的原始表述滞后于代码事实。
- 真实残余债务不是“platform-users 还在混类/学生/实例”，而是 admin 侧 feature 仍直接引用 `@/api/teaching`、`getTeacherInstances`、`destroyTeacherInstance` 和 `Teacher*` 语义。
- `api/teacher/*` 已经存在薄 wrapper owner，因此新增 `api/admin/teaching.ts` 承接 admin 侧教学目录能力，是当前最小、最可审查的收口路径。

## 任务切片

### Slice 1：新增 admin teaching wrapper owner

- 目标：
  - 在 `api/admin/` 下新增教学目录能力入口，承接 admin 班级、学生目录、实例目录和实例销毁的语义 owner。
- 预期改动：
  - `code/frontend/src/api/admin/index.ts`
  - `code/frontend/src/api/admin/teaching.ts`
- 说明：
  - 只做薄 wrapper / re-export，不改底层 `api/teaching/*` 实现。

### Slice 2：替换 platform feature 的 teacher 语义 API 依赖

- 目标：
  - 让 `platform-class-management`、`platform-student-management`、`platform-instance-management` 与 `admin-notification-publisher` 从 `api/admin/teaching.ts` 取能力，不再直接引用 `@/api/teaching`。
- 预期改动：
  - `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
  - `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
  - `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
  - `code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.ts`
- review focus：
  - page owner 不变
  - query / pagination / destroy owner 不回流到 view
  - 只是 API owner 调整，不改请求语义

### Slice 3：同步测试与 backlog / review

- 目标：
  - 让 platform 相关测试与 admin 通知发布测试断言新的 API owner。
  - 把 backlog 第 2 条改写成当前事实，避免后续继续跟着过时描述做错方向。
- 预期改动：
  - `code/frontend/src/views/platform/__tests__/ClassManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/StudentManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
  - `code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-26-admin-teaching-api-owner-alignment-review.md`

## 验证

- `npm run test:run -- src/views/platform/__tests__/ClassManage.test.ts src/views/platform/__tests__/StudentManage.test.ts src/views/platform/__tests__/InstanceManage.test.ts src/features/admin-notification-publisher/model/useAdminNotificationPublisher.test.ts`
- `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`

## 回退 / 恢复说明

- 这轮回退粒度按 API owner 切：可以整体回退 `api/admin/teaching.ts` 和四个 platform/admin feature 的 import 切换，不涉及数据迁移或 UI 结构回退。

## 残余风险

- 这轮只能先把 admin 侧教学目录 owner 从 `api/teaching` 上摘下来，不能一次性消除所有 admin / teacher 共享页面与 contract 命名历史。
- `TeacherClassItem`、`TeacherInstanceItem` 等 contract 名称仍然带 teacher 前缀；如果后续要彻底去 teacher 语义，需要单独评估类型别名或 contract 重命名成本。
