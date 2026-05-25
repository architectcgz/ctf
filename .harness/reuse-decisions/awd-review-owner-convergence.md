# Reuse Decision

## Change type
feature / widget / view / test / docs

## Existing code searched
- `code/frontend/src/views/platform/AWDReviewIndex.vue`
- `code/frontend/src/views/platform/PlatformAwdReviewDetail.vue`
- `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
- `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`
- `code/frontend/src/features/teacher-awd-review/**`
- `code/frontend/src/features/awd-review-detail-workspace/**`
- `code/frontend/src/widgets/teacher-awd-review/**`

## Similar implementations found
- `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewIndex.ts`
- `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewExportFlow.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/widgets/teacher-awd-review/index.ts`
- `code/frontend/src/features/class-students-workspace/**`
- `code/frontend/src/features/student-review-archive-workspace/**`

## Decision
refactor_existing

## Reason
AWD 复盘已经同时被 teacher / platform 两侧 route view 复用，但共享 owner 仍挂在 `teacher-awd-review` 路径下。继续保留这层 role-specific owner，会让平台页继续跨到 teacher 语义目录，弱化共享边界，也让后续 source-boundary 检查长期携带例外。现有实现已经足够复用，不需要再造一套新 widget 或 feature；更低风险的做法是直接把共享 feature / widget 迁到中性 owner，更新调用方和测试，然后删除旧目录，不留兼容出口。

## Files to modify
- `.harness/reuse-decisions/awd-review-owner-convergence.md`
- `docs/plan/impl-plan/2026-05-24-awd-review-owner-convergence-implementation-plan.md`
- `code/frontend/src/features/teacher-awd-review/**`
- `code/frontend/src/features/awd-review-workspace/**`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewExportFlow.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/widgets/teacher-awd-review/**`
- `code/frontend/src/widgets/awd-review-workspace/**`
- `code/frontend/src/widgets/awd-review-workspace/TeacherAWDReviewContestDirectory.vue`
- `code/frontend/src/widgets/awd-review-workspace/TeacherAWDReviewContestHead.vue`
- `code/frontend/src/widgets/awd-review-workspace/TeacherAWDReviewContestRow.vue`
- `code/frontend/src/widgets/awd-review-workspace/TeacherAWDReviewContestRowCta.vue`
- `code/frontend/src/widgets/awd-review-workspace/TeacherAWDReviewContestRowMetrics.vue`
- `code/frontend/src/widgets/awd-review-workspace/TeacherAWDReviewContestRowStatusTags.vue`
- `code/frontend/src/widgets/awd-review-workspace/TeacherAWDReviewDirectorySection.vue`
- `code/frontend/src/widgets/awd-review-workspace/TeacherAWDReviewDirectoryState.vue`
- `code/frontend/src/widgets/awd-review-workspace/TeacherAWDReviewIndexFilters.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewStatusChip.vue`
- `code/frontend/src/widgets/awd-review-workspace/TeacherAWDReviewSummaryPanel.vue`
- `code/frontend/src/widgets/awd-review-workspace/TeacherAWDReviewSurfaceShell.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.vue`
- `code/frontend/src/components/teacher/awd-review/AwdReviewEvidenceGrid.vue`
- `code/frontend/src/components/teacher/awd-review/AwdReviewAnalysisSection.vue`
- `code/frontend/src/components/teacher/awd-review/AwdReviewTeamDrawer.vue`
- `code/frontend/src/widgets/awd-review-workspace/TeacherAWDReviewWorkspaceActions.vue`
- `code/frontend/src/widgets/awd-review-workspace/TeacherAWDReviewWorkspaceHeader.vue`
- `code/frontend/src/widgets/awd-review-workspace/TeacherAWDReviewWorkspaceState.vue`
- `code/frontend/src/views/platform/AWDReviewIndex.vue`
- `code/frontend/src/views/platform/PlatformAwdReviewDetail.vue`
- `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
- `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`
- `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## After implementation
- AWD 复盘共享 owner 统一从中性目录暴露，teacher / platform route view 不再直连 role-specific `teacher-awd-review` 路径。
- AWD 目录页共享 feature 对外不再暴露 `useTeacherAwdReviewIndex` 这类 teacher-specific 命名。
- AWD 目录页共享 widget 对外不再暴露 `TeacherAWDReviewIndexWorkspace` 这类 teacher-specific 命名。
- AWD 共享 widget 公共出口只保留 `AwdReviewWorkspace` 与 `AwdReviewIndexWorkspace` 两个中性入口，不再暴露次级 `TeacherAWDReview*` 组合部件名。
- AWD 详情页共享 feature 对外不再暴露 `useTeacherAwdReviewExportFlow` 这类 teacher-specific 命名。
- AWD 详情页共享 widget 对外不再暴露 `TeacherAWDReviewWorkspace` 这类 teacher-specific 命名。
- 仍然带有 `Teacher*` 前缀的内部 widget 名称先视为实现细节，不在本轮扩大为命名清洗。
