# Student Analysis Review UI Owner Cleanup 计划

## Objective

- 把 student analysis 中的 review / writeup / manual review / evidence 相关 UI 从 `student-analysis-workspace` 收回 `student-analysis-review`。
- 保持 `student-analysis-workspace` 只承接 page shell、tab 组合和学员洞察总体装配。

## Non-goals

- 不改 `useStudentAnalysisPage.ts`、`useReviewWorkspace.ts`、`useSubmissionReviewFlows.ts` 的运行逻辑。
- 不处理 `StudentInsightOverviewSection.vue`、`StudentInsightRecommendationsSection.vue` 的 entity 化。
- 不继续拆 `StudentAnalysisPage.vue` 或 `StudentAnalysisWorkspacePage.vue`。

## Source Inputs

- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightWriteupsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightManualReviewSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentReviewWorkspace.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/studentInsightShared.ts`
- `code/frontend/src/features/teaching/student-analysis-review/index.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/index.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- `student-analysis-review` 已经是 review workspace / writeup moderation / review archive export 的 workflow owner，让这组 UI 继续挂在 workspace feature 下会制造错位。
- 最小切片是收 UI owner，不顺带碰 model 契约和 route owner。

## Task Slices

### Slice 1: 收回 review UI 文件落点

- 目标：把 review / writeup / manual review / evidence 相关 UI 和 presentation helper 移到 `student-analysis-review/ui`。
- 风险：移动后相对 import 和邻近测试 owner 需要一起收口。

### Slice 2: 让 workspace 只消费 review feature public API

- 目标：`StudentInsightPanel.vue` 不再引用 workspace 内旧路径，而是通过 `@/features/teaching/student-analysis-review` 组合 review UI。
- 风险：如果只换文件路径，不补 public API，会留下新的 cross-feature internal import。

### Slice 3: 更新护栏与 backlog

- 目标：更新 teacher student analysis raw-source 断言、邻近 component tests 和 backlog 进展。
- 风险：如果测试仍盯旧 feature 路径，后续 owner 回流时难以及时发现。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision student-analysis-review-ui-owner-cleanup`
- `cd code/frontend && npm run test:run -- src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.test.ts src/features/teaching/student-analysis-review/ui/studentReviewWorkspacePresentation.test.ts src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/student-analysis-review-ui-owner-cleanup.md docs/plan/impl-plan/2026-06-01-student-analysis-review-ui-owner-cleanup-plan.md docs/reviews/frontend/2026-06-01-student-analysis-review-ui-owner-cleanup-review.md code/frontend/src/features/teaching/student-analysis-review/ui/index.ts code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue code/frontend/src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.vue code/frontend/src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.test.ts code/frontend/src/features/teaching/student-analysis-review/ui/studentInsightShared.ts code/frontend/src/features/teaching/student-analysis-review/ui/studentReviewWorkspacePresentation.test.ts code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue code/frontend/src/features/teaching/student-analysis-workspace/ui/index.ts code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review Focus

- `student-analysis-review` 是否成为 review / writeup / manual review / evidence 这组 UI 的唯一 feature owner。
- `StudentInsightPanel.vue` 是否只通过 review feature public API 消费这组区块。
- 邻近测试 owner 是否同步迁移，避免留下“代码已迁、测试仍在旧 feature”的半收口状态。
