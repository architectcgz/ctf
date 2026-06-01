> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ChallengeTopologyStudioPage.vue` 当前结构、拓扑页既有 `topology/*` section 抽层模式
> 替代：无

# Challenge Topology Canvas Quick Editor Owner Convergence Implementation Plan

## 目标

- 把 `ChallengeTopologyStudioPage.vue` 中当前仍由父页直接内联的“画布快速编辑”区块抽到独立组件。
- 保持父页面继续拥有 `useChallengeTopologyStudioPage`、拓扑草稿、选中态、路由入口、模板工作流、保存/删除/导出动作。
- 让新组件只承接当前选中节点 / 连线的局部编辑展示与事件发射，不接管远端请求、草稿整体增删改、模板写回或页面级错误策略。

## 非目标

- 本轮不处理 `ChallengeTopologyStudioPage.vue` 的保存动作、模板侧栏、网络分段区、节点编辑区、链路策略区。
- 本轮不修改 `useChallengeTopologyStudioPage.ts`、`useTopologySelectionState.ts`、`useTopologyEdgeEditing.ts` 的核心行为。
- 本轮不把 challenge 模式额外的“网络快速编辑”区块一并抽出，避免把一个稳定切片扩大成整块工作台重排。

## 输入依据

- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `docs/reviews/frontend/README.md`
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateSidePanel.vue`
- `code/frontend/src/components/platform/topology/TopologyNetworkSection.vue`
- `code/frontend/src/components/platform/topology/TopologyNodeSection.vue`
- `code/frontend/src/components/platform/topology/TopologyConnectivitySections.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`

## 当前结论

- `ChallengeTopologyStudioPage.vue` 目前约 `1892` 行，`TD-1` 已经完成五轮拓扑页切片，但 challenge / template 两种模式下的“画布快速编辑”模板仍由父页直接拥有。
- 该区块虽然在两种模式下字段不完全相同，但 owner 边界已经清楚：
  - 父页负责选中对象、拓扑草稿、node/network/link/policy 的真实数据更新。
  - 快速编辑区只负责基于当前选中对象展示输入控件，并把局部变更转发回父页。
- 现有拓扑页已经有 `TopologyTemplateSidePanel.vue`、`TopologyNetworkSection.vue`、`TopologyNodeSection.vue`、`TopologyConnectivitySections.vue` 等抽层先例，本轮应复用同一分层方向。

## 任务切片

### Slice 1：抽出共用的画布快速编辑组件

- 目标：
  - 新建 `TopologyCanvasQuickEditor.vue`，承接 template / challenge 两种模式下的节点 / 连线快速编辑壳层与输入控件。
  - 父页改为只传 props 与绑定 emits，不再直接持有这块模板。
- 预期改动：
  - `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
  - `code/frontend/src/components/platform/topology/TopologyCanvasQuickEditor.vue`
- 组件边界：
  - 父页继续拥有：
    - `draft`
    - `selectedNodeDraft / selectedEdgeMeta`
    - `selectedEdgeSourceKey / selectedEdgeTargetKey / selectedEdgeKind`
    - `updateCanvasQuickNumber / toggleSelectedNodeNetwork`
    - `updateSelectedEdgeSourceKey / updateSelectedEdgeTargetKey / handleSelectedEdgeKindChange`
  - 子组件只接收上面这些现成状态和 handler，不新增 page model 或内部异步逻辑。
- 验证：
  - `git diff --check -- code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue code/frontend/src/components/platform/topology/TopologyCanvasQuickEditor.vue`
  - `npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/__tests__/asyncChunkBoundaries.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Review focus：
  - 父页面是否真正退回到组合 owner，而不是把大段模板简单搬走后仍保留重复的局部样式和条件分支。
  - 新组件是否只承接局部编辑展示，不偷偷吸入草稿 owner、页面动作或远端依赖。

### Slice 2：回写前端技术债事实源

- 目标：
  - 把本轮之后的 `TD-1` 状态和前端技术债快速核查入口写回 review 事实源，减少后续重复扫描。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "快速核查入口|TopologyCanvasQuickEditor|StudentInsightPanel.vue" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否把“已经收口的 touched surface”和“仍然保留的 backlog”明确分开，避免旧结论继续污染活动目录。

## 风险

- `ChallengeTopologyStudio.test.ts` 当前大量依赖源码断言；如果只是抽组件但不调整断言，测试会把合理的 owner 下沉误判成缺失。
- challenge 模式和 template 模式的快速编辑字段不完全一致；若为了复用过度抽象 props，容易把简单的条件渲染变成难维护的巨型配置接口。
- `TopologyCanvasBoard.vue` 仍由父页异步加载；本轮不应顺手改动画布异步边界或页面选中逻辑，否则会把一个模板切片扩大成 page controller 重构。

## 回退方式

- 如快速编辑抽层引入回归，可回退 `TopologyCanvasQuickEditor.vue` 并恢复 `ChallengeTopologyStudioPage.vue` 内联模板。
- 因本轮不涉及 API、route、topology draft schema 或模板保存契约，回退只影响组件层，不涉及数据迁移或兼容清洗。
