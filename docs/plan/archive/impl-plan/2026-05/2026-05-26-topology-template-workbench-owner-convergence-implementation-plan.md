> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ChallengeTopologyStudioPage.vue` 当前结构、拓扑页既有 `Topology*Section/Panel` 子组件分层模式
> 替代：无

# Topology Template Workbench Owner Convergence Implementation Plan

## 目标

- 让 `ChallengeTopologyStudioPage.vue` 把 template-library 模式里的 `topology-workbench` 装配壳抽成独立子组件。
- 保持父组件继续拥有 `activeWorkbenchTab`、draft 编辑数据、模板搜索 / 应用 / 删除动作以及页面级 owner。
- 让新子组件只承接标签切换区、四个 tab 的展示组合和模板侧栏装配。

## 非目标

- 本轮不改 `TopologyCanvasWorkspaceSection.vue`、`TopologyNodeSection.vue`、`TopologyNetworkSection.vue`、`TopologyConnectivitySections.vue`、`TopologyTemplateSidePanel.vue` 的内部逻辑。
- 本轮不改 challenge 模式主工作区和 route view。
- 本轮不改 `useChallengeTopologyStudioPage` 的数据结构和行为 owner。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyCanvasWorkspaceSection.vue`
- `code/frontend/src/components/platform/topology/TopologyNodeSection.vue`
- `code/frontend/src/components/platform/topology/TopologyNetworkSection.vue`
- `code/frontend/src/components/platform/topology/TopologyConnectivitySections.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateSidePanel.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `.harness/reuse-decisions/topology-template-workbench-owner-convergence.md`

## 当前结论

- 挑战模式右侧 `context rail` 已抽走后，父页 template-library 模式里最明显的大块模板就是 `topology-workbench`。
- 这块已经由多个稳定子组件构成，继续留在父页只会维持大段装配模板，不增加 owner 清晰度。
- 最小安全切片是：新组件接收当前 tab、现成数据与透传事件；父页继续持有页面级状态和动作 owner。

## 任务切片

### Slice 1：抽出 template workbench 子组件

- 目标：
  - 新建 `TopologyTemplateWorkbench.vue`，承接 template-library 模式下的 tab strip、四个 tab 展示块和模板侧栏。
  - `ChallengeTopologyStudioPage.vue` 只保留页面级 owner 与模式切换壳。
- 预期改动：
  - `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
  - `code/frontend/src/components/platform/topology/TopologyTemplateWorkbench.vue`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
- 组件边界：
  - 父组件继续拥有 `activeWorkbenchTab`、draft 数据、模板列表、模板动作、selection owner
  - 子组件只接收现成值，透传 `update:activeWorkbenchTab` 和现有业务事件
- 验证：
  - `git diff --check -- code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue code/frontend/src/components/platform/topology/TopologyTemplateWorkbench.vue code/frontend/src/__tests__/architectureAllowlist.ts`
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - 父组件是否继续持有 selection / draft / template action owner
  - 新组件是否只是 mode-specific workbench shell，不引入新的本地业务状态

### Slice 2：回写 TD-1 进展

- 目标：
  - 把 template-library workbench 切片进展写回主索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "TopologyTemplateWorkbench|template workbench|topology-workbench" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否清楚说明这是 template-library 模式装配壳收口，而不是下沉页面 owner

## 风险

- `TopologyTemplateWorkbench.vue` 需要透传较多 props 和 emits，遗漏会导致 tab 交互或模板动作回归。
- 新组件会直接 type import `challenge-topology-studio/model`，需要同步 architecture allowlist。
- 如果顺手混入 hero 区或 challenge 模式主区，会超出本轮边界。

## 回退方式

- 如 `TopologyTemplateWorkbench.vue` 抽层引入回归，可回退新组件并恢复父页 template-library workbench 模板。
- 本轮只影响前端组件层、测试护栏和文档，不涉及 API、route 或服务端逻辑。
