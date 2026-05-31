# Reuse Decision

## Change type
frontend refactor / awd review detail owner convergence

## Existing code searched
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/pages/awd-review/__tests__/PlatformAwdReviewDetail.test.ts`
- `code/frontend/src/pages/awd-review/__tests__/TeacherAWDReviewDetail.test.ts`
- `code/frontend/src/features/contest-projector/model/useContestProjectorPage.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisDataState.ts`

## Similar implementations found
- `useAwdReviewDetailPage.ts` 目前同时持有 route query、详情读取、team 选择、breadcrumb 和导出编排。
- `useContestProjectorPage.ts` 已经采用 `page + data` 组合方式，让 page 只保留页面级 orchestration。
- `useStudentAnalysisDataState.ts` 展示了把详情读取和派生状态收口到 data owner 的现成模式。

## Decision
refactor_existing

## Reason
当前最小正确切片是把 AWD 复盘详情里的数据 owner 先拆到 `useAwdReviewDetailData`：

- `useAwdReviewDetailData`：承接 `getAwdReviewByRole`、`review`、`selectedTeamId` 和详情派生数据。
- `useAwdReviewDetailPage`：保留 route query / navigation、breadcrumb、export flow 和页面级 watch。

这样可以：

- 去掉 `useAwdReviewDetailPage` 里混合的详情读取与 team 选择 owner
- 保持 AWD detail feature 的业务语义，不把它过早抽成 shared loader

本轮不做：

- 不改 `AwdReviewWorkspace` widget 和 detail route page 模板
- 不改 `useAwdReviewExportFlow`
- 不追加 round/team query contract 变更

## Files to modify
- `.harness/reuse-decisions/awd-review-detail-owner.md`
- `docs/plan/impl-plan/2026-05-31-awd-review-detail-owner-plan.md`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailData.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailData.test.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/index.ts`
- `code/frontend/src/pages/awd-review/__tests__/PlatformAwdReviewDetail.test.ts`
- `code/frontend/src/pages/awd-review/__tests__/TeacherAWDReviewDetail.test.ts`

## After implementation
- AWD 复盘详情的数据读取、team 抽屉选择和摘要派生会集中到 `useAwdReviewDetailData`。
- `useAwdReviewDetailPage` 只保留页面级路由与 workflow 编排。
