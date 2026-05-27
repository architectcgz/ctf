# Reuse Decision

## Change type
+component / widget / docs / test

## Existing code searched
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.test.ts`
- `code/frontend/src/widgets/awd-review-workspace/index.ts`
- `code/frontend/src/components/teacher/awd-review/AwdReviewRoundSelector.vue`
- `code/frontend/src/components/teacher/awd-review/AwdReviewAnalysisSection.vue`
- `code/frontend/src/components/teacher/awd-review/AwdReviewEvidenceGrid.vue`
- `code/frontend/src/components/teacher/awd-review/AwdReviewTeamDrawer.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/teacher/__tests__/teacherAwdReviewWorkspaceExtraction.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `ClassReportExportDialog` 和 `ReviewArchiveWorkspace` 已经分别通过 `components/reports`、`components/review-archive` 补过中立 public owner，说明共享 UI 壳的 neutral barrel 收口模式已经稳定。
- `AwdReviewWorkspace` 当前同样被 teacher / platform 两边复用，teacher 组件入口只是历史目录归属残留。

## Decision
refactor_existing

## Reason
- `AwdReviewWorkspace` 仍直连四个 `components/teacher/awd-review/*` 路径，会继续放大 shared AWD review workspace 对 teacher 组件命名空间的结构依赖。
- 这次只增加 `components/awd-review` barrel、迁移 widget import，并同步 allowlist / 测试，不移动实际组件文件，也不改 AWD review 页面行为。

## Files to modify
- `.harness/reuse-decisions/awd-review-component-public-owner-neutralization.md`
- `docs/plan/impl-plan/2026-05-27-awd-review-component-public-owner-neutralization-implementation-plan.md`
- `code/frontend/src/components/awd-review/index.ts`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/teacher/__tests__/teacherAwdReviewWorkspaceExtraction.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-awd-review-component-public-owner-neutralization-review.md`

## After implementation
- `AwdReviewWorkspace` 将通过中立 `components/awd-review` public owner 读取 round selector / analysis / evidence / team drawer，相关 allowlist 也会从四条 teacher 组件路径进一步缩小。
