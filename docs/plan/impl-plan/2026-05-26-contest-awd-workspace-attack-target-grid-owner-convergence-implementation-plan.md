> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ContestAWDWorkspacePanel.vue` 当前结构、AWD 学员战场既有 `awd/*` 子组件分层模式
> 替代：无

# Contest AWD Workspace Attack Target Grid Owner Convergence Implementation Plan

## 目标

- 让 `ContestAWDWorkspacePanel.vue` 把中区目标卡片列表抽成独立子组件。
- 保持父组件继续拥有 `flagInputs`、`openingTargetKey`、`submittingKey`、`buildAttackStateKey`、`handleSubmit`、`openTarget` 和结果提示 owner。
- 让新子组件只承接目标卡片模板、Flag 输入壳和动作意图 emit。

## 非目标

- 本轮不改攻击结果 footer，不改 `submitAttack` 的 owner。
- 本轮不改左侧服务列表和防守操作编排。
- 本轮不改 `ContestDetail.vue`、路由、feature model 或请求层。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `.harness/reuse-decisions/contest-awd-workspace-attack-target-grid-owner-convergence.md`

## 当前结论

- `ContestAWDWorkspacePanel.vue` 已经把 HUD、情报栏、防守告警和攻击筛选条拆出，当前剩余的中区大块是目标卡片列表。
- 这块同时包含模板体量和多个父级 owner 绑定，是下一刀应该收口但不能把 owner 一并搬走的表面。
- 最小安全切片是：子组件接收现成目标数组、状态值和 flag map，通过事件把打开目标、更新 flag、提交动作上抛给父组件。

## 任务切片

### Slice 1：抽出 attack target grid 子组件

- 目标：
  - 新建 `AWDAttackTargetGrid.vue`，承接目标卡片列表和局部样式。
  - `ContestAWDWorkspacePanel.vue` 保留状态 owner 和动作 handler。
- 预期改动：
  - `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - `code/frontend/src/components/contests/awd/AWDAttackTargetGrid.vue`
- 组件边界：
  - 父组件继续拥有 `filteredTargets`、`flagInputs` 和所有动作 handler
  - 子组件只消费 props，并通过 `open-target`、`update-flag`、`submit` emit 返回意图
- 验证：
  - `git diff --check -- code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue code/frontend/src/components/contests/awd/AWDAttackTargetGrid.vue`
  - `cd code/frontend && npm run test:run -- src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/views/contests/__tests__/ContestDetail.test.ts -t AWD`
- Review focus：
  - 父组件是否保持攻击工作流 owner，而不是把 submit/open target 下沉进子组件
  - 子组件的 props / emits 是否足够直接，避免引入新的隐藏状态

### Slice 2：回写 TD-1 进展

- 目标：
  - 把目标卡片列表切片写回前端主索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "ContestAWDWorkspacePanel|target grid|Flag|打开目标" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否明确当前只收口了目标卡片列表，攻击动作 owner 仍留在父组件

## 风险

- raw source 护栏需要同步迁移到新子组件，否则会把合理抽层误判成缺失。
- 这块首次触及父组件动作 owner 周边，若 emit 契约设计不清晰，容易引入 flag 输入或按钮禁用态回归。
- 如果顺手把结果 footer 或空态判断也一起迁走，会扩大这轮切片边界。

## 回退方式

- 如 `AWDAttackTargetGrid.vue` 抽层引入回归，可回退新组件并恢复父组件的卡片列表模板。
- 本轮仍只影响前端组件层、测试护栏和文档，不涉及 API、route 或服务侧行为。
