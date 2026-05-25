> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ChallengeTopologyStudioPage.vue` 当前结构、拓扑页既有 `topology/*` 子组件分层模式
> 替代：无

# Challenge Topology Entry Node Section Owner Convergence Implementation Plan

## 目标

- 把 `ChallengeTopologyStudioPage.vue` 中重复的“入口节点”卡片抽成独立组件。
- 保持父页面继续拥有 `draft.entry_node_key`、`nodeOptions`、`saving`、`topology` 和删除拓扑动作。
- 让新组件只承接入口节点选择和可选删除按钮的模板 / emit 边界。

## 非目标

- 本轮不改变入口节点的业务语义、校验逻辑或删除拓扑动作的实现。
- 本轮不处理右侧 rail、题包来源卡片或模板侧栏。
- 本轮不改 `useChallengeTopologyStudioPage.ts`。

## 输入依据

- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasQuickEditor.vue`
- `code/frontend/src/components/platform/topology/TopologyNetworkQuickEditor.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`

## 当前结论

- `ChallengeTopologyStudioPage.vue` 当前仍保留两份“入口节点”卡片模板，模板库模式和挑战模式只在删除按钮是否显示上有差异。
- 该区块属于稳定的局部交互 card，不需要继续占住父页模板体积。
- 既有 topology debt 收口已经证明：把稳定 card/quick editor 抽到 `topology/*` 子组件，同时让父页保留 page owner，是当前最低风险路径。

## 任务切片

### Slice 1：抽出入口节点 section

- 目标：
  - 新建 `TopologyEntryNodeSection.vue`。
  - 父页改成只传 `entryNodeKey / nodeOptions / showDeleteAction / deleteDisabled`，并监听更新与删除事件。
- 预期改动：
  - `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
  - `code/frontend/src/components/platform/topology/TopologyEntryNodeSection.vue`
  - `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- 组件边界：
  - 父页继续拥有：`draft.entry_node_key`、删除拓扑动作、保存态与 topology 是否存在的判断。
  - 子组件只发 `updateEntryNodeKey` 与 `deleteTopology`，不直接接触 `draft`、API 或页面级布局 owner。
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/asyncChunkBoundaries.test.ts`
  - `npm run typecheck`
  - `git diff --check`
- Review focus：
  - 父页是否真的只保留组合 owner。
  - 新组件是否保持模板库 / 挑战模式的轻差异，而不是引入新的 page 级条件分支。

### Slice 2：回写当前事实

- 目标：
  - 把本轮 `TD-1` 拓扑页进展写回主索引，继续收窄 backlog。
- 预期改动：
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`

## 风险

- 这块模板同时存在于两种模式下，若 props 设计过宽，容易把简单差异包装成难懂的组件接口。
- 如果用 `v-model` 直接把整个 `draft` 下沉，会把子组件变成新的草稿 owner，这不符合当前收口方向。

## 回退方式

- 如抽层回归，可删除 `TopologyEntryNodeSection.vue` 并恢复父页内联卡片。
- 本轮不涉及 API、route 或保存契约，回退只影响组件层。
