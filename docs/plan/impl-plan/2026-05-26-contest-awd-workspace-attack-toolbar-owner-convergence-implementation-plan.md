> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ContestAWDWorkspacePanel.vue` 当前结构、AWD 学员战场既有 `awd/*` 子组件分层模式
> 替代：无

# Contest AWD Workspace Attack Toolbar Owner Convergence Implementation Plan

## 目标

- 让 `ContestAWDWorkspacePanel.vue` 把中区顶部筛选条抽成独立子组件。
- 保持父组件继续拥有 `activeChallengeKey`、`targetKeyword`、过滤结果、目标列表和 Flag 提交流程 owner。
- 让新子组件只承接“目标题目 / 队伍筛选”输入模板与局部样式。

## 非目标

- 本轮不改攻击卡片列表、打开目标按钮、Flag 输入和提交按钮。
- 本轮不改 `ContestDetail.vue`、路由、feature model 或远端请求逻辑。
- 本轮不顺手处理左侧服务编排壳。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `.harness/reuse-decisions/contest-awd-workspace-attack-toolbar-owner-convergence.md`

## 当前结论

- `ContestAWDWorkspacePanel.vue` 仍保留中区顶部筛选条内联模板和样式。
- 这段 toolbar 只消费 `runtimeChallenges`、`activeChallengeKey`、`targetKeyword`，并通过输入回写父组件状态，不碰攻击动作 owner。
- 相比直接下刀目标卡片或提交区，这一段是当前最小且最稳定的继续切片。

## 任务切片

### Slice 1：抽出 attack toolbar 子组件

- 目标：
  - 新建 `AWDAttackToolbar.vue`，承接中区顶部筛选条模板与样式。
  - `ContestAWDWorkspacePanel.vue` 只保留状态 owner，并通过 props / emits 连接子组件。
- 预期改动：
  - `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - `code/frontend/src/components/contests/awd/AWDAttackToolbar.vue`
- 组件边界：
  - 父组件继续拥有 `activeChallengeKey`、`targetKeyword` 和 `runtimeChallenges`
  - 子组件只接收可选 challenge 列表与当前值，通过 update emit 请求父组件改值
- 验证：
  - `git diff --check -- code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue code/frontend/src/components/contests/awd/AWDAttackToolbar.vue`
  - `cd code/frontend && npm run test:run -- src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/views/contests/__tests__/ContestDetail.test.ts -t AWD`
- Review focus：
  - 父组件是否保持攻击工作流 owner，没有把目标打开 / Flag 提交状态一并下沉
  - 新 toolbar 子组件是否只承接输入壳，不引入本地派生或业务副作用

### Slice 2：回写 TD-1 进展

- 目标：
  - 把攻击筛选条切片进展写回前端主索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "ContestAWDWorkspacePanel|attack toolbar|目标题目|队伍筛选" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否清楚说明只抽了顶部筛选条，攻击卡片和提交区仍在 backlog

## 风险

- `contestAwdWorkspacePanelSource.test.ts` 依赖 raw source 护栏；抽层后需要把相关断言迁到新 toolbar 子组件。
- 若输入 props / emits 契约不清晰，可能引入选中题目或关键字同步回归。
- 如果这轮顺手把目标列表也抽出去，会跨过“输入壳”边界，放大 review 面。

## 回退方式

- 如 `AWDAttackToolbar.vue` 抽层引入回归，可回退新组件并恢复父组件的 toolbar 模板。
- 本轮只影响前端组件层、raw source 护栏和文档事实源，不涉及 API、route 或提交动作。
