# Reuse Decision

## Change type
+component / widget / docs / test

## Existing code searched
- `code/frontend/src/widgets/teacher-review-archive/ReviewArchiveWorkspace.vue`
- `code/frontend/src/widgets/teacher-review-archive/ReviewArchiveWorkspace.test.ts`
- `code/frontend/src/widgets/review-archive-workspace/index.ts`
- `code/frontend/src/components/teacher/review-archive/ReviewArchiveHero.vue`
- `code/frontend/src/components/teacher/review-archive/ReviewArchiveObservationStrip.vue`
- `code/frontend/src/components/teacher/review-archive/ReviewArchiveEvidencePanel.vue`
- `code/frontend/src/components/teacher/review-archive/ReviewArchiveReflectionPanel.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/teacher/__tests__/teacherStudentReviewArchiveWorkspaceExtraction.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformStudentReviewArchive.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `ClassReportExportDialog` 已经通过 `components/reports` 新增了中立 public owner，说明共享 UI 壳先补 neutral barrel、再迁消费面的方式在本仓库可复用。
- `widgets/review-archive-workspace/index.ts` 已经给 review archive workspace 提供了中立 widget owner，剩下的是 widget 内部仍直连 teacher 组件入口。

## Decision
refactor_existing

## Reason
- `ReviewArchiveWorkspace` 当前被 teacher / platform 共用，但内部还直连四个 `components/teacher/review-archive/*` 路径，继续保留会放大 shared widget 对 teacher 组件命名空间的结构依赖。
- 这次只增加 `components/review-archive` barrel、迁移 widget import，并同步 allowlist / 测试，不移动实际组件文件，也不改 review archive 页面行为。

## Files to modify
- `.harness/reuse-decisions/review-archive-component-public-owner-neutralization.md`
- `docs/plan/impl-plan/2026-05-27-review-archive-component-public-owner-neutralization-implementation-plan.md`
- `code/frontend/src/components/review-archive/index.ts`
- `code/frontend/src/widgets/teacher-review-archive/ReviewArchiveWorkspace.vue`
- `code/frontend/src/widgets/teacher-review-archive/ReviewArchiveWorkspace.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/teacher/__tests__/teacherStudentReviewArchiveWorkspaceExtraction.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-review-archive-component-public-owner-neutralization-review.md`

## After implementation
- `ReviewArchiveWorkspace` 将通过中立 `components/review-archive` public owner 读取共享面板，相关 allowlist 也会从四条 teacher 组件路径收缩到一条中立 barrel。
