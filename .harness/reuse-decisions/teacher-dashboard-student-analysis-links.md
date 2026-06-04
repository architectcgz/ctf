# Reuse Decision

## Change type
page / component / composition

## Existing code searched
- code/frontend/src/features/teacher/dashboard/model/useDashboardPage.ts
- code/frontend/src/features/teacher/dashboard/model/useDashboardMetrics.ts
- code/frontend/src/features/teacher/dashboard/model/teacherDashboardInsightBuilders.ts
- code/frontend/src/features/teacher/dashboard/model/teacherDashboardOverviewBuilders.ts
- code/frontend/src/features/teacher/dashboard/model/teacherDashboardRoutes.ts
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardStudentInsightPanel.vue
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardReviewPanel.vue
- code/frontend/src/features/teacher/student-management/model/teacherStudentManagementRoutes.ts
- code/frontend/src/shared/ui/navigation/AppRouteLink.vue
- code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts

## Similar implementations found
- `teacherStudentManagementRoutes.ts` 已有教师端学生详情 route target builder，可直接复用到 dashboard modal 列表。
- `TeacherDashboardPage.vue` 已经通过 `classManagementRoute` + `AppRouteLink` 消费 route target，适合继续让 page owner 统一编排面板交互。
- `ClassReportExportDialog.vue` 已示范在 feature 内使用 `AdminSurfaceModal` 承接教师工作台里的弹窗交互。

## Decision
extend_existing

## Reason
这次是 dashboard 新增“查看学生信息”的行为能力，但交互要分流：
- 风险组 / 头部样本 / 薄弱维度：先看学生名单 modal，再进入个人复盘。
- 班级相关条目：直接复用现有 class review 页面。

最小正确改动是复用现有 `TeacherStudentAnalysis` 与 `TeacherClassReview` route target 模式，在 dashboard model 内把 overview 聚合结果收口成“学生名单或班级复盘入口”，再由 page owner 统一编排 modal。

## Files to modify
- .harness/reuse-decisions/teacher-dashboard-student-analysis-links.md
- docs/plan/impl-plan/2026-06-04-teacher-dashboard-student-analysis-links-implementation-plan.md
- code/frontend/src/features/teacher/dashboard/model/teacherDashboardRoutes.ts
- code/frontend/src/features/teacher/dashboard/model/teacherDashboardInsightBuilders.ts
- code/frontend/src/features/teacher/dashboard/model/teacherDashboardOverviewBuilders.ts
- code/frontend/src/features/teacher/dashboard/model/useDashboardMetrics.ts
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardStudentListDialog.vue
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardStudentInsightPanel.vue
- code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardReviewPanel.vue
- code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts

## After implementation
- `学生洞察` 的风险组等学生粒度条目会先打开学生名单 modal。
- `教学复盘` 与班级粒度条目会直接进入 class review 页面。
- modal 内每个学生再通过现有 route target 进入学生详情。
- route target owner 继续留在 dashboard feature/model，panel 不直接 import `vue-router`。
- 对于 overview DTO 无法可靠映射学生名单的摘要条目，不伪造弹窗内容或跳转入口。
