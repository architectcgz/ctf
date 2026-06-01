# Class Report Export Dialog Decomposition 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-class-report-export-dialog-decomposition-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/class-report-export-dialog-decomposition.md`
    - `docs/plan/impl-plan/2026-05-28-class-report-export-dialog-decomposition-plan.md`
    - `docs/reviews/frontend/2026-05-28-class-report-export-dialog-decomposition-review.md`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
    - `code/frontend/src/features/teacher-class-report-export/model/useClassReportExport.ts`
    - `code/frontend/src/features/teacher-class-report-export/ui/ClassReportExportDialog.vue`
    - `code/frontend/src/features/teacher-class-report-export/ui/ClassReportExportContextSection.vue`
    - `code/frontend/src/features/teacher-class-report-export/ui/ClassReportExportPreviewSection.vue`
    - `code/frontend/src/features/teacher-class-report-export/ui/ClassReportExportTaskRail.vue`
    - `code/frontend/src/features/teacher-class-report-export/ui/classReportExportDialog.css`
    - `code/frontend/src/components/teacher/reports/__tests__/ClassReportExportDialog.test.ts`
- Classification check：同意按 `teacher-class-report-export` feature 内部超大 dialog surface 收口处理，属于非 trivial frontend refactor。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `ClassReportExportDialog.vue` 现在只保留 `AdminSurfaceModal` shell、`modelValue` / default context contract、`dialogVisible` 与两组 `watch()` 的 sync owner，不再继续混放多个稳定 section 和整段样式。
- `ClassReportExportContextSection.vue` 已明确承接当前教师上下文、导出设置表单壳、preview snapshot 和 reload / export 动作入口；父层不再同时承担表单展示与 remote workflow 同步。
- `ClassReportExportPreviewSection.vue` 已明确承接 preview error / loading / empty / KPI / trend-review-insights stack；这块展示 owner 不再和 dialog shell 混写。
- `ClassReportExportTaskRail.vue` 已明确承接 latest task banner、details、download button 与 guide rail；最近一次任务的展示逻辑不再停留在父 SFC。
- `classReportExportDialog.css` 已承接 dialog shell、section shell、task rail 与 responsive 样式；拆出 section 之后没有再依赖父 SFC `scoped` 样式去命中子组件内部节点。
- `ClassReportExportDialog.test.ts` 已改成聚合源码视角，继续覆盖 `AdminSurfaceModal`、表单原语和导出行为，不会因为 section / CSS 下沉而误报父文件“未使用原语”。
- `ClassReportExportDialog.vue` 文件体量从约 `880` 行降到 `169` 行；本轮 touched surface 上的“dialog shell + sync watch + 多 section 模板 + 大段样式”混写债已经完成收口。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/components/teacher/reports/__tests__/ClassReportExportDialog.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；按 `development-pipeline` 的独立 gate 要求，依然缺少单独 reviewer 上下文的复核证据。当前流程里没有额外派生 reviewer，所以这条缺口需要在交付说明里明确。
- `ClassReportExportContextSection.vue` 仍然约 `250` 行，因为它继续同时承接表单区和 preview snapshot；如果 class report export 后续继续增长，下一刀更适合在 feature 内继续按 `form shell / snapshot shell` 细分，而不是回退到父 dialog 再混写。
- raw-source 护栏里的 CSS 读取现在依赖 `code/frontend` 作为测试工作目录；当前项目测试命令一直从这个目录执行，因此本轮可接受，但如果后续统一从仓库根目录驱动这组前端测试，需要同步调整这条读取方式。

## Touched known-debt status

- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 中新增的“其它 feature 内残余大组件”P2 已在 `ClassReportExportDialog.vue` 这块 touched surface 上完成一刀收口；当前 residual 重点已经转向 `challenge-writeup-editor`、`awd-inspector` 这类仍未进一步拆分的 feature 内大 surface，而不再是 class report dialog 父壳继续混写 section 与样式。
