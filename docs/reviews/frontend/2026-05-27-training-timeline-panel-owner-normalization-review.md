# Training Timeline Panel Owner Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-training-timeline-panel-owner-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/training-timeline-panel-owner-normalization.md`
    - `docs/plan/impl-plan/2026-05-27-training-timeline-panel-owner-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-27-training-timeline-panel-owner-normalization-review.md`
    - `code/frontend/src/components/training/TrainingTimelinePanel.vue`
    - `code/frontend/src/components/teacher/StudentInsightPanel.vue`
    - `code/frontend/src/features/student-dashboard/ui/studentDashboardPanelRegistry.ts`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts`
    - `code/frontend/src/views/__tests__/metricPanelSurfaceOwnership.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 shared panel owner 收口，保留 dashboard / teacher 的数据 owner，只切 timeline 展示入口。
- Gate verdict：Pass after targeted verification

## Findings

- None.

## Material findings

- None.

## Senior implementation assessment

- `TrainingTimelinePanel.vue` 已承接原 `StudentTimelinePage.vue` 的分页、分组、metric-panel 摘要和时间线展示逻辑，student dashboard 与 teacher 学员洞察都通过同一个共享 panel 入口消费它。
- `StudentInsightPanel.vue` 与 `studentDashboardPanelRegistry.ts` 只替换了 timeline 展示入口，没有把 timeline 数据 owner 从各自上游页面 / feature model 打散。
- `legacyComponentPageAllowlist` 中最后一条 student dashboard 历史 page 例外已移除，raw-source 与 metric-panel / theme token 护栏也已切到新路径。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/metricPanelSurfaceOwnership.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## Residual risk

- timeline 相关 helper 目前仍留在 `components/dashboard/student/utils.ts`。命名层面还有一层 student 语义残留，但它不再影响共享 panel owner；后续如果继续做 shared utility 命名收口，可以再把这组 timeline helper 单独移走。

## Touched known-debt status

- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 中 student dashboard 这组历史 page-sized UI 已在 touched surface 上全部迁离 `components/**Page.vue` 路径。
