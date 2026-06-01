> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ChallengeTopologyStudioPage.vue` 当前结构、拓扑页既有 header/hero 子组件分层模式
> 替代：无

# Topology Challenge Header Owner Convergence Implementation Plan

## 目标

- 让 `ChallengeTopologyStudioPage.vue` 把 challenge 模式顶部 `workspace-topbar` 和 heading/summary 区抽成独立子组件。
- 保持父组件继续拥有返回、刷新、导出、保存动作和相关禁用态 owner。
- 让新子组件只承接 challenge header 展示壳与事件透传。

## 非目标

- 本轮不改 challenge 主工作区和右侧 `context rail`。
- 本轮不改 template-library `PageHeader`。
- 本轮不改 `TopologySummaryGrid.vue` 内部逻辑。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologySummaryGrid.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `.harness/reuse-decisions/topology-challenge-header-owner-convergence.md`

## 当前结论

- template hero 和 template workbench 都已抽走后，challenge 模式最显眼的父页模板块就是顶部 header 组合。
- 这块不持有 draft 或 selection，只透传页面动作，因此适合继续下沉。
- 最小安全切片是：新组件接收现成文案、summary、保存/导出状态和按钮事件；父页继续持有动作 owner。

## 任务切片

### Slice 1：抽出 challenge header 子组件

- 目标：
  - 新建 `TopologyChallengeWorkspaceHeader.vue`，承接 `workspace-topbar` 与其后的 heading/summary 区。
  - `ChallengeTopologyStudioPage.vue` 只保留 challenge 模式壳与主工作区 owner。
- 预期改动：
  - `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
  - `code/frontend/src/components/platform/topology/TopologyChallengeWorkspaceHeader.vue`
- 组件边界：
  - 父组件继续拥有 `back / refresh / export / save` 及其禁用态
  - 子组件只接收现成值并透传事件
- 验证：
  - `git diff --check -- code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue code/frontend/src/components/platform/topology/TopologyChallengeWorkspaceHeader.vue`
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts -t "challenge header|workspace topbar"`
- Review focus：
  - 动作 owner 是否仍留在父页
  - 新组件是否只是展示壳，不持有业务状态

### Slice 2：回写 TD-1 进展

- 目标：
  - 把 challenge header 切片进展写回主索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "TopologyChallengeWorkspaceHeader|challenge header|workspace-topbar" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否清楚说明这是 challenge 模式 header 壳收口

## 风险

- source 护栏要把 topbar/heading 断言迁到新组件。
- 如果顺手把 challenge 主工作区也一起挪走，会超出本轮边界。

## 回退方式

- 如 `TopologyChallengeWorkspaceHeader.vue` 抽层引入回归，可回退新组件并恢复父页 challenge header 模板。
- 本轮只影响前端组件层、测试护栏和文档，不涉及 API、route 或服务端逻辑。
