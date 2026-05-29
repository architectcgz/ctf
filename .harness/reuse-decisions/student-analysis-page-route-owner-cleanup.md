# Reuse Decision

## Change type
frontend refactor / student analysis page route owner cleanup

## Existing code searched
- code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts
- code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts
- code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts
- code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts
- code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts
- code/frontend/src/composables/routeQueryTransport.ts
- code/frontend/src/composables/routeNavigationTransport.ts
- code/frontend/src/features/student-dashboard/model/studentDashboardRoutes.ts
- code/frontend/src/features/student-review-archive-workspace/model/studentReviewArchiveRoutes.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `composables/routeQueryTransport.ts`
- `composables/routeNavigationTransport.ts`
- `features/student-review-archive-workspace/model/studentReviewArchiveRoutes.ts`
- `features/student-dashboard/model/studentDashboardRoutes.ts`

## Decision
refactor_existing

## Reason
`useStudentAnalysisPage.ts` 现在直接依赖 `vue-router`，但它实际碰 router 的职责面可以拆得更清楚：

- 读取 `className / studentId` params 与 review workspace query
- 写回 `reviewMode / reviewResult / reviewChallengeId`
- 打开班级学生页、题目详情和复盘归档 3 条薄导航

这条如果继续把 `vue-router` 留在 page model，`featureRouterImportAllowlist` 不会下降；但如果为它再包一层 route wrapper，也只是新增中间态。更小的收口方式是：

- route `params / query` 读侧下沉到共享 query transport
- `push / replace` 下沉到共享 navigation transport
- 班级学生页 / 题目详情 / 复盘归档 3 条薄导航落到本地 `studentAnalysisRoutes.ts`
- review workspace query owner、student analysis 数据加载和 breadcrumb owner 继续留在 page model

## Files to modify
- .harness/reuse-decisions/student-analysis-page-route-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-student-analysis-page-route-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-student-analysis-page-route-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/student-analysis-workspace/model/studentAnalysisRoutes.ts
- code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts
- code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts
- code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts
- code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts

## After implementation
- `useStudentAnalysisPage.ts` 不再 import `vue-router`
- review workspace query owner 和 student analysis page workflow 继续留在 page model
- 班级学生页 / 题目详情 / 复盘归档走本地 route target helper + shared navigation transport
- `featureRouterImportAllowlist` 再收掉 `features/student-analysis-workspace/model/useStudentAnalysisPage.ts -> vue-router`
