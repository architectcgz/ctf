# Reuse Decision

## Change type
frontend refactor / component contract / style owner

## Existing code searched
- code/frontend/src/shared/ui/common/SectionCard.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspaceContent.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue
- code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue
- code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue
- code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue
- code/frontend/src/widgets/review-archive-workspace/ReviewArchiveWorkspace.vue
- code/frontend/src/widgets/review-archive-workspace/ReviewArchiveSummarySection.vue
- code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts
- code/frontend/scripts/vue-deep-allowlist.json

## Similar implementations found
- `SectionCard.vue` 已经是 student analysis、review archive、topology studio 共享的结构壳，但当前只暴露 `title / subtitle`，没有显式样式 contract。
- student analysis 与 review archive 当前都在父级通过 `:deep(.section-card*)` 反向改 `SectionCard` 内部 header / body / border / surface。
- `AppCard.vue` 已经采用“组件内部稳定结构 + variant / CSS variable contract”模式，这条路径比继续在 consumer 里写深度选择器更符合现有 shared UI owner。

## Decision
extend_existing

## Reason
这轮不是给每个页面各写一套本地样式，而是把 `SectionCard` 的共享视觉差异收口成显式 contract。

最小正确改动是：

- 扩展 `SectionCard` 的变体 / CSS variable contract
- 让 student analysis 与 review archive 的 consumer 直接声明自己需要的 section card 变体
- 删除这条链上的父级 `:deep(.section-card*)` 覆盖
- 同步更新原始源码断言和 `:deep` allowlist

本轮不做：

- 不处理 topology studio 对 `SectionCard` 的另一套深度覆盖
- 不处理 modal / drawer / action menu 的 slot-style contract
- 不做 `SectionCard` 目录迁移或 shared/common 体系重排

## Files to modify
- .harness/reuse-decisions/frontend-section-card-style-contract-convergence.md
- docs/plan/impl-plan/2026-06-04-section-card-style-contract-convergence-plan.md
- code/frontend/src/shared/ui/common/SectionCard.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspaceContent.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue
- code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue
- code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue
- code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue
- code/frontend/src/widgets/review-archive-workspace/ReviewArchiveWorkspace.vue
- code/frontend/src/widgets/review-archive-workspace/ReviewArchiveSummarySection.vue
- code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts
- code/frontend/scripts/vue-deep-allowlist.json
- docs/reviews/frontend/2026-06-04-section-card-style-contract-convergence-review.md

## After implementation
- student analysis / review archive 不再通过父级 `:deep(.section-card*)` 反向碰 `SectionCard` 内部结构。
- `SectionCard` 拥有可复用的显式样式 contract，后续同类页面优先声明 contract，而不是再写新的深度覆盖。
