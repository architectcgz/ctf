> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ChallengeTopologyStudioPage.vue` 当前结构、拓扑页既有 header/hero/workbench 子组件分层模式
> 替代：无

# Topology Template Library Header Owner Convergence Implementation Plan

## 目标

- 让 `ChallengeTopologyStudioPage.vue` 把 template-library 模式顶部 `PageHeader` 与其动作按钮抽成独立子组件。
- 保持父组件继续拥有刷新模板目录、重置模板编辑器等页面级动作 owner。
- 让新子组件只承接 template-library header 展示壳与事件透传。

## 非目标

- 本轮不改 template hero、template workbench 和 challenge 工作区。
- 本轮不继续抽样式区块之外的其他 template-library 结构。
- 本轮不改 `useChallengeTopologyStudioPage` 的业务逻辑。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyChallengeWorkspaceHeader.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `.harness/reuse-decisions/topology-template-library-header-owner-convergence.md`

## 当前结论

- challenge header、template hero 和 template workbench 都已抽走后，template-library 模式顶部 `PageHeader` 成了父页里最明显的一块稳定模板壳。
- 这块不持有拓扑草稿、selection 或模板列表状态，只透传重置与刷新动作，因此适合继续下沉。
- 最小安全切片是：新组件接收现成文案并负责两个按钮模板，父页继续保留动作 owner。

## 任务切片

### Slice 1：抽出 template-library header 子组件

- 目标：
  - 新建 `TopologyTemplateLibraryHeader.vue`，承接 template-library 模式顶部 `PageHeader` 和动作按钮。
  - `ChallengeTopologyStudioPage.vue` 不再直接内联这段 header 模板。
- 预期改动：
  - `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
  - `code/frontend/src/components/platform/topology/TopologyTemplateLibraryHeader.vue`
- 组件边界：
  - 父组件继续拥有 `reset / refresh` 动作与 `pageHeader` 文案 owner
  - 子组件只接收现成值并透传事件
- 验证：
  - `git diff --check -- code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue code/frontend/src/components/platform/topology/TopologyTemplateLibraryHeader.vue`
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts -t "template library header|template header"`
- Review focus：
  - 模板 header 动作 owner 是否仍留在父页
  - 子组件是否只承接展示壳，不吸入页面级逻辑

### Slice 2：回写 TD-1 进展

- 目标：
  - 把 template-library header 切片进展写回主索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "TopologyTemplateLibraryHeader|template-library header|PageHeader" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否清楚说明这是 template-library header 壳收口

## 风险

- source 护栏要把 `PageHeader` 和 header 按钮断言迁到新组件。
- 如果继续把 template-library 样式区块大范围迁走，会超出本轮边界。

## 回退方式

- 如 `TopologyTemplateLibraryHeader.vue` 抽层引入回归，可回退新组件并恢复父页 template-library header 模板。
- 本轮只影响前端组件层、测试护栏和文档，不涉及 API、route 或服务端逻辑。
