> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ChallengeTopologyStudioPage.vue` 当前结构、拓扑页既有 `topology/*` 抽层模式
> 替代：无

# Challenge Topology Network Quick Editor Owner Convergence Implementation Plan

## 目标

- 把 `ChallengeTopologyStudioPage.vue` 中 challenge-only 的“网络快速编辑”区块抽成独立组件。
- 保持父页面继续拥有 `draft.networks`、`updateNetworkDraft`、拓扑保存 / 删除动作和页面级布局 owner。
- 让新组件只承接网络快速编辑的模板、局部输入控件和 patch 事件发射。

## 非目标

- 本轮不处理 `TopologyNetworkSection.vue` 的主网络分段编辑区，也不合并 quick editor 与主编辑区。
- 本轮不处理“入口节点”卡片、右侧 context rail、模板库模式布局或 `useChallengeTopologyStudioPage.ts`。
- 本轮不引入新的异步逻辑、store 或 API 调用。

## 输入依据

- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasQuickEditor.vue`
- `code/frontend/src/components/platform/topology/TopologyNetworkSection.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`

## 当前结论

- `ChallengeTopologyStudioPage.vue` 当前约 `1693` 行，画布快速编辑已在上一轮切片中抽出，但 challenge-only 模式下与之并列的“网络快速编辑”仍由父页内联持有。
- 该区块只负责渲染 `draft.networks` 的快速表单，字段为 `key / name / internal`，属于局部编辑壳，不应继续占住父页模板体积。
- 仓库里已有 `TopologyCanvasQuickEditor.vue` 这类“父页持有草稿 owner，子组件只 emit patch”的最新收口先例，本轮应直接复用这一模式。

## 任务切片

### Slice 1：抽出网络快速编辑组件

- 目标：
  - 新建 `TopologyNetworkQuickEditor.vue`，承接 challenge-only 的网络快速编辑表单。
  - 父页改成只传 `draft.networks` 和 `updateNetworkDraft`，不再直接内联这块模板。
- 预期改动：
  - `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
  - `code/frontend/src/components/platform/topology/TopologyNetworkQuickEditor.vue`
  - `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- 组件边界：
  - 父页继续拥有：`draft.networks`、`updateNetworkDraft`、页面布局、保存 / 删除 / 导出动作。
  - 子组件只接收 `networks`，发出 `{ uid, patch }`，不直接写 `draft`、不接远端依赖、也不决定页面级错误处理。
- 验证：
  - `git diff --check -- code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue code/frontend/src/components/platform/topology/TopologyNetworkQuickEditor.vue`
  - `npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/asyncChunkBoundaries.test.ts`
  - `npm run typecheck`
- Review focus：
  - 父页是否真正退成组合 owner，而不是只把模板文字搬到新文件。
  - 新组件是否仍保持局部 quick editor 边界，没有把主网络编辑 owner 混进去。

### Slice 2：回写 TD-1 当前事实

- 目标：
  - 把本轮 `TD-1` 拓扑页进展写回前端主索引，收窄后续 backlog 读取成本。
- 预期改动：
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
- Review focus：
  - 文档是否明确区分“已完成的拓扑页切片”和“仍然保留的剩余债面”。

## 风险

- `ChallengeTopologyStudio.test.ts` 当前同时做渲染断言和源码边界断言；如果不更新源码断言，合理的 owner 下沉会继续被测试误判为回归。
- 如果把 `draft.networks` 的直接可变结构下沉到子组件内部处理，容易让 quick editor 变成第二个网络 owner，和 `TopologyNetworkSection.vue` 形成并行写入面。

## 回退方式

- 如 quick editor 抽层引入回归，可删除 `TopologyNetworkQuickEditor.vue` 并恢复父页内联模板。
- 因本轮不涉及 API、route、draft schema 与保存契约，回退只影响前端组件层。
