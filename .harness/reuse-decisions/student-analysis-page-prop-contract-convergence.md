# Reuse Decision

## Change type
component / contract

## Existing code searched
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue`
- `code/frontend/src/views/teacher/TeacherClassStudents.vue`
- `code/frontend/src/views/platform/PlatformClassStudents.vue`

## Decision
refactor_existing

## Reason
- `StudentAnalysisPage.vue` 当前模板只消费 `selectedStudent`、`loadingDetails`、`error`、overview hero 数据和 `StudentInsightPanel` 相关 workflow 数据；`classes`、`students`、`selectedClassName`、`selectedStudentId`、`loadingClasses`、`loadingStudents` 都只剩 props 声明。
- `TeacherStudentAnalysis.vue` 与 `PlatformStudentAnalysis.vue` 还在持续传这些无效 props，并监听 `openClassManagement`、`selectClass`、`selectStudent` 三个当前页面不会发出的事件，说明 contract 已经漂移。
- `useStudentAnalysisPage.ts` 里原本还保留 `getClasses()` 初始化前置，但当前页面壳已经没有任何 class list 消费方；继续保留只会让无关接口失败扩大成整页错误态。
- 仓库里现有共享 workspace page 的最小正确做法，是直接收紧现有组件和 route view 契约，而不是再造桥接 wrapper 或保留“可能将来会用”的死参数。

## Files to modify
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`

## After implementation
- 如果 `student-analysis-workspace` 的共享 page contract 继续收紧，可以再检查 `components/class-management` 和 `features/student-analysis-workspace` 的公共导出是否还有只为历史桥接保留的无效符号。
