# Student Analysis Router Helper Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-student-analysis-router-helper-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/student-analysis-router-helper-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-student-analysis-router-helper-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-student-analysis-router-helper-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts`
  - `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts`
  - `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
  - `code/frontend/src/features/student-analysis-workspace/model/__tests__/useStudentAnalysisReviewQuerySync.test.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- Classification check：预期按单一 feature 下游 helper cleanup 处理；`useStudentAnalysisNavigation.ts` 与 `useStudentAnalysisReviewQuerySync.ts` 都不属于合理的 route-aware helper owner。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useStudentAnalysisPage.ts` 继续持有 `useRoute()` / `useRouter()` 合理，因为它本来就是 page owner。
- `useStudentAnalysisNavigation.ts` 与 `useStudentAnalysisReviewQuerySync.ts` 应该只保留 helper contract，不应继续 import `vue-router`。
- 这两条 allowlist 如果继续保留，会给 page 下游 helper 混入 router 开后门；本轮应当在 touched surface 内收掉。
- 当前实现已经把具体导航动作与 review query 写回统一回收到 `useStudentAnalysisPage.ts`，helper 只保留局部 contract 与状态判断，TeacherStudentAnalysis / helper 单测都维持通过。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/features/student-analysis-workspace/model/__tests__/useStudentAnalysisReviewQuerySync.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/student-analysis-router-helper-cleanup.md docs/plan/impl-plan/2026-05-29-student-analysis-router-helper-cleanup-plan.md docs/reviews/frontend/2026-05-29-student-analysis-router-helper-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts code/frontend/src/features/student-analysis-workspace/model/__tests__/useStudentAnalysisReviewQuerySync.test.ts code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `featureRouterImportAllowlist` 还有大量条目，需要后续继续按“page owner / non-page owner”逐条判定。
