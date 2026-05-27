> 状态：Current
> 事实源：`api/contracts.ts`、classes API client、class/student/instance workspace、fronted backlog
> 替代：无

# Class Directory Contract Naming Neutralization Implementation Plan

## 目标

- 把共享班级目录 DTO `TeacherClassItem` 收口成中性命名 `ClassDirectoryItem`。
- 保持 teacher / platform 班级目录、学生目录、实例目录和 admin 通知发布里的班级选项行为不变，只收 contract 命名语义。

## 非目标

- 本轮不改 `TeacherInstanceItem`、`TeacherStudentItem`、`TeacherClassSummaryData` 等其他 class / instance 相关 DTO。
- 本轮不改 `TeacherAWDReviewContestItemData`、`TeacherAttackSessionQuery`。
- 本轮不调整 `getClasses()` 的 HTTP path、分页 contract 或 admin / teacher API owner。

## 输入依据

- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teacher/classes.ts`
- `code/frontend/src/api/teaching/classes.ts`
- `code/frontend/src/api/admin/teaching.ts`
- `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/teacher-class-management/model/useClassManagementPage.ts`
- `code/frontend/src/features/teacher-student-management/model/useStudentManagementPage.ts`
- `code/frontend/src/features/teacher-instances/model/useInstances.ts`
- `code/frontend/src/features/teacher-workspace/model/useWorkspace.ts`
- `code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.ts`
- `code/frontend/src/components/teacher/class-management/ClassManagementPage.vue`
- `code/frontend/src/components/teacher/student-management/StudentManagementPage.vue`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue`
- `code/frontend/src/components/platform/student/StudentManageWorkspacePanel.vue`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `TeacherClassItem` 当前被 teacher、platform 和 admin 侧多个目录 / 选项流共享消费，继续保留 teacher 前缀已经与真实 owner 语义不一致。
- 这组 DTO 只表达班级目录项，比 `TeacherInstanceItem`、`TeacherStudentItem` 的行为语义更薄，适合作为独立命名收口切片。
- 最小安全切片是只收 `TeacherClassItem -> ClassDirectoryItem`，同步 API client、feature、组件、测试与事实文档。

## 任务切片

### Slice 1：收口共享 class directory contract 名称

- 目标：
  - 在 `api/contracts.ts` 提供中性 `ClassDirectoryItem`，同步 classes API client 返回类型。
- 预期改动：
  - `code/frontend/src/api/contracts.ts`
  - `code/frontend/src/api/teacher/classes.ts`
  - `code/frontend/src/api/teaching/classes.ts`
  - `code/frontend/src/api/admin/teaching.ts`
- review focus：
  - `getClasses()` 的返回值、分页分支和现有 admin wrapper 行为不变

### Slice 2：同步 class / student / instance / notification 消费面

- 目标：
  - 让所有共享班级目录消费面切到 `ClassDirectoryItem`。
- 预期改动：
  - `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
  - `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
  - `code/frontend/src/features/teacher-class-management/model/useClassManagementPage.ts`
  - `code/frontend/src/features/teacher-student-management/model/useStudentManagementPage.ts`
  - `code/frontend/src/features/teacher-instances/model/useInstances.ts`
  - `code/frontend/src/features/teacher-workspace/model/useWorkspace.ts`
  - `code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.ts`
  - `code/frontend/src/components/teacher/class-management/ClassManagementPage.vue`
  - `code/frontend/src/components/teacher/student-management/StudentManagementPage.vue`
  - `code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue`
  - `code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue`
  - `code/frontend/src/components/platform/student/StudentManageWorkspacePanel.vue`
  - `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- review focus：
  - 班级筛选、班级目录表格、实例班级选项和通知发布班级选项不回归

### Slice 3：同步 contract 文档、backlog 与 review 证据

- 目标：
  - 更新 contract 文档和 backlog，对齐本轮命名收口后的剩余项。
- 预期改动：
  - `docs/contracts/api-contract-v1.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-class-directory-contract-naming-neutralization-review.md`

## 验证

- `npm run test:run -- src/api/__tests__/teacher.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`

## 回退 / 恢复说明

- 本轮回退按命名切：整体撤销 `ClassDirectoryItem` 及其消费面引用更新即可，不涉及运行时 API owner 或行为回退。

## 残余风险

- `TeacherAWDReviewContestItemData`、`TeacherAttackSessionQuery` 等其他共享 contract 命名残余还在。
