# Reuse Decision

## Change type
+entity / component / test / docs

## Existing code searched
- `code/frontend/src/components/training/TrainingTimelinePanel.vue`
- `code/frontend/src/components/dashboard/student/utils.ts`
- `code/frontend/src/components/dashboard/student/__tests__/utils.test.ts`
- `code/frontend/src/features/student-dashboard/ui/studentDashboardPanelRegistry.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- `code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts`
- `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts`
- `code/frontend/src/pages/__tests__/journalUserShellStyles.test.ts`
- `code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts`
- `docs/plan/impl-plan/2026-05-27-training-timeline-panel-owner-normalization-plan.md`
- `docs/reviews/frontend/2026-05-27-training-timeline-panel-owner-normalization-review.md`

## Similar implementations found
- `entities/challenge/ui/*` 已经承接稳定业务对象的共享展示组件，说明跨 dashboard / teacher 共用、且主要回答“训练时间线是什么、如何稳定展示”的 timeline 面板更适合落在 `entities/*`。
- 之前的 `components/training/TrainingTimelinePanel.vue` 已经完成 page -> shared panel 收口，这次是在既有共享 owner 基础上继续完成最终层级归位，不重做行为。

## Decision
refactor_existing

## Reason
- `TrainingTimelinePanel.vue` 当前只剩 student dashboard 和 teacher student insight 两个 consumer，且二者都把它当成稳定的 timeline 展示块使用。
- 它不是 workflow owner，也不是 page/widget shell；继续留在 `components/training` 只会保留历史中间态。
- timeline 的文案/样式 helper 仍藏在 `components/dashboard/student/utils.ts`，这会让共享 timeline owner 继续依赖 student 命名残余；应随同迁到 timeline entity owner。

## Files to modify
- `.harness/reuse-decisions/training-timeline-entity-owner-migration.md`
- `code/frontend/src/entities/training-timeline/index.ts`
- `code/frontend/src/entities/training-timeline/ui/TrainingTimelinePanel.vue`
- `code/frontend/src/entities/training-timeline/model/index.ts`
- `code/frontend/src/entities/training-timeline/model/presentation.ts`
- `code/frontend/src/entities/training-timeline/model/presentation.test.ts`
- `code/frontend/src/features/student-dashboard/ui/studentDashboardPanelRegistry.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- `code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts`
- `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts`
- `code/frontend/src/pages/__tests__/journalUserShellStyles.test.ts`
- `code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts`
- `code/frontend/src/components/training/TrainingTimelinePanel.vue`
- `code/frontend/src/components/dashboard/student/utils.ts`
- `code/frontend/src/components/dashboard/student/__tests__/utils.test.ts`

## After implementation
- `components/training` 应被清空。
- timeline 展示语义与 helper 将落在统一的 `entities/training-timeline/*` owner，student / teacher consumer 只依赖该实体公开入口。
