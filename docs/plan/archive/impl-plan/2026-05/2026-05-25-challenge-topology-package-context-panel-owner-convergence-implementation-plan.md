> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ChallengeTopologyStudioPage.vue` 当前 challenge context rail、拓扑页既有 `topology/*` 子组件分层模式
> 替代：无

# Challenge Topology Package Context Panel Owner Convergence Implementation Plan

## 目标

- 把 `ChallengeTopologyStudioPage.vue` challenge 模式右侧 context rail 中的题包上下文卡片抽成独立组件。
- 保持父页面继续拥有导出动作、`packageSourceSummary` 等展示数据的实际来源和页面级上下文排布。
- 让新组件只承接“题包来源 / 题包文件 / 修订历史”卡片模板和导出事件边界。

## 非目标

- 本轮不改题包导出接口、导出成功后的反馈链路或 `useChallengeTopologyStudioPage`。
- 本轮不处理 challenge 模式的画布区、topbar 或 template side panel。
- 本轮不改 package summary 的业务判定逻辑。

## 输入依据

- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyEntryNodeSection.vue`
- `code/frontend/src/components/platform/topology/TopologyStatusNotes.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`

## 当前结论

- challenge 模式右侧 context rail 目前仍有一段纯展示型题包上下文模板，包含题包来源、题包文件和修订历史三张卡片。
- 这段模板与 `draft`、selected canvas state、模板写回、保存删除动作没有 state owner 混合，只有导出按钮需要把动作留在父页。
- 既有 topology debt 收口已经证明：稳定 card cluster 下沉到 `topology/*` 子组件、父页保留 page owner，是当前最小风险路径。

## 任务切片

### Slice 1：抽出题包上下文 panel

- 目标：
  - 新建 `TopologyPackageContextPanel.vue`。
  - 父页改成只传 `packageSourceSummary / packageBaselineSummary / packageFiles / packageRevisionHistory / exporting`，并监听导出事件。
- 预期改动：
  - `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
  - `code/frontend/src/components/platform/topology/TopologyPackageContextPanel.vue`
  - `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- 组件边界：
  - 父页继续拥有：题包 summary 数据来源、导出动作、context rail 的整体布局编排。
  - 子组件只发 `exportPackage`，不直接接触 `draft`、API 或页面级异步 owner。
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/asyncChunkBoundaries.test.ts`
  - `npm run typecheck`
  - `git diff --check`
- Review focus：
  - 父页是否真的只保留 challenge context rail 的组合 owner。
  - 新组件是否只是展示 / 事件转发，而没有继续吸入格式化、导出状态判断之外的页面级逻辑。

### Slice 2：回写当前事实

- 目标：
  - 把本轮 `TD-1` 拓扑页进展写回主索引，继续收窄 backlog。
- 预期改动：
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## 风险

- 如果为了减少 props 数量把 package summary 业务判定一起迁进组件，会把 presentation owner 从 feature/composable 拖回页面或子组件，偏离当前收口方向。
- 题包文件和修订历史列表都带 slice 展示，若抽层时把“省略剩余项”文案写坏，会造成 challenge 页面信息回归。

## 回退方式

- 如抽层回归，可删除 `TopologyPackageContextPanel.vue` 并恢复父页中的三张题包卡片。
- 本轮不涉及 API、route 或保存契约，回退只影响组件层。
