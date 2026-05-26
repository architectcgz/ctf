> 状态：Current
> 事实源：`ContestAWDWorkspacePanel.vue` 当前脚本结构、AWD 工作台既有 feature/composable 分层、`TD-1` 超大组件专题拆分进度
> 替代：无

# AWD Workspace Summary Owner Convergence Implementation Plan

## 目标

- 把 `ContestAWDWorkspacePanel.vue` 里剩余的 HUD 摘要和防守告警派生收口到独立 composable。
- 让父页继续只保留工作区数据获取、攻击提交流程、防守操作装配和三栏布局 owner。
- 为 AWD 工作台补源码护栏，防止 summary / alert presentation 再回流到父页。

## 非目标

- 本轮不改 `AWDWorkspaceHudStrip.vue`、`AWDDefenseColumn.vue` 和 `AWDWorkspaceIntelColumn.vue` 的模板结构。
- 本轮不改 `useContestAWDWorkspace.ts` 的请求、刷新、轮询或副作用 owner。
- 本轮不再重新拆分攻击向量和防守 access 逻辑。

## 输入依据

- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceHudStrip.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseColumn.vue`
- `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspacePresentation.ts`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `.harness/reuse-decisions/awd-workspace-summary-owner-convergence.md`

## 当前结论

- AWD 工作台前几刀已经把 HUD strip、情报列、防守列装配壳、攻击列装配壳、攻击向量局部 state、防守 access 局部 state、以及情报与结果文案都从父页拆开。
- 当前父页剩余最成组的展示脚本是 summary / alert presentation：回合与同步文案、排名/服务数/最高分派生，以及 `defenseAlerts` 这组防守告警。
- 这组逻辑不持有副作用，也不需要继续留在 route-level parent；最小安全切片是新建 `useAwdWorkspaceSummary.ts` 收口，并补 source guard。

## 任务切片

### Slice 1：抽出 workspace summary composable

- 目标：
  - 新建 `useAwdWorkspaceSummary.ts`，收口 HUD 摘要文案、排名/统计派生和 `defenseAlerts`。
  - 父页只消费 composable 返回值，不再本地定义 `defenseAlerts`、`formatRoundStatusLabel` 等 helper。
- 预期改动：
  - `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceSummary.ts`
  - `code/frontend/src/features/contest-awd-workspace/model/index.ts`
  - `code/frontend/src/features/contest-awd-workspace/index.ts`
- 边界：
  - composable 只处理基于现成数据的派生展示，不发请求、不持有副作用。
  - 父页继续保留 `useContestAWDWorkspace`、`useAwdWorkspaceAttackVector`、`useAwdDefenseAccessPanel` 等 workflow owner。
  - 防守告警仍只消费 AWD runtime challenge 语义，不回退到历史 `challenge_id`。
- 验证：
  - `git diff --check -- code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceSummary.ts code/frontend/src/features/contest-awd-workspace/model/index.ts code/frontend/src/features/contest-awd-workspace/index.ts`
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-workspace/model/useAwdWorkspaceSummary.test.ts`
- Review focus：
  - summary composable 是否只承接展示派生，没有把副作用重新迁进去
  - `defenseAlerts` 是否继续保持 AWD runtime challenge 身份边界

### Slice 2：补测试护栏与 review 进度

- 目标：
  - 给新 composable 增加摘要与告警派生测试。
  - 更新 source guard 和 frontend review 索引，记录 AWD 页面 `TD-1` 当前 touched surface 已收口。
- 预期改动：
  - `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceSummary.test.ts`
  - `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-workspace/model/useAwdWorkspaceSummary.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
  - `cd code/frontend && npm run typecheck`
  - `bash scripts/check-consistency.sh`
- Review focus：
  - source guard 是否明确限制父页不再持有 summary / alert presentation
  - review 文档是否把 AWD 页面 `TD-1` 进度同步为已收口

## 风险

- `defenseAlerts` 文案和严重等级不能因为抽层而变化。
- HUD 摘要里的回合同步文案、排名、服务数和最高分不能改变现有展示口径。
- 如果 composable 输入边界过宽，父页 owner 会被重新拉散；因此只传现成 workspace / scoreboard / runtime challenge / service map。

## 回退方式

- 如新 composable 引入回归，可回退 `useAwdWorkspaceSummary.ts` 并恢复父页本地 summary / alert 派生。
- 本轮只影响前端 feature/composable、父页装配、源码护栏和 review 文档，不涉及 API、route 或服务端逻辑。
