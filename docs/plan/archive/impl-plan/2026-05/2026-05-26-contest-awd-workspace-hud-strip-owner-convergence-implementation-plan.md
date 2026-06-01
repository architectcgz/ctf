> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ContestAWDWorkspacePanel.vue` 当前结构、AWD 学员战场既有 `awd/*` 子组件分层模式
> 替代：无

# Contest AWD Workspace HUD Strip Owner Convergence Implementation Plan

## 目标

- 让 `ContestAWDWorkspacePanel.vue` 把顶部 HUD strip 从父组件中抽成独立子组件。
- 保持父组件继续拥有 AWD 工作台数据派生、刷新动作和页面级工作流 owner。
- 让新子组件只承接“当前回合 / 我的战队 / 战队服务 / 最高分 / 刷新状态”展示模板与局部样式。

## 非目标

- 本轮不改 `ContestDetail.vue` 的 tab、路由、页面装配或 feature route model。
- 本轮不处理 `ContestAWDWorkspacePanel.vue` 中区攻击区与左侧防守列。
- 本轮不引入新的 composable、store、feature API 或共享 KPI shell。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `.harness/reuse-decisions/contest-awd-workspace-hud-strip-owner-convergence.md`

## 当前结论

- `ContestAWDWorkspacePanel.vue` 已完成右侧 intelligence rail 抽层，但父组件仍直接内联顶部 HUD strip。
- HUD strip 只读父组件已有派生值，并只触发一个刷新动作，是当前页面剩余区块里最小、最稳定、最不易打散 owner 的切片。
- 相比左侧防守列和中区攻击区，HUD strip 不携带 SSH、重启、复制、目标筛选或 Flag 提交流程，更适合作为下一刀。

## 任务切片

### Slice 1：抽出 HUD strip 子组件

- 目标：
  - 新建 `AWDWorkspaceHudStrip.vue`，承接顶部 KPI strip 模板与样式。
  - `ContestAWDWorkspacePanel.vue` 只保留派生值、状态格式化和刷新 action owner。
- 预期改动：
  - `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - `code/frontend/src/components/contests/awd/AWDWorkspaceHudStrip.vue`
- 组件边界：
  - 父组件继续拥有所有派生值和 `refreshAll`
  - 子组件只接收展示所需 props，并通过 `refresh` emit 交回刷新动作
- 验证：
  - `git diff --check -- code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue code/frontend/src/components/contests/awd/AWDWorkspaceHudStrip.vue`
  - `cd code/frontend && npm run test:run -- src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/views/contests/__tests__/ContestDetail.test.ts -t AWD`
- Review focus：
  - 父组件是否真正退回到页面 owner，而不是继续内联 HUD 样式和模板
  - 新子组件是否保持纯展示，不直接持有 composable 或业务状态

### Slice 2：回写 TD-1 进展

- 目标：
  - 把 HUD strip 切片进展写回前端主索引，继续降低后续重复扫描成本。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "ContestAWDWorkspacePanel|HUD|当前回合|最高分" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否把 HUD strip 进展记成已处理的 touched surface，同时保留剩余攻击区 / 防守编排 backlog

## 风险

- `contestAwdWorkspacePanelSource.test.ts` 当前做源码护栏；抽层后需要改为“父组件挂载新 HUD 子组件 + 关键 HUD 文案位于新子组件”，否则会把合理抽层误判成退化。
- HUD strip 当前与父组件共用局部 class；若样式迁移不完整，会造成顶部 KPI 栅格回归。
- 若顺手把左侧防守列一起抽，会把当前切片从“纯展示区块收口”扩大成带动作 owner 的大重排，超出边界。

## 回退方式

- 如 HUD strip 抽层引入回归，可回退 `AWDWorkspaceHudStrip.vue` 并恢复 `ContestAWDWorkspacePanel.vue` 的顶部内联模板。
- 本轮不涉及 API、route、feature model、攻击或防守动作，回退只影响前端组件层与相关测试。
