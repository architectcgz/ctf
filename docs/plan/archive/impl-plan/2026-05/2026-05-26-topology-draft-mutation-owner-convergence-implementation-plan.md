> 状态：Current
> 事实源：`ChallengeTopologyStudioPage.vue` 当前脚本结构、拓扑工作台既有 feature/model 分层、`TD-1` 超大组件专题拆分进度
> 替代：无

# Topology Draft Mutation Owner Convergence Implementation Plan

## 目标

- 把 `ChallengeTopologyStudioPage.vue` 里剩余的本地 `draft` 变更 helper 收口到 feature model。
- 扩展 `useTopologyStructureMutations.ts`，让它统一承接拓扑结构的增删改 owner。
- 让页面继续只保留 route-level 装配、模板分支与样式壳层，不再直接改 `draft.value`。

## 非目标

- 本轮不重排 `TopologyTemplateWorkbench.vue`、`TopologyChallengeContextRail.vue` 或其他已拆出的展示组件。
- 本轮不改 `useChallengeTopologyStudioPage.ts` 里的加载、保存、导出、模板应用与选择 owner。
- 本轮不处理拓扑页样式体量和 scoped CSS 分拆。

## 输入依据

- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/features/challenge-topology-studio/model/useChallengeTopologyStudioPage.ts`
- `code/frontend/src/features/challenge-topology-studio/model/useTopologyStructureMutations.ts`
- `code/frontend/src/features/challenge-topology-studio/model/useTopologySelectionState.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `.harness/reuse-decisions/topology-draft-mutation-owner-convergence.md`

## 当前结论

- 拓扑页前几刀已经把 challenge header、template hero、template workbench、challenge context rail、canvas workspace、network quick editor、entry node section 等大块模板从父页拆开。
- 当前页面本地剩余最成组的脚本 owner 是 `updateNetworkDraft`、`updateNodeDraft`、`updateSelectedNodeField`、`updateEntryNodeKey`、`updateLinkDraft`、`removeLinkDraft`、`updatePolicyDraft`、`removePolicyDraft` 这组 `draft` 变更 helper。
- 这组逻辑和 `useTopologyStructureMutations.ts` 的“拓扑结构变更 owner”是同一能力域，继续留在页面会让 owner 再次分裂。

## 任务切片

### Slice 1：扩展 topology structure mutations owner

- 目标：
  - 把页面本地 `draft` 变更 helper 并进 `useTopologyStructureMutations.ts`。
  - `useChallengeTopologyStudioPage.ts` 统一对页面暴露新增的更新/删除方法。
- 预期改动：
  - `code/frontend/src/features/challenge-topology-studio/model/useTopologyStructureMutations.ts`
  - `code/frontend/src/features/challenge-topology-studio/model/useChallengeTopologyStudioPage.ts`
  - `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
- 边界：
  - `useTopologyStructureMutations.ts` 只负责 `draft` 本地结构变更，不发请求、不做 toast 以外的新副作用。
  - `useChallengeTopologyStudioPage.ts` 继续是页面级 feature owner，对外暴露同名方法。
  - 页面只做事件转发与模板装配，不再直接 `Object.assign(draft.value...)`。
- 验证：
  - `git diff --check -- code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue code/frontend/src/features/challenge-topology-studio/model/useChallengeTopologyStudioPage.ts code/frontend/src/features/challenge-topology-studio/model/useTopologyStructureMutations.ts`
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - draft 变更 owner 是否已经完全回到 feature model
  - 页面是否还残留直接操作 `draft.value` 的局部 helper

### Slice 2：补源码护栏与 review 进度

- 目标：
  - 为拓扑页补 source guard，阻止 `draft` 变更 helper 回流到页面。
  - 更新 frontend review 索引，记录拓扑页 `TD-1` 的继续收口进度。
- 预期改动：
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
  - `cd code/frontend && npm run typecheck`
  - `bash scripts/check-consistency.sh`
- Review focus：
  - source guard 是否明确卡住页面本地 `draft` 变更 helper
  - review 文档是否把这刀进度记录为拓扑页继续向局部尾项收敛

## 风险

- `updateSelectedNodeField` 牵涉当前选中节点草稿，若签名改坏会影响画布右侧快速编辑。
- `removeLinkDraft` / `removePolicyDraft` 需要保持只按 `uid` 删除当前草稿项，不能误删其他结构。
- 如果 owner 扩展方式不对，可能把 `useTopologyStructureMutations.ts` 变成过宽工具；因此本轮只收与 `draft` 结构变更直接同域的函数。

## 回退方式

- 如扩展 `useTopologyStructureMutations.ts` 后出现回归，可回退该 composable 的新增更新函数，并恢复页面本地 helper。
- 本轮只影响拓扑前端 feature/model、页面装配、源码护栏和 review 文档，不涉及 API、路由或服务端逻辑。
