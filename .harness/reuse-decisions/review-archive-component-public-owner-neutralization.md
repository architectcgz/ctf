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
- `widgets/review-archive-workspace/index.ts` 已经提供了中立 workspace 入口，说明 review archive 最终 owner 本来就应落在 shared widget，而不是继续停在 `teacher-review-archive` 或 `components/review-archive`。
- AWD review 已经验证过“共享 detail 子块直接并入最终 widget owner、删除中间 bridge”这条路径可行。

## Decision
refactor_existing

## Reason
- `ReviewArchiveWorkspace` 当前被 teacher / platform 共用，但代码还分散在 `widgets/teacher-review-archive/*`、`components/review-archive/index.ts` 与 `components/teacher/review-archive/*` 三层历史目录里。
- 这次直接把 workspace、本地 model 和四个 detail 子组件全部并入 `widgets/review-archive-workspace/*`，同步更新测试、类型索引和 backlog，并删除 `widgets/teacher-review-archive` 与 `components/review-archive` / `components/teacher/review-archive` 的残余入口，不保留中间 bridge。

## Files to modify
- `.harness/reuse-decisions/review-archive-component-public-owner-neutralization.md`
- `docs/plan/impl-plan/2026-05-27-review-archive-component-public-owner-neutralization-implementation-plan.md`
- `code/frontend/src/widgets/review-archive-workspace/index.ts`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveWorkspace.vue`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveWorkspace.test.ts`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveState.vue`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveState.test.ts`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveSummarySection.vue`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveSummarySection.test.ts`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveHero.vue`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveObservationStrip.vue`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveEvidencePanel.vue`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveEvidencePanel.test.ts`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveReflectionPanel.vue`
- `code/frontend/src/widgets/review-archive-workspace/reviewArchiveCases.ts`
- `code/frontend/src/widgets/review-archive-workspace/model/*`
- `code/frontend/src/widgets/teacher-review-archive/*`
- `code/frontend/src/widgets/teacher-review-archive/model/*`
- `code/frontend/src/components/review-archive/index.ts`
- `code/frontend/src/components/teacher/review-archive/*`
- `code/frontend/src/components.d.ts`
- `code/frontend/src/pages/review-archive/__tests__/TeacherStudentReviewArchive.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherSurface.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherEyebrowSharedStyles.test.ts`
- `code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-review-archive-component-public-owner-neutralization-review.md`

## After implementation
- `ReviewArchiveWorkspace` 与其 hero / observation / evidence / reflection / state / summary / presentation model 将直接落在 `widgets/review-archive-workspace/*`，teacher / platform route view 只保留对这一组中性 widget owner 的依赖。
