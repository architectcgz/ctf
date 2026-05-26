> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ChallengeTopologyStudioPage.vue` 当前结构、拓扑页既有 `Topology*Section` / `TopologyChallengeContextRail` 分层模式
> 替代：无

# Topology Challenge Workbench Owner Convergence Implementation Plan

## 目标

- 让 `ChallengeTopologyStudioPage.vue` 把 challenge 模式下的主工作区装配壳抽成独立子组件。
- 保持父组件继续拥有 selection、draft、模板动作、导出/删除/保存等页面级 owner。
- 让新子组件只承接 challenge 模式左侧主列和右侧 `context rail` 的布局装配与事件透传。

## 非目标

- 本轮不改 `useChallengeTopologyStudioPage` 的业务逻辑。
- 本轮不改 `TopologyCanvasWorkspaceSection.vue`、`TopologyEntryNodeSection.vue`、`TopologyNetworkSection.vue`、`TopologyNodeSection.vue`、`TopologyConnectivitySections.vue`、`TopologyChallengeContextRail.vue` 的内部实现。
- 本轮不改 template-library 模式的 hero / workbench / header。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateWorkbench.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeContextRail.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `.harness/reuse-decisions/topology-challenge-workbench-owner-convergence.md`

## 当前结论

- challenge 模式的右侧 `context rail` 已经是独立组件，左侧画布和编辑区也都由稳定子组件组成。
- 父页当前主要还是在拼接一段 mode-specific workbench 模板，不再承担新的 owner 价值。
- 最小安全切片是：新组件接收现成状态与事件，并负责主工作区布局；父页继续持有所有页面级动作和业务状态。

## 任务切片

### Slice 1：抽出 challenge workbench 子组件

- 目标：
  - 新建 `TopologyChallengeWorkbench.vue`，承接 challenge 模式主工作区的布局组合。
  - `ChallengeTopologyStudioPage.vue` 不再直接内联画布、入口节点、网络、节点、策略和 `context rail` 组合模板。
- 预期改动：
  - `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
  - `code/frontend/src/components/platform/topology/TopologyChallengeWorkbench.vue`
- 组件边界：
  - 父组件继续拥有 selection、draft、模板动作、删除拓扑、导出题包等 owner
  - 子组件只接收现成值并透传已有事件
- 验证：
  - `git diff --check -- code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue code/frontend/src/components/platform/topology/TopologyChallengeWorkbench.vue`
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts -t "challenge workbench|context rail|画布工作区"`
- Review focus：
  - 页面级 owner 是否仍留在父页
  - 新组件是否只承接 challenge 模式装配壳，没有吸入业务逻辑

### Slice 2：回写 TD-1 进展

- 目标：
  - 把 challenge workbench 切片进展写回主索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "TopologyChallengeWorkbench|challenge workbench|context rail" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否清楚说明这是 challenge 模式装配壳收口

## 风险

- 新组件 props / emits 较多，遗漏会导致画布交互、模板动作或删除拓扑回归。
- 如果把数据 owner 一并下沉，会超出本轮边界。

## 回退方式

- 如 `TopologyChallengeWorkbench.vue` 抽层引入回归，可回退新组件并恢复父页 challenge 模式模板。
- 本轮只影响前端组件层、测试护栏和文档，不涉及 API、route 或服务端逻辑。
