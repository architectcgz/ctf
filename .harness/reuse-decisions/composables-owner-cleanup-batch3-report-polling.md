# Reuse Decision

## Change type
composition / shared-model / docs / test

## Existing code searched
- `code/frontend/src/composables/useReportStatusPolling.ts`
- `code/frontend/src/features/**/model/*Export*.ts`
- `code/frontend/src/features/**/model/*Report*.ts`
- `code/frontend/src/features/profile/model/useUserProfilePage.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/teaching/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`
- `code/frontend/src/features/teaching/class-report-export/model/useClassReportExport.ts`
- `code/frontend/src/features/contest-workbench/model/useContestExportFlow.ts`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`

## Similar implementations found
- `shared/model/common/useToast.ts`、`shared/model/common/useDestructiveConfirm.ts`、`shared/model/common/usePagination.ts` 已承接跨 feature 共享 workflow / 状态 owner
- `useReportStatusPolling` 直接依赖 `getReportStatus()` 与 `ReportExportData`，语义上属于共享报告导出状态 owner，而不是纯浏览器或路由同步能力
- `routeNavigationTransport`、`routeQueryTransport`、`useRouteQueryTabs`、`useUrlSyncedTabs` 仍明显属于 router / query transport 方向，不应和报告轮询混在同一批

## Decision
refactor_existing

## Reason
- `useReportStatusPolling` 的复用点不是 DOM、事件或 URL，而是“报告导出状态轮询”这条跨 feature 的共享业务中性流程
- 它依赖报告导出 contract，和 `usePagination` 一样更接近 `shared/model/common/*` 的共享状态 owner，而不是历史 `src/composables`
- 先单独收口这一个 composable，可以避免 route/query 那组 owner 还没定型时把两类语义混成一批

## Files to modify
- `.harness/reuse-decisions/composables-owner-cleanup-batch3-report-polling.md`
- `docs/plan/impl-plan/2026-05-31-composables-owner-cleanup-batch3-report-polling-plan.md`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `code/frontend/src/composables/useReportStatusPolling.ts`
- `code/frontend/src/shared/model/common/useReportStatusPolling.ts`
- `code/frontend/src/shared/model/common/__tests__/useReportStatusPolling.test.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/features/contest-workbench/model/useContestExportFlow.ts`
- `code/frontend/src/features/profile/model/useUserProfilePage.ts`
- `code/frontend/src/features/teaching/class-report-export/model/useClassReportExport.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/teaching/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`

## After implementation
- 报告导出状态轮询从历史 `src/composables` 收口到 `shared/model/common`
- `src/composables` 继续保留更偏 router / query / transport 的通用能力，owner 边界更清楚
