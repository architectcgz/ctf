# Reuse Decision

## Change type
frontend refactor / feature-owned teacher class report export dialog decomposition

## Existing code searched
- code/frontend/src/features/teacher-class-report-export/ui/ClassReportExportDialog.vue
- code/frontend/src/features/teacher-class-report-export/model/useClassReportExport.ts
- code/frontend/src/components/teacher/reports/__tests__/ClassReportExportDialog.test.ts
- code/frontend/src/views/teacher/__tests__/ClassManagement.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts
- code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- 最近的 `PlatformContestFormPanel.vue`、`ContestChallengeOrchestrationPanel.vue`、`ContestProjectorAttackMap.vue` 都已经按“父层保留唯一 workflow / shell owner，稳定展示分区下沉，护栏切到聚合源码视角”的模式完成收口；`ClassReportExportDialog.vue` 也适合沿同一模式继续收口。
- `useClassReportExport()` 已经把班级报告导出的远端 workflow、预览加载、导出任务、下载和轮询状态都收成单点 owner；这说明当前主矛盾不是 workflow 还没抽，而是 dialog template / styles 仍然过宽。
- `ClassTrendPanel`、`ClassReviewPanel`、`ClassInsightsPanel` 已经是预览区内部的稳定展示区块，因此继续把外层 preview shell 下沉不会破坏既有 feature contract。

## Decision
refactor_existing

## Reason
`ClassReportExportDialog.vue` 当前约 `880` 行，父组件同时混放：

- dialog shell owner：`modelValue`、`closeDialog()`、`dialogVisible`
- prop -> export context sync：两组 `watch()` + `loadPreview()`
- 多个稳定展示区块：context chips、导出设置、preview summary、preview stack、latest task、guide
- 大体量 dialog 样式

最小正确改动不是再往 `useClassReportExport()` 里塞展示逻辑，也不是继续让单文件承载所有 dialog section，而是：

- 保持 `ClassReportExportDialog.vue` 继续做 dialog shell owner 与 prop/context sync owner
- 新增 `ClassReportExportContextSection.vue` 承接当前上下文、导出设置、preview snapshot 和创建导出动作
- 新增 `ClassReportExportPreviewSection.vue` 承接 error / loading / empty / preview stack
- 新增 `ClassReportExportTaskRail.vue` 承接 latest task 与 guide rail
- 新增 `classReportExportDialog.css` 承接 dialog shell 与 section 样式
- 同步把 raw-source 护栏改成聚合源码视角

本轮不调整 `useClassReportExport()` 的 workflow owner，不改 `AdminSurfaceModal` contract，不改 teacher / platform 路由对 `ClassReportExportDialog` 的 public API。

## Files to modify
- .harness/reuse-decisions/class-report-export-dialog-decomposition.md
- docs/plan/impl-plan/2026-05-28-class-report-export-dialog-decomposition-plan.md
- docs/reviews/frontend/2026-05-28-class-report-export-dialog-decomposition-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/teacher-class-report-export/ui/ClassReportExportDialog.vue
- code/frontend/src/features/teacher-class-report-export/ui/ClassReportExportContextSection.vue
- code/frontend/src/features/teacher-class-report-export/ui/ClassReportExportPreviewSection.vue
- code/frontend/src/features/teacher-class-report-export/ui/ClassReportExportTaskRail.vue
- code/frontend/src/features/teacher-class-report-export/ui/classReportExportDialog.css
- code/frontend/src/components/teacher/reports/__tests__/ClassReportExportDialog.test.ts

## After implementation
- `ClassReportExportDialog.vue` 会回到“dialog shell owner + prop/context sync owner”这一层，不再继续内联三个大 section 和整段 dialog 样式。
- teacher / platform 侧调用入口保持不变，但 class report export 这条 feature 的内部 owner 会更清晰，后续若还要继续瘦身，只需在 feature 内局部处理。
- backlog 里的 feature 内大组件债会新增一条这次收口记录，便于后续重新评估剩余超大 feature surface。
