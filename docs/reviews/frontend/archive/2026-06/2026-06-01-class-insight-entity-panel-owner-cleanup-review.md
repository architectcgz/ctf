# Class Insight Entity Panel Owner Cleanup Review

## Review Target

- Repository: `ctf`
- Branch: `main`
- Diff source: working tree changes for `class-insight-entity-panel-owner-cleanup`
- Files reviewed:
  - `.harness/reuse-decisions/class-insight-entity-panel-owner-cleanup.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-01-class-insight-entity-panel-owner-cleanup-plan.md`
  - `code/frontend/src/entities/class-insight/index.ts`
  - `code/frontend/src/entities/class-insight/ui/index.ts`
  - `code/frontend/src/entities/class-insight/ui/ClassInsightsPanel.vue`
  - `code/frontend/src/entities/class-insight/ui/ClassReviewPanel.vue`
  - `code/frontend/src/entities/class-insight/ui/ClassTrendPanel.vue`
  - `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue`
  - `code/frontend/src/features/teaching/class-students-workspace/ui/index.ts`
  - `code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportPreviewSection.vue`
  - `code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportDialog.test.ts`
  - `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
  - `code/frontend/src/pages/teacher/__tests__/teacherEyebrowSharedStyles.test.ts`
  - `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Classification Check

- 结论：同意 `non-trivial frontend refactor` 分类。
- 原因：这次同时触达 entity / feature 边界、跨 feature 依赖方向、raw-source 护栏和 backlog 事实更新，不是局部样式或单文件修补。

## Gate Verdict

- `pass with minor issues`
- 说明：这次收口方向正确，未发现 material finding；但这份 review 仍由当前实现上下文完成，只能算显式自审，不能冒充独立 reviewer gate。

## Findings

- 无代码级 blocker / major / minor finding。

## Material Findings

- 无。

## Senior Implementation Assessment

- 把三块 panel 提到 `entities/class-insight` 是这次最小且边界正确的收口方式，因为它们描述的是班级洞察对象的稳定展示，而不是 `class-students-workspace` 或 `class-report-export` 的页面 workflow。
- `ClassStudentsPage.vue` 继续保留 workspace owner，`ClassReportExportPreviewSection.vue` 继续保留预览区组合 owner；两侧只改 import 落点，没有把 route、query、dialog 状态错吸进实体层。
- `class-students-workspace/ui/index.ts` 不再继续导出旧 panel 文件路径，这样可以避免“文件已经挪走，但 feature 还假装拥有它们”的半收口状态。

## Required Re-validation

- `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/TeacherClassStudents.test.ts src/features/teaching/class-report-export/ui/ClassReportExportDialog.test.ts src/pages/teacher/__tests__/teacherEyebrowSharedStyles.test.ts src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/class-insight-entity-panel-owner-cleanup.md docs/plan/archive/impl-plan/2026-06/2026-06-01-class-insight-entity-panel-owner-cleanup-plan.md code/frontend/src/entities/class-insight/index.ts code/frontend/src/entities/class-insight/ui/index.ts code/frontend/src/entities/class-insight/ui/ClassInsightsPanel.vue code/frontend/src/entities/class-insight/ui/ClassReviewPanel.vue code/frontend/src/entities/class-insight/ui/ClassTrendPanel.vue code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue code/frontend/src/features/teaching/class-students-workspace/ui/index.ts code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportPreviewSection.vue code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts code/frontend/src/pages/teacher/__tests__/teacherEyebrowSharedStyles.test.ts code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportDialog.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual Risk

- 还没有独立 subagent review 证据；如果后续这条迁移线需要严格满足 pipeline 的独立 reviewer gate，需要补一轮脱离当前实现上下文的审查。
- 历史 plan / reuse 文档里仍会提到旧 feature 路径，这些属于历史证据，不代表当前 owner 已回退。

## Touched Known-Debt Status

- 顶层 `ClassInsightsPanel` / `ClassReviewPanel` / `ClassTrendPanel` 的 cross-feature deep import 债务本轮已收口。
- `StudentInsightPanel.vue` 和更深层 workspace owner 仍是这组 backlog 的后续主面，但不再阻塞这次 class insight panel owner 的完成。
