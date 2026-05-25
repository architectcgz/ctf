> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ContestAWDWorkspacePanel.vue` 当前结构、AWD 学员战场既有 `awd/*` 子组件分层模式
> 替代：无

# Contest AWD Workspace Defense Alerts Owner Convergence Implementation Plan

## 目标

- 让 `ContestAWDWorkspacePanel.vue` 把左侧防守列顶部的 `defenseAlerts` 告警列表抽成独立子组件。
- 保持父组件继续拥有 `defenseAlerts` 计算、服务选择、SSH/服务动作和防守工作台整体 owner。
- 让新子组件只承接告警列表模板与局部样式，不接管服务动作或页面刷新流程。

## 非目标

- 本轮不改 `ContestDetail.vue` 的路由、页签、页面装配或 feature model。
- 本轮不处理中区攻击区 owner，也不重排左侧服务列表和防守操作面板之间的关系。
- 本轮不新增 composable、store、feature API 或新的防守动作组件。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `.harness/reuse-decisions/contest-awd-workspace-defense-alerts-owner-convergence.md`

## 当前结论

- `ContestAWDWorkspacePanel.vue` 已完成 HUD strip 和右侧 intelligence rail 抽层，剩余 touched surface 中最稳定的展示块是左侧顶部 `defenseAlerts`。
- `defenseAlerts` 本身仍由父组件按 `runtimeChallenges` 和 `workspace service` 派生，这部分 owner 不应下沉。
- 当前最小安全切片是把“告警展示模板 + 局部样式”抽成一个独立子组件，让父组件保留派生和动作协调。

## 任务切片

### Slice 1：抽出 defense alerts 子组件

- 目标：
  - 新建 `AWDDefenseAlertsPanel.vue`，承接左侧顶部告警列表模板与样式。
  - `ContestAWDWorkspacePanel.vue` 只保留 `defenseAlerts` 计算和子组件装配。
- 预期改动：
  - `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - `code/frontend/src/components/contests/awd/AWDDefenseAlertsPanel.vue`
- 组件边界：
  - 父组件继续拥有 `defenseAlerts` 的计算逻辑
  - 子组件只接收已归一化好的告警数组，不直接感知 challenge/service 原始数据
- 验证：
  - `git diff --check -- code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue code/frontend/src/components/contests/awd/AWDDefenseAlertsPanel.vue`
  - `cd code/frontend && npm run test:run -- src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/views/contests/__tests__/ContestDetail.test.ts -t AWD`
- Review focus：
  - 父组件是否只退出了展示模板，而没有把告警计算 owner 一并拆散
  - 新子组件是否仍保持纯展示，不吸入服务操作或刷新事件

### Slice 2：回写 TD-1 进展

- 目标：
  - 把左侧防守告警切片写回前端主索引，减少后续重复扫描。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "ContestAWDWorkspacePanel|defense alert|防守告警|我的防守" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否清楚说明本轮只收口了左侧告警展示，而左侧服务编排和中区攻击 owner 仍在 backlog

## 风险

- `contestAwdWorkspacePanelSource.test.ts` 目前依赖 raw source 护栏；抽层后需要同步断言到新告警组件，避免把展示抽层误判成结构缺失。
- 告警样式当前和父组件共处一处 scoped style；若样式迁移不完整，左侧防守区首屏视觉会回退。
- 如果这轮继续向下碰 `AWDDefenseServiceList` 或 `AWDDefenseOperationsPanel` 的交互 owner，会超出“纯展示切片”的边界。

## 回退方式

- 如 `AWDDefenseAlertsPanel.vue` 抽层引入回归，可回退新文件并恢复父组件的告警模板。
- 本轮不涉及 API、route、服务动作和 SSH workflow，回退只影响前端组件层与测试护栏。
