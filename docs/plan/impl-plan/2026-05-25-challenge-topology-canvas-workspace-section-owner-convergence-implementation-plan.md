> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ChallengeTopologyStudioPage.vue` 当前双模态画布壳、拓扑页既有 `topology/*` 子组件分层模式
> 替代：无

# Challenge Topology Canvas Workspace Section Owner Convergence Implementation Plan

## 目标

- 把 `ChallengeTopologyStudioPage.vue` 中模板库模式和 challenge 模式的“图形画布”工作区收成独立组件。
- 保持父页面继续拥有 `interactionMode`、selected canvas state、`draftValidationIssues`、`canvasGraph` 和所有画布更新动作。
- 让新组件只承接画布 section 模板、模式按钮和 quick editor 组合展示。

## 非目标

- 本轮不改 `useChallengeTopologyStudioPage`、画布数据结构或节点/边编辑语义。
- 本轮不处理 template side panel、challenge topbar、template toolbar tabs。
- 本轮不重构 `TopologyCanvasBoard.vue` 或 `TopologyCanvasQuickEditor.vue` 的内部实现。

## 输入依据

- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasBoard.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasQuickEditor.vue`
- `code/frontend/src/components/platform/topology/TopologyNetworkQuickEditor.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `code/frontend/src/views/__tests__/asyncChunkBoundaries.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`

## 当前结论

- 当前页面还内联着两段大体重复的“图形画布” section，已经成为 `ChallengeTopologyStudioPage.vue` 剩余模板体积的主要来源。
- 两段 section 的真正 owner 都在父页：画布模式、选中态、快速编辑动作、节点拖拽位置更新和校验提示都不应该继续下沉到 feature/composable 之外的第二个 owner。
- 因此可以把这块收成一个支持 `template / challenge` variant 的 section 组件，用 props + emits 明确边界。

## 任务切片

### Slice 1：抽出画布工作区 section

- 目标：
  - 新建 `TopologyCanvasWorkspaceSection.vue`。
  - 父页改成只传画布展示数据、selected canvas state 和更新动作。
- 预期改动：
  - `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
  - `code/frontend/src/components/platform/topology/TopologyCanvasWorkspaceSection.vue`
  - `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - `code/frontend/src/views/__tests__/asyncChunkBoundaries.test.ts`
- 组件边界：
  - 父页继续拥有：`interactionMode`、selected node/edge state、`draftValidationIssues`、`canvasGraph`、节点位置更新和 quick editor 更新动作。
  - 子组件只发模式切换、画布事件和 quick editor 事件，不直接接触 route/query、API 或保存/导出动作。
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/asyncChunkBoundaries.test.ts`
  - `npm run typecheck`
  - `git diff --check`
- Review focus：
  - 父页是否仍然是唯一的 canvas state owner。
  - variant 差异是否只落在模板文案和布局，而不是再造第二套组件逻辑。

### Slice 2：回写当前事实

- 目标：
  - 把本轮 `TD-1` 拓扑页进展写回主索引，继续收窄 backlog。
- 预期改动：
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## 风险

- 画布工作区 props / emits 比前几刀更多，若接口设计混乱，容易把 owner 边界写糊。
- 如果新组件继续直接引入 feature model 内部类型，可能重新踩到前面 architecture allowlist 的边界测试。

## 回退方式

- 如抽层回归，可删除 `TopologyCanvasWorkspaceSection.vue` 并恢复父页中的两段“图形画布” section。
- 本轮不涉及 API、route 或保存契约，回退只影响组件层。
