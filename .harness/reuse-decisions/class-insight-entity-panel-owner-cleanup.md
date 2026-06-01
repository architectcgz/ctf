# Reuse Decision

## Change type
frontend refactor / entity presentation owner cleanup

## Existing code searched
- `code/frontend/src/entities/challenge/*`
- `code/frontend/src/entities/training-timeline/*`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassInsightsPanel.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassReviewPanel.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassTrendPanel.vue`
- `code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportPreviewSection.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherEyebrowSharedStyles.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportDialog.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `entities/challenge` 已经承接 challenge 对象的共享展示组件与公共入口。
- `entities/training-timeline` 已经承接教学时间线的稳定展示面板，并由多个 feature 复用。
- 当前 `ClassInsightsPanel`、`ClassReviewPanel`、`ClassTrendPanel` 都是围绕班级洞察对象的稳定展示块，不包含 route、dialog 或导出 workflow owner。

## Decision
refactor_existing

## Reason
当前问题不是 panel 还留在旧 `components/teacher`，而是它们虽然已经搬进 `class-students-workspace`，却仍被 `class-report-export` 直接深导入 feature 内部 UI 文件。这样会把一个稳定展示块错误地挂在单个 workflow feature 下，继续保留 cross-feature internal import。

这次最小正确收口方式是：

- 新建 `entities/class-insight` 作为班级洞察稳定展示 owner
- 让 `ClassStudentsPage.vue` 和 `ClassReportExportPreviewSection.vue` 都只通过实体 public API 消费三块 panel
- 同步更新 raw-source 护栏和 backlog 进展，明确这组三个 panel 已经脱离 feature internal owner

本轮不改：

- `class-insight-window` query / window workflow owner
- `student-analysis-workspace`、`StudentInsightPanel.vue`
- 三个 panel 的业务逻辑和视觉结构

## Files to modify
- `.harness/reuse-decisions/class-insight-entity-panel-owner-cleanup.md`
- `docs/plan/impl-plan/2026-06-01-class-insight-entity-panel-owner-cleanup-plan.md`
- `code/frontend/src/entities/class-insight/index.ts`
- `code/frontend/src/entities/class-insight/ui/index.ts`
- `code/frontend/src/entities/class-insight/ui/ClassInsightsPanel.vue`
- `code/frontend/src/entities/class-insight/ui/ClassReviewPanel.vue`
- `code/frontend/src/entities/class-insight/ui/ClassTrendPanel.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/index.ts`
- `code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportPreviewSection.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherEyebrowSharedStyles.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportDialog.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `entities/class-insight` 会成为 `ClassInsightsPanel`、`ClassReviewPanel`、`ClassTrendPanel` 的唯一稳定展示 owner。
- `class-students-workspace` 保留 page/workspace owner，不再承接这组三个共享 panel 的内部文件落点。
- `class-report-export` 不再深导入另一个 feature 的内部 UI 文件。
