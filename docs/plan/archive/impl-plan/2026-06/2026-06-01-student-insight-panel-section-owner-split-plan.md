# Student Insight Panel Section Owner Split 计划

## Objective

- 让 `StudentInsightPanel.vue` 退回 shell owner，不再直接内联 primary sections 和 review sections 的双重装配。
- 继续按 feature owner 收口 `student-analysis-workspace` 与 `student-analysis-review` 的 section 组合边界。

## Non-goals

- 不改 `StudentAnalysisWorkspaceContent.vue` 的 page-level content owner。
- 不修改 `useStudentAnalysisPage.ts`、`useReviewWorkspace.ts` 或 `useSubmissionReviewFlows.ts` 的 workflow。
- 不为 `StudentInsightSection` 新建新的共享 landing zone。

## Source Inputs

- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspaceContent.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/index.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/studentInsightShared.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- 当前最小有效切口不是继续减 `StudentInsightPanel.vue` 的行数，而是把 primary / review 两组 section assembly 分别交给正确 feature。
- `StudentInsightSection` 类型 owner 目前仍有些偏，但如果本刀同时改 landing zone，会扩大为新的跨-feature 语义设计；先收 section 组合 owner 更稳。

## Task Slices

### Slice 1: 抽离 workspace primary sections owner

- 目标：新增 `StudentInsightPrimarySections.vue`，承接 overview / recommendations / timeline。
- 风险：overview section 仍依赖 loading/student shell 外层判定，不能把 empty/loading 状态下沉进去。

### Slice 2: 抽离 review sections owner

- 目标：新增 `StudentInsightReviewSections.vue`，承接 writeups / manual review / evidence。
- 风险：manual review 不是顶层 tab，但仍可能通过 `activeSection` 进入，因此 review group 内部要保留这条可见性逻辑。

### Slice 3: 收口 panel shell 与边界测试

- 目标：让 `StudentInsightPanel.vue` 只保留 empty/loading shell 和对子组件事件桥接，并同步 teacher source-boundary 护栏与 backlog 进展。
- 风险：如果测试不更新，后续很难防止 primary/review 组合再次回流到 panel shell。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision student-insight-panel-section-owner-split`
- `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/student-insight-panel-section-owner-split.md docs/plan/archive/impl-plan/2026-06/2026-06-01-student-insight-panel-section-owner-split-plan.md docs/reviews/frontend/2026-06-01-student-insight-panel-section-owner-split-review.md code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue code/frontend/src/features/teaching/student-analysis-workspace/ui/index.ts code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightReviewSections.vue code/frontend/src/features/teaching/student-analysis-review/ui/index.ts code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review Focus

- `StudentInsightPanel.vue` 是否真的退回 shell owner，而不是继续混放两组 section 装配。
- `StudentInsightPrimarySections.vue` 是否只承接 workspace primary sections，而不反向依赖 review feature internals。
- `StudentInsightReviewSections.vue` 是否成为 review feature 的组合 owner，而不是只做路径中转。
