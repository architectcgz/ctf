> 状态：Current
> 事实源：`StudentManagementPage.vue` 当前 owner、教师学生管理 route view 与 `teacher-student-management` feature public API 边界
> 替代：无

# Student Management Feature UI Migration Implementation Plan

## 目标

- 把 `StudentManagementPage.vue` 从 `components/teacher/student-management/` 迁到 `features/teacher-student-management/ui/`。
- 让 `views/teacher/TeacherStudentManagement.vue` 直接通过 `features/teacher-student-management` public API 组合 page-sized UI 与 `useStudentManagementPage()`。
- 收掉 `StudentManagementPage.vue` 对应的 `legacyComponentPageAllowlist` 例外，并同步更新教师端 raw-source 测试路径。

## 非目标

- 本轮不改 `useStudentManagementPage()` 的初始化、搜索、防抖、分页、跳转、导出班级报告或错误处理 owner。
- 本轮不新增 route hook，也不改 `TeacherStudentManagement.vue` 组合 `ClassReportExportDialog` 的方式。
- 本轮不改学生管理页的用户可见文案、筛选交互、表格字段和分页行为。

## 输入依据

- `docs/architecture/frontend/06-components.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `code/frontend/src/components/teacher/student-management/StudentManagementPage.vue`
- `code/frontend/src/views/teacher/TeacherStudentManagement.vue`
- `code/frontend/src/features/teacher-student-management/index.ts`
- `code/frontend/src/features/teacher-student-management/model/useStudentManagementPage.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `StudentManagementPage.vue` 现在只承担 page shell、展示数据消费和事件桥接，真正的业务 owner 已经在 `useStudentManagementPage.ts`。
- 这页本身不持有 `useRoute/useRouter`、API import 或 lifecycle owner，因此这轮不需要新增 route model。
- 当前残留结构债是它仍以 `components/*Page.vue` 的形式存在，并继续占用 `legacyComponentPageAllowlist`。

## 设计边界

### route view 继续负责

- 组合 `useStudentManagementPage()`、`StudentManagementPage` 与 `ClassReportExportDialog`
- 不直接持有初始化、搜索、防抖、分页、班级管理跳转或学员分析跳转 owner

### `features/teacher-student-management/model` 继续负责

- 班级目录加载、默认班级选择
- 搜索词 / 学号筛选、防抖加载、分页切换
- 班级管理跳转、学员分析跳转、班级报告导出弹窗状态

### `features/teacher-student-management/ui` 本轮负责

- 教师学生管理页 page-sized shell
- 消费上层派生状态与事件 handler
- 不直接持有 API、router 或对话框状态 owner

## 任务切片

### Slice 1：迁移学生管理 page shell 到 feature ui

- 目标：
  - 新增 `features/teacher-student-management/ui/StudentManagementPage.vue`
  - `features/teacher-student-management/index.ts` 导出 `ui`
  - `views/teacher/TeacherStudentManagement.vue` 改从 feature public API 引用
- 预期改动：
  - `code/frontend/src/features/teacher-student-management/index.ts`
  - `code/frontend/src/features/teacher-student-management/ui/*`
  - `code/frontend/src/views/teacher/TeacherStudentManagement.vue`
- 验证：
  - `npm run test:run -- src/views/teacher/__tests__/TeacherStudentManagement.test.ts`
- Review focus：
  - route view 是否继续只保留组合壳
  - page shell 是否仍未吸回业务 owner

### Slice 2：更新 guardrail、测试与 backlog

- 目标：
  - 更新 raw-source 测试路径与自动注册类型路径
  - 清理 `legacyComponentPageAllowlist` 中对应的 page 例外
  - 记录 backlog 进展
- 预期改动：
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/components.d.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherRootShellCleanup.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts`
  - `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
  - `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
  - `code/frontend/src/views/__tests__/sharedPaginationControls.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherStudentManagement.test.ts src/views/teacher/__tests__/teacherRootShellCleanup.test.ts src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts src/views/teacher/__tests__/teacherSurface.test.ts src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedPaginationControls.test.ts`
- Review focus：
  - allowlist 是否真实下降
  - 教师端 surface / shell 测试是否已经切到新 owner

## 结构收口检查

- `StudentManagementPage.vue` 不再作为 `components/*Page.vue` 存在。
- `views/teacher/TeacherStudentManagement.vue` 只保留 feature public API 组合壳。
- `features/teacher-student-management` public API 明确导出 `model + ui`。
- touched surface 上至少移除一条 `legacyComponentPageAllowlist`。

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherStudentManagement.test.ts src/views/teacher/__tests__/teacherRootShellCleanup.test.ts src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts src/views/teacher/__tests__/teacherSurface.test.ts src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedPaginationControls.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/student-management-feature-ui-migration.md docs/plan/impl-plan/2026-05-27-student-management-feature-ui-migration-implementation-plan.md docs/reviews/frontend/2026-05-27-student-management-feature-ui-migration-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/features/teacher-student-management code/frontend/src/views/teacher/TeacherStudentManagement.vue code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/components.d.ts code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts code/frontend/src/views/teacher/__tests__/teacherRootShellCleanup.test.ts code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts code/frontend/src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts code/frontend/src/views/__tests__/workspaceShellStyles.test.ts code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts code/frontend/src/views/__tests__/sharedPaginationControls.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review 关注点

- `features/teacher-student-management/ui` 是否成为学生管理页 page shell 的唯一 owner。
- `useStudentManagementPage()` 是否仍然保持初始化、搜索、防抖、分页与路由跳转 owner，不被迁移反向打散。
- 教师端 raw-source 测试与 allowlist 是否同步反映新边界。

## 回退 / 恢复说明

- 如迁移后出现问题，可把 `StudentManagementPage.vue` 移回 `components/teacher/student-management/` 并恢复 route view import。
- 如教师端 source 测试出现误报，可先回退测试路径切换，再单独重做 page shell 迁移。

## 残余风险

- `studentNoQuery` / `updateStudentNoQuery` 当前仍由 model 暴露但 page shell 未使用，本轮保持契约不动，避免把 page shell 迁移和交互语义收口混在一起。
- 其它 teacher legacy page 仍在 backlog 中，本轮不一并迁移。
