# Reuse Decision

## Change type
frontend refactor / feature router helper cleanup

## Existing code searched
- code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts
- code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts
- code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts
- code/frontend/src/features/student-analysis-workspace/model/__tests__/useStudentAnalysisReviewQuerySync.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `useStudentAnalysisPage.ts` 已经是 student analysis route 的 page owner，天然持有 `useRoute()` 与 `useRouter()`。
- `useStudentAnalysisNavigation.ts` 和 `useStudentAnalysisReviewQuerySync.ts` 当前都只是 page 下游 helper，更适合消费 callback / route-like contract，而不是直接依赖 `vue-router` 类型。

## Decision
refactor_existing

## Reason
`featureRouterImportAllowlist` 中，`features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts -> vue-router` 与 `features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts -> vue-router` 都不是合理长期例外。它们当前不是 page owner，只是导航 helper 和 review query sync helper。

最小正确改动是：

- 让 `useStudentAnalysisNavigation()` 改为消费显式导航 callback，而不是 `Router` 类型
- 让 `useStudentAnalysisReviewQuerySync()` 改为消费本地 `route-like` / `replaceQuery` contract，而不是 `vue-router` 类型
- 保留 `useStudentAnalysisPage.ts` 作为唯一 route-aware page owner
- 删除对应 allowlist 条目，并补 raw-source 护栏

本轮不做：

- 不重构 `useStudentAnalysisPage.ts` 本身的 page owner 身份
- 不改 student analysis 的 UI、review workspace 数据流或教学 API
- 不处理 `featureRouterImportAllowlist` 其它剩余条目

## Files to modify
- .harness/reuse-decisions/student-analysis-router-helper-cleanup.md
- docs/plan/impl-plan/2026-05-29-student-analysis-router-helper-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-student-analysis-router-helper-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts
- code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts
- code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts
- code/frontend/src/features/student-analysis-workspace/model/__tests__/useStudentAnalysisReviewQuerySync.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts

## After implementation
- `useStudentAnalysisNavigation.ts` 不再 import `vue-router`
- `useStudentAnalysisReviewQuerySync.ts` 不再 import `vue-router`
- student analysis 这组 route/query owner 明确回到 `useStudentAnalysisPage.ts`
- `featureRouterImportAllowlist` 缩小两条
