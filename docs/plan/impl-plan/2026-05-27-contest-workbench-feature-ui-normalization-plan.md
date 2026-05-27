> 状态：Current
> 事实源：contest workbench feature-owned UI normalization
> 替代：无

# Contest Workbench Feature UI Normalization Plan

## 目标

- 把 `ContestWorkbenchStageTabs.vue`、`ContestWorkbenchSummaryStrip.vue`、`ContestChallengeFilterStrip.vue`、`ContestChallengeSummaryStrip.vue`、`ContestChallengeOrchestrationPanel.vue` 迁入 `features/contest-workbench/ui`。
- 把 `ContestEditWorkspacePanel.vue` 迁入 `features/platform-contests/ui`。
- 让 `ContestEdit.vue` 改为通过 `features/contest-workbench` 与 `features/platform-contests` public API 组合编辑页。

## 非目标

- 本轮不改 `ContestEditTopbarPanel.vue` 的落点。
- 本轮不改 `useContestEditPage.ts`、`useContestWorkbench.ts`、`useContestChallengeOrchestration.ts` 的 model owner。
- 本轮不重做 workbench 视觉或 stage 切换交互。

## 输入依据

- `code/frontend/src/components/platform/contest/ContestWorkbenchStageTabs.vue`
- `code/frontend/src/components/platform/contest/ContestWorkbenchSummaryStrip.vue`
- `code/frontend/src/components/platform/contest/ContestChallengeFilterStrip.vue`
- `code/frontend/src/components/platform/contest/ContestChallengeSummaryStrip.vue`
- `code/frontend/src/components/platform/contest/ContestChallengeOrchestrationPanel.vue`
- `code/frontend/src/components/platform/contest/ContestEditWorkspacePanel.vue`
- `code/frontend/src/features/contest-workbench/index.ts`
- `code/frontend/src/features/platform-contests/index.ts`
- `code/frontend/src/views/platform/ContestEdit.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `contest-workbench` 已有稳定 model owner，但缺少 `ui/` 落点，导致 `ContestEdit.vue` 和 `ContestEditWorkspacePanel.vue` 继续回头引用 legacy component 路径。
- `ContestEditWorkspacePanel.vue` 同时组合 `PlatformContestFormPanel`、`ContestChallengeOrchestrationPanel`、`AWDChallengeConfigPanel`、`ContestAwdPreflightPanel`，更接近 `platform-contests` 的 route shell，而不是纯 workbench 子面板。
- 最小收口面是按 owner 分两层落位：workbench 子面板进 `features/contest-workbench/ui`，编辑页 workspace shell 进 `features/platform-contests/ui`。

## 设计边界

### `features/contest-workbench/ui/*` 本轮负责

- workbench stage tabs
- workbench summary strip
- challenge filter / summary strip
- challenge orchestration panel

### `features/platform-contests/ui/*` 本轮负责

- `ContestEditWorkspacePanel.vue` 这类 route shell 级编辑工作区组合面板

### `features/*/model/*` 本轮继续负责

- stage 计算、query sync、AWD stage 数据、题目编排读写、保存流程

### `ContestEdit.vue` 本轮继续负责

- route 组合
- 顶层 loading overlay
- topbar、stage tabs、workspace panel 的拼装

## 任务切片

### Slice 1：contest-workbench UI 落位

- 目标：
  - 新增 `features/contest-workbench/ui/*`
  - `features/contest-workbench/index.ts` 暴露 UI public API
  - `ContestEditWorkspacePanel.vue` 改从 feature public API 取 workbench 子面板
- 验证：
  - `npm run test:run -- src/components/platform/__tests__/ContestWorkbenchStageTabs.test.ts src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts`
- Review focus：
  - orchestration panel 是否仍只经由 feature model 拿数据
  - child strip 组件是否仍保持纯 props / emits contract

### Slice 2：platform-contests workspace shell 落位

- 目标：
  - 新增 `features/platform-contests/ui/ContestEditWorkspacePanel.vue`
  - `ContestEdit.vue` 改从 `features/platform-contests` / `features/contest-workbench` public API 读取
  - allowlist、类型声明、raw-source 测试同步更新
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ContestEdit.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - `ContestEditWorkspacePanel.vue` 的 owner 是否从 legacy component 目录完整收口
  - `ContestEdit.vue` 是否只保留 route shell 组合职责

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/ContestWorkbenchStageTabs.test.ts src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts src/views/platform/__tests__/ContestEdit.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## 残余风险

- `ContestEditTopbarPanel.vue`、`ContestChallengeEditorDialog.vue`、`AWDChallengeConfigPanel.vue` 等 contest edit 相关面板仍会留在原路径，本轮只先收口当前被 allowlist 冻结的 workbench / workspace shell UI。
