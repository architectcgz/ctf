# Reuse Decision

## Change type
refactor_existing / owner-tightening / docs / test

## Existing code searched
- `code/frontend/src/shared/model/common/useReportStatusPolling.ts`
- `code/frontend/src/shared/model/common/__tests__/useReportStatusPolling.test.ts`
- `code/frontend/src/features/**/model/*Report*.ts`
- `code/frontend/src/features/**/model/*Export*.ts`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `docs/plan/impl-plan/2026-05-31-composables-owner-cleanup-batch3-report-polling-plan.md`

## Similar implementations found
- `shared/model/layout/*` 已按共享状态 owner 的更细语义拆出独立子域，而不是把所有共享 owner 都继续塞在 `common`
- `useReportStatusPolling` 当前所有调用点都围绕报告导出 / 报告生成状态，不存在其他真正通用的 polling 语义复用

## Decision
refactor_existing

## Reason
- `useReportStatusPolling` 虽然已经从历史 `composables/` 收口出来，但当前停在 `shared/model/common` 仍然过宽
- 它不是 `common` 级的共享反馈或通用状态，而是共享 reporting workflow owner
- 把它收紧到 `shared/model/reporting/` 可以去掉“common 只是临时落点”的歧义，同时不引入新的抽象层

## Files to modify
- `.harness/reuse-decisions/report-status-polling-reporting-owner-tightening.md`
- `docs/plan/impl-plan/2026-05-31-report-status-polling-reporting-owner-tightening-plan.md`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `code/frontend/src/shared/model/common/useReportStatusPolling.ts`
- `code/frontend/src/shared/model/common/__tests__/useReportStatusPolling.test.ts`
- `code/frontend/src/shared/model/reporting/useReportStatusPolling.ts`
- `code/frontend/src/shared/model/reporting/__tests__/useReportStatusPolling.test.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/features/contest-workbench/model/useContestExportFlow.ts`
- `code/frontend/src/features/profile/model/useUserProfilePage.ts`
- `code/frontend/src/features/teaching/class-report-export/model/useClassReportExport.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/teaching/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`

## After implementation
- 报告导出状态轮询的共享 owner 从 `shared/model/common` 收紧到 `shared/model/reporting`
- `shared/model/common` 继续只保留更宽泛的共享反馈 / 通用状态 owner
