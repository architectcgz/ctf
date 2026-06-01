# Reuse Decision

## Change type
frontend refactor / shared student analysis page owner cleanup

## Existing code searched
- `code/frontend/src/pages/teacher/TeacherStudentAnalysisRoutePage.vue`
- `code/frontend/src/pages/platform/PlatformStudentAnalysisRoutePage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/index.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `TeacherClassStudentsRoutePage.vue` 已经在上一轮退回薄壳，由 feature 内部 `TeacherClassStudentsPage.vue` 组合 page shell 与导出对话框。
- `TeacherClassManagementPage.vue`、`PlatformInstanceManagementPage.vue` 等最近目录页迁移都在让 route page 退回薄壳，feature page 自己组合 page model 和内部 workspace。
- 当前 `TeacherStudentAnalysisRoutePage.vue` 与 `PlatformStudentAnalysisRoutePage.vue` 基本是同一份模板，只在根节点 class 上有角色差异，但都还直接组合 `StudentAnalysisPage`、`ClassReportExportDialog` 和 `useStudentAnalysisPage()`。

## Decision
refactor_existing

## Reason
下一轮最小正确切片不是继续拆 student analysis workflow，而是先把这一组共享 route owner 收口到 feature：

- 在 `features/teaching/student-analysis-workspace/ui` 下新增共享 `StudentAnalysisWorkspacePage.vue`，由它统一组合 page model、`StudentAnalysisPage` 和 `ClassReportExportDialog`。
- `TeacherStudentAnalysisRoutePage.vue` 与 `PlatformStudentAnalysisRoutePage.vue` 退回只渲染 feature page 的薄壳，只保留根类名差异。
- 保留共享 `useStudentAnalysisPage()` 命名，不把同时被 teacher / platform 消费的 page model 误改成单角色 owner。
- 不移动文件路径，不改 student analysis 的 tab、review workspace、writeup moderation、review archive export 或导出默认班级逻辑。

## Files to modify
- `.harness/reuse-decisions/student-analysis-page-owner-cleanup.md`
- `docs/plan/impl-plan/2026-06-01-student-analysis-page-owner-cleanup-plan.md`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspacePage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/index.ts`
- `code/frontend/src/pages/teacher/TeacherStudentAnalysisRoutePage.vue`
- `code/frontend/src/pages/platform/PlatformStudentAnalysisRoutePage.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- student analysis 的 page shell owner 会从 teacher / platform route page 收回共享 feature。
- teacher / platform route page 只保留 route-level 渲染职责，不再直接耦合导出对话框或 page model。
- 共享 `useStudentAnalysisPage()` 继续保持中性命名，避免和多角色 consumer 冲突。
