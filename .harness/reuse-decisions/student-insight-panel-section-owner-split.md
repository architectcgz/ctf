# Reuse Decision

## Change type
frontend refactor / section owner split

## Existing code searched
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/index.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/studentInsightShared.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `student-analysis-review` 已经承接 review / writeup / manual review / evidence 相关 workflow 和 UI owner。
- `student-analysis-workspace` 已经承接 page shell、overview hero、tab shell 与 primary sections 组合。

## Decision
refactor_existing

## Reason
当前 `StudentInsightPanel.vue` 仍混放两类 section owner：

- workspace 自己的 primary sections：overview / recommendations / timeline
- review feature 的 review sections：writeups / manual review / evidence

最小正确收口方式是：

- 在 `student-analysis-workspace/ui` 内新增 primary sections 组合组件，承接 overview / recommendations / timeline
- 在 `student-analysis-review/ui` 内新增 review sections 组合组件，承接 writeups / manual review / evidence
- 让 `StudentInsightPanel.vue` 退回 empty/loading shell 与两组区块的桥接 owner

本轮不做：

- 不继续改 `StudentAnalysisWorkspaceContent.vue` 或 `useStudentAnalysisPage.ts`
- 不把 `StudentInsightSection` 类型迁出当前 review helper owner；这会引入新的共享 landing zone 设计，超出本刀范围
- 不拆 `StudentInsightOverviewSection.vue` 或 `StudentInsightRecommendationsSection.vue` 内部模板

## Files to modify
- `.harness/reuse-decisions/student-insight-panel-section-owner-split.md`
- `docs/plan/impl-plan/2026-06-01-student-insight-panel-section-owner-split-plan.md`
- `docs/reviews/frontend/2026-06-01-student-insight-panel-section-owner-split-review.md`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/index.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightReviewSections.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/index.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `StudentInsightPanel.vue` 会退回 shell owner，不再直接内联 primary / review 各 section。
- `student-analysis-workspace` 会只承接 primary sections 组合。
- `student-analysis-review` 会承接 review sections 组合，而不是只承接 review 内部单块 section。
