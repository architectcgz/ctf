# Section Card Style Contract Convergence Review

## Review Target

- Repository: `ctf`
- Branch: `main`
- Diff source: working tree changes for `frontend-section-card-style-contract-convergence`
- Files reviewed:
  - `.harness/reuse-decisions/frontend-section-card-style-contract-convergence.md`
  - `docs/plan/impl-plan/2026-06-04-section-card-style-contract-convergence-plan.md`
  - `code/frontend/src/shared/ui/common/SectionCard.vue`
  - `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
  - `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspaceContent.vue`
  - `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.vue`
  - `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue`
  - `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue`
  - `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue`
  - `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue`
  - `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveWorkspace.vue`
  - `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveSummarySection.vue`
  - `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
  - `code/frontend/scripts/check-vue-deep-guard.mjs`
  - `code/frontend/scripts/vue-deep-allowlist.json`
  - `code/frontend/package.json`
  - `scripts/check-frontend-architecture.sh`
  - `scripts/check-consistency.sh`

## Classification Check

- 结论：同意 `non-trivial frontend refactor` 分类。
- 原因：这次同时触达共享组件样式 contract、teacher surface consumer、前端源码 guardrail 和原始源码护栏测试，不是局部样式替换。

## Gate Verdict

- `pass with minor issues`
- 说明：当前结论来自同上下文显式自审归档，不替代独立 reviewer gate。

## Findings

- 无 blocker / major / minor finding。

## Material Findings

- 无。

## Senior Implementation Assessment

- 当前切口合理：把 `SectionCard` 的 teacher surface 差异收口成显式 `variant` 和 CSS variable contract，比继续在 page / widget 父级追加 `:deep(.section-card*)` 更稳定，也更接近共享 owner 应承担的职责。
- 这次没有把 `:deep` 只是“挪位置”，而是把 student analysis / review archive 这批 consumer 迁到声明式 contract，上层只保留局部变量覆写，避免以后再从父级穿透共享组件内部结构。
- 在 touched surface 上顺手把 SFC block order 统一回 `template -> style -> script` 是必要收口；否则这批改动会继续把已知仓库约定违背带进后续批次。

## Required Re-validation

- `cd code/frontend && npm run check:vue-deep`
- `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `cd code/frontend && npm run typecheck`
- `bash scripts/check-frontend-architecture.sh --quick`
- `bash scripts/check-consistency.sh`
- `git diff --check -- code/frontend/src/shared/ui/common/SectionCard.vue code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspaceContent.vue code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.vue code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue code/frontend/src/widgets/review-archive-workspace/ReviewArchiveWorkspace.vue code/frontend/src/widgets/review-archive-workspace/ReviewArchiveSummarySection.vue code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts code/frontend/scripts/check-vue-deep-guard.mjs code/frontend/scripts/vue-deep-allowlist.json code/frontend/package.json scripts/check-frontend-architecture.sh scripts/check-consistency.sh .harness/reuse-decisions/frontend-section-card-style-contract-convergence.md docs/plan/impl-plan/2026-06-04-section-card-style-contract-convergence-plan.md docs/reviews/frontend/2026-06-04-section-card-style-contract-convergence-review.md`

## Residual Risk

- 本轮只收 `SectionCard` 这条共享 owner 链，不处理 topology studio 中仍依赖 `:deep(.section-card*)` 的 surface，也不处理 modal / drawer / action menu 这类 slot-style contract；这些存量仍在 allowlist 中，需要后续按独立批次继续收口。
- `code/frontend/src/features/teacher/dashboard/*` 这组并行工作树改动不在本次 review 结论内；本次没有基于那组变更判断回归。
- 独立 reviewer gate 还没有满足；如果要把这批作为真正完成态推进提交，仍需要一次脱离当前实现上下文的复审。

## Touched Known-Debt Status

- 本次直接触达的已知 debt 是 `SectionCard` 样式 owner 漂移和 teacher detail / review archive 通过父级 `:deep` 覆盖共享组件内部结构。
- 在本批次 touched surface 内，这部分 debt 已收口完成：共享组件新增显式 contract，首批 consumer 不再依赖 `:deep(.section-card*)`，并新增 guardrail 阻止未登记的 `:deep` 回流。
- 仓库级剩余 `:deep` 存量仍然存在，但属于本批次明确未触达的其他 owner 面，不构成本批 diff 的 blocker。
