> 状态：Current
> 事实源：contest edit surface feature-owned UI normalization
> 替代：无

# Contest Edit Surface Feature UI Normalization Plan

## 目标

- 把 `ContestEditTopbarPanel.vue` 迁入 `features/platform-contests/ui`。
- 把 `ContestChallengeEditorDialog.vue`、`ContestAwdChallengeSelectorSection.vue`、`ContestChallengeSettingsSection.vue` 迁入 `features/contest-workbench/ui`。
- 让 `ContestEdit.vue` 与 `ContestChallengeOrchestrationPanel.vue` 改为通过 feature public API 组合。

## 非目标

- 本轮不迁 `AWDChallengeConfigPanel.vue`。
- 本轮不改 `useContestEditPage.ts` 或 `useContestChallengeOrchestration.ts` 的 model owner。
- 本轮不重做 dialog 的表单逻辑、分页逻辑或 topbar 视觉样式。

## 输入依据

- `code/frontend/src/components/platform/contest/ContestEditTopbarPanel.vue`
- `code/frontend/src/components/platform/contest/ContestChallengeEditorDialog.vue`
- `code/frontend/src/components/platform/contest/ContestAwdChallengeSelectorSection.vue`
- `code/frontend/src/components/platform/contest/ContestChallengeSettingsSection.vue`
- `code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue`
- `code/frontend/src/features/contest-workbench/ui/ContestChallengeOrchestrationPanel.vue`
- `code/frontend/src/views/platform/ContestEdit.vue`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `ContestEditTopbarPanel.vue` 是典型的 route shell UI，只服务 contest edit 页面。
- `ContestChallengeEditorDialog.vue` 与其两个 section 只服务 `contest-workbench` 的题目编排面板，迁父不迁子会留下半迁移状态。
- 当前最小正确切片是按 owner 分到 `platform-contests/ui` 与 `contest-workbench/ui`，而不是再新增第三个桶。

## 设计边界

### `features/platform-contests/ui/*` 本轮负责

- contest edit topbar panel

### `features/contest-workbench/ui/*` 本轮负责

- challenge editor dialog
- awd challenge selector section
- challenge settings section

### `views/platform/ContestEdit.vue` 本轮继续负责

- route shell 组合
- 顶层 loading overlay

## 任务切片

### Slice 1：contest edit topbar 落位

- 目标：
  - 新增 `features/platform-contests/ui/ContestEditTopbarPanel.vue`
  - `platform-contests/ui/index.ts` 暴露 topbar
  - `ContestEdit.vue` 改走 feature public API
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/contestEditTopbarExtraction.test.ts src/views/platform/__tests__/ContestEdit.test.ts`
- Review focus：
  - topbar 是否仍只保留 props / emits contract
  - route owner 是否没有回流到 topbar

### Slice 2：challenge editor dialog 落位

- 目标：
  - 新增 `features/contest-workbench/ui/ContestChallengeEditorDialog.vue`
  - 同步迁移 `ContestAwdChallengeSelectorSection.vue` 与 `ContestChallengeSettingsSection.vue`
  - `ContestChallengeOrchestrationPanel.vue` 改走 feature 内部相对 UI import
  - raw-source / dialog / duplicate-action / typography 护栏同步更新
- 验证：
  - `npm run test:run -- src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts src/views/__tests__/studentDirectoryTypographyBoundary.test.ts src/components/common/__tests__/BackofficeDialogAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts`
- Review focus：
  - dialog 的 submit/in-flight owner 是否不变
  - selector/settings sections 是否不再滞留旧组件目录

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/contestEditTopbarExtraction.test.ts src/views/platform/__tests__/ContestEdit.test.ts src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts src/views/__tests__/studentDirectoryTypographyBoundary.test.ts src/components/common/__tests__/BackofficeDialogAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## 残余风险

- `AWDChallengeConfigPanel.vue` 仍留在旧路径，本轮只先收口 contest edit 线上剩余的 topbar 与 challenge editor dialog surface。
