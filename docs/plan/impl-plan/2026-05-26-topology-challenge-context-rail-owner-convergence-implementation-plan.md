> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ChallengeTopologyStudioPage.vue` 当前结构、拓扑页既有 `Topology*Section/Panel` 子组件分层模式
> 替代：无

# Topology Challenge Context Rail Owner Convergence Implementation Plan

## 目标

- 让 `ChallengeTopologyStudioPage.vue` 把 challenge 模式右侧 `context rail` 的装配壳抽成独立子组件。
- 保持父组件继续拥有导出题包、模板搜索 / 载入 / 应用 / 删除、页面级数据和 draft 编辑 owner。
- 让新子组件只承接 challenge 侧栏布局与现有子组件组合。

## 非目标

- 本轮不改 `TopologyStatusNotes.vue`、`TopologyPackageContextPanel.vue`、`TopologyTemplateSidePanel.vue` 内部逻辑。
- 本轮不改 template-library 模式 hero / workbench 的交互 owner。
- 本轮不新建 composable，不改 `useChallengeTopologyStudioPage` 的数据结构。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyStatusNotes.vue`
- `code/frontend/src/components/platform/topology/TopologyPackageContextPanel.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateSidePanel.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `.harness/reuse-decisions/topology-challenge-context-rail-owner-convergence.md`

## 当前结论

- 拓扑页大块编辑区已经基本按 owner-safe 切片收口，剩余明显的模板壳之一是 challenge 模式右侧 `context rail`。
- 该区域已经由三个现成子组件组成，继续留在父页只会维持大段装配模板，不增加 owner 清晰度。
- 最小安全切片是：新组件接收现成值与 template-side-panel 的透传事件，父页继续持有页面级动作和数据 owner。

## 任务切片

### Slice 1：抽出 challenge context rail 子组件

- 目标：
  - 新建 `TopologyChallengeContextRail.vue`，承接 challenge 模式右侧侧栏 wrapper、stack 和三块子组件装配。
  - `ChallengeTopologyStudioPage.vue` 只保留页面级 owner 与 action handler。
- 预期改动：
  - `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
  - `code/frontend/src/components/platform/topology/TopologyChallengeContextRail.vue`
- 组件边界：
  - 父组件继续拥有 `statusCard`、`secondaryCard`、题包上下文数据、模板列表与所有导出 / 模板动作
  - 子组件只接收现成值并透传事件
- 验证：
  - `git diff --check -- code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue code/frontend/src/components/platform/topology/TopologyChallengeContextRail.vue`
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts -t "context rail|题包上下文区|共享 ui-btn 原语"`
- Review focus：
  - 父组件是否继续持有导出题包、模板搜索 / 应用 / 删除 owner
  - 新子组件是否只是组合壳，不新增本地业务状态

### Slice 2：回写 TD-1 进展

- 目标：
  - 把拓扑页剩余页面编排壳的新切片进展写回前端主索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "ChallengeTopologyStudioPage|context rail|TopologyChallengeContextRail" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否清楚说明这是拓扑页剩余页面编排壳中的一刀，而不是改动页面 owner

## 风险

- `TopologyTemplateSidePanel` 的 `v-model` 与动作 emits 较多，透传遗漏会导致模板搜索 / 编辑回归。
- raw source 护栏需要从父页断言转移到新组合壳。
- 如果顺手继续改 template-library 模式，会超出本轮边界。

## 回退方式

- 如 `TopologyChallengeContextRail.vue` 抽层引入回归，可回退新组件并恢复父页 challenge 侧栏模板。
- 本轮只影响前端组件层、测试护栏和文档，不涉及 API、route 或服务端逻辑。
