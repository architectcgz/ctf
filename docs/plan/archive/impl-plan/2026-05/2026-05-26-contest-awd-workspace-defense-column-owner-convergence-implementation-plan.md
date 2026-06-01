> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ContestAWDWorkspacePanel.vue` 当前结构、AWD 学员战场既有 `awd/*` 子组件分层模式
> 替代：无

# Contest AWD Workspace Defense Column Owner Convergence Implementation Plan

## 目标

- 让 `ContestAWDWorkspacePanel.vue` 把左侧防守列的装配壳抽成独立子组件。
- 保持父组件继续拥有服务选择、SSH / 复制、刷新、重启动作和相关派生 owner。
- 让新子组件只承接“我的防守”列布局与现有子组件组合。

## 非目标

- 本轮不改 `AWDDefenseServiceList.vue`、`AWDDefenseOperationsPanel.vue` 内部动作逻辑。
- 本轮不改中区攻击流程、结果提示、路由或 feature model。
- 本轮不新建 composable 或 store。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseAlertsPanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `.harness/reuse-decisions/contest-awd-workspace-defense-column-owner-convergence.md`

## 当前结论

- 当前 `ContestAWDWorkspacePanel.vue` 的展示块已经基本收口，剩余明显的大段模板就是左侧防守列装配壳。
- 这块已经由三个现成子组件组成，因此再拆内部逻辑收益低，应该直接抽布局壳。
- 最小安全切片是：子组件接收现成 props 和动作 emits，父组件继续持有 workflow owner。

## 任务切片

### Slice 1：抽出 defense column 子组件

- 目标：
  - 新建 `AWDDefenseColumn.vue`，承接左侧列 wrapper、header、滚动区和三块子组件装配。
  - `ContestAWDWorkspacePanel.vue` 只保留数据 owner 和事件 handler。
- 预期改动：
  - `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - `code/frontend/src/components/contests/awd/AWDDefenseColumn.vue`
- 组件边界：
  - 父组件继续拥有 `defenseAlerts`、`defenseServiceCards`、`selectedServiceId`、`selectedDefenseServiceCard`、SSH / copy / refresh / restart actions
  - 子组件只接收现成值并透传事件
- 验证：
  - `git diff --check -- code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue code/frontend/src/components/contests/awd/AWDDefenseColumn.vue`
  - `cd code/frontend && npm run test:run -- src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/views/contests/__tests__/ContestDetail.test.ts -t AWD`
- Review focus：
  - 父组件是否继续持有动作 owner，而不是把 open/restart/ssh/copy 下沉进新列组件
  - 新列组件是否只是组合壳，不新建本地业务状态

### Slice 2：回写 TD-1 进展

- 目标：
  - 把左侧防守列收口进展写回前端主索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "ContestAWDWorkspacePanel|防守列|DefenseColumn|我的防守" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否清楚说明 `ContestAWDWorkspacePanel.vue` 的大模板 surface 已基本收口

## 风险

- raw source 护栏需要把“我的防守”这一层级结构迁到新列组件。
- 这块 props 较多，若透传契约遗漏，可能引入打开服务、SSH、复制状态回归。
- 若顺手调整 `AWDDefenseServiceList.vue` 或 `AWDDefenseOperationsPanel.vue` 内部逻辑，会超出本轮边界。

## 回退方式

- 如 `AWDDefenseColumn.vue` 抽层引入回归，可回退新组件并恢复父组件的左列模板。
- 本轮只影响前端组件层、测试护栏和文档，不涉及 API、route 或服务端逻辑。
