> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ChallengeTopologyStudioPage.vue` 当前样式分布、challenge 已抽出的 header / workbench / canvas 结构
> 替代：无

# Topology Challenge Style Owner Convergence Implementation Plan

## 目标

- 把 challenge 模式下的 header / workbench / canvas 样式 owner 从 `ChallengeTopologyStudioPage.vue` 收回到对应子组件。
- 保持父页只保留页面壳主题变量、模式容器样式和加载 / 空状态分支。
- 让 challenge 模式的结构 owner 和样式 owner 一致。

## 非目标

- 本轮不改 `useChallengeTopologyStudioPage`、拓扑数据流或任何业务逻辑。
- 本轮不改 template-library 模式结构和样式。
- 本轮不继续新增结构组件。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeWorkspaceHeader.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeWorkbench.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasWorkspaceSection.vue`
- `.harness/reuse-decisions/topology-challenge-style-owner-convergence.md`

## 当前结论

- challenge 结构壳已经基本拆完，父页里剩余最重的是一批还没跟着迁走的样式。
- 这些样式已经有明确的组件 owner，继续留在父页只会维持 page shell 和子组件壳之间的跨组件样式耦合。
- 最小安全切片是：只迁移样式 owner，不碰模板结构和事件 contract。

## 任务切片

### Slice 1：收回 challenge header / workbench / canvas 样式 owner

- 目标：
  - `TopologyChallengeWorkspaceHeader.vue` 持有 topbar / heading / challenge action button 的样式。
  - `TopologyChallengeWorkbench.vue` 持有主工作区栅格、context rail、section card、输入皮肤和嵌套 action button 皮肤。
  - `TopologyCanvasWorkspaceSection.vue` 持有 allow 模式按钮和 validation banner 样式。
- 预期改动：
  - `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
  - `code/frontend/src/components/platform/topology/TopologyChallengeWorkspaceHeader.vue`
  - `code/frontend/src/components/platform/topology/TopologyChallengeWorkbench.vue`
  - `code/frontend/src/components/platform/topology/TopologyCanvasWorkspaceSection.vue`
- 验证：
  - `git diff --check -- code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue code/frontend/src/components/platform/topology/TopologyChallengeWorkspaceHeader.vue code/frontend/src/components/platform/topology/TopologyChallengeWorkbench.vue code/frontend/src/components/platform/topology/TopologyCanvasWorkspaceSection.vue`
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
  - `cd code/frontend && npm run typecheck`
- Review focus：
  - 样式 owner 是否和 challenge 结构边界一致
  - 父页是否明显减少对 challenge 子组件的跨组件样式覆盖

### Slice 2：回写 TD-1 进展

- 目标：
  - 把 challenge 样式 owner 收口进展写回主索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "challenge 样式 owner|TopologyChallengeWorkbench|TopologyChallengeWorkspaceHeader" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否明确这是样式 owner 收口，而不是新结构拆分

## 风险

- challenge workbench 的深层 `:deep` 样式迁移后，嵌套输入和 section card 的皮肤容易漏掉。
- header 和 workbench 都有 `topology-action-btn`，迁移时要避免按钮变量定义一边丢失。

## 回退方式

- 如迁移后出现样式回归，可回退对应子组件样式 block，并恢复父页里的 challenge 覆盖规则。
- 本轮只影响前端组件样式层、测试护栏和 review 文档，不涉及 API、route 或服务端逻辑。
