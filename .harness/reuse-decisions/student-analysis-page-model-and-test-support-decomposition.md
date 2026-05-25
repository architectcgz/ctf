# Reuse Decision

## Change type
refactor / test-structure

## Existing code searched
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `docs/plan/impl-plan/2026-05-24-platform-route-owner-decoupling-implementation-plan.md`
- `docs/reviews/frontend/2026-05-25-student-analysis-page-prop-contract-convergence-review.md`

## Similar implementations found
- `code/frontend/src/features/teacher-student-analysis/model/useReviewWorkspace.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useSubmissionReviewFlows.ts`
- `code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts`

## Decision
refactor_existing

## Reason
- `useStudentAnalysisPage.ts` 已经不是“缺功能”，而是 page owner 混了过多稳定职责；应该沿着现有 feature 目录继续拆成本地 helper，而不是新建平行 feature。
- review workspace query 同步和主数据加载都已经有清晰输入/输出，适合从 page model 中抽出，不需要改变对外 public API。
- teacher/platform 学员分析测试重复的是 route mock、API mock 和基础 fixture，不是业务断言本身；应抽共享 support，保留各自 test file 承接独有风险。

## Files to modify
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisDataState.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts`
- `code/frontend/src/features/student-analysis-workspace/model/index.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `code/frontend/src/views/__tests__/studentAnalysisRouteTestSupport.ts`
- `docs/plan/impl-plan/2026-05-25-student-analysis-page-model-and-test-support-decomposition-implementation-plan.md`
- `.harness/reuse-decisions/student-analysis-page-model-and-test-support-decomposition.md`

## After implementation
- 如果这次拆完后 `useStudentAnalysisPage.ts` 仍继续累积新的 review / export / routing owner，可以再按 capability 把剩余流程继续下切，但那会是下一轮独立切片，不在本轮混做。
