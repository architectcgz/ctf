# Reuse Decision

## Change type
frontend refactor / feature-owned form panel decomposition

## Existing code searched
- code/frontend/src/features/platform-contests/ui/PlatformContestFormPanel.vue
- code/frontend/src/features/platform-contests/ui/PlatformContestFormDialog.vue
- code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue
- code/frontend/src/features/platform-contests/ui/ContestOrchestrationPage.vue
- code/frontend/src/features/platform-contests/ui/index.ts
- code/frontend/src/features/platform-contests/model/index.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `ContestEditWorkspacePanel.vue` 已经作为 feature 内组合壳承接竞赛编辑 workspace，本轮更适合让它继续消费更瘦的 `PlatformContestFormPanel`，而不是把表单 owner 再上提到 route / page。
- `ContestAnnouncementsWorkspacePanel.vue`、`ContestOperationsHubWorkspacePanel.vue` 这类 feature-owned workspace panel 都采用“父壳持有 workflow owner，稳定展示区拆成子面板”的模式，适合作为本轮 `PlatformContestFormPanel` 的切法基线。
- 当前 `PlatformContestFormPanel.vue` 内三大分区的结构高度稳定，只共享 `draft / fieldErrors / fieldLocks / statusOptions` 等输入，适合抽成 section 级子组件，而不需要新建 store 或新的 page model。

## Decision
refactor_existing

## Reason
`PlatformContestFormPanel.vue` 当前同时承接本地 draft 同步、字段校验、提交入口、三块表单区和整套布局样式，约 652 行，已经属于 feature 内部过宽 surface。最小正确改动不是改它的对外 API，也不是把校验挪到上层页面，而是：

- 保持 `PlatformContestFormPanel.vue` 继续做唯一的 draft owner、校验 owner 和 submit owner
- 新增一个共享 section shell，统一 icon / title / description / content 装配
- 把“基础信息”“赛制与状态”“赛程时间轴”“底部动作条”拆成 feature 内局部子组件
- 局部样式跟随这些子组件或 feature 内共享 section shell 收口，避免父面板继续持有整块展示样式全集

本轮不调整 `ContestFormDraft` / `ContestFieldLocks` contract，不改变 `PlatformContestFormDialog.vue`、`ContestEditWorkspacePanel.vue`、`ContestOrchestrationPage.vue` 的使用方式，也不把表单状态提升到 page model。

## Files to modify
- .harness/reuse-decisions/platform-contest-form-panel-decomposition.md
- docs/plan/impl-plan/2026-05-28-platform-contest-form-panel-decomposition-plan.md
- docs/reviews/frontend/2026-05-28-platform-contest-form-panel-decomposition-review.md
- code/frontend/src/features/platform-contests/ui/PlatformContestFormPanel.vue
- code/frontend/src/features/platform-contests/ui/PlatformContestFormSectionShell.vue
- code/frontend/src/features/platform-contests/ui/PlatformContestIdentitySection.vue
- code/frontend/src/features/platform-contests/ui/PlatformContestRulesSection.vue
- code/frontend/src/features/platform-contests/ui/PlatformContestTimelineSection.vue
- code/frontend/src/features/platform-contests/ui/PlatformContestFormActions.vue
- code/frontend/src/features/platform-contests/ui/platformContestFormPanel.css
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `PlatformContestFormPanel.vue` 会回到“draft sync + validate + save”这一层 owner，不再直接内联三块 section 模板和大段 section 样式。
- `features/platform-contests/ui` 会形成清晰的 contest form section cluster，后续 `ContestChallengeOrchestrationPanel.vue` 或 dialog/workspace 再调整时可以直接复用。
- 当前 P2 backlog 里 `PlatformContestFormPanel.vue` 这一项可以从“超大组件待拆”转成已完成或至少显著收口。
