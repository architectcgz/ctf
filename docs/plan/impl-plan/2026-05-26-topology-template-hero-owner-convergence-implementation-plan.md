> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ChallengeTopologyStudioPage.vue` 当前结构、拓扑页既有 `Topology*Section/Panel` 子组件分层模式
> 替代：无

# Topology Template Hero Owner Convergence Implementation Plan

## 目标

- 让 `ChallengeTopologyStudioPage.vue` 把 template-library 模式顶部 hero 展示区抽成独立子组件。
- 保持父组件继续拥有 `heroEyebrow`、`heroTitle`、`heroDescription`、`topologySummary`、`statusCard`、`secondaryCard` owner。
- 让新子组件只承接 hero 装配与现有展示子组件组合。

## 非目标

- 本轮不改 `TopologyStatusNotes.vue`、`TopologySummaryGrid.vue` 内部逻辑。
- 本轮不改 template-library `PageHeader` 动作区和 `TopologyTemplateWorkbench.vue`。
- 本轮不改 challenge 模式壳层。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyStatusNotes.vue`
- `code/frontend/src/components/platform/topology/TopologySummaryGrid.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `.harness/reuse-decisions/topology-template-hero-owner-convergence.md`

## 当前结论

- template-library `workbench` 已抽走后，父页 template-library 模式里最明显的大块模板就是顶部 hero 展示区。
- 这块没有业务动作，只是 summary/status 的组合壳，适合继续下沉。
- 最小安全切片是：新组件接收现成 props；父页继续持有所有展示数据 owner。

## 任务切片

### Slice 1：抽出 template hero 子组件

- 目标：
  - 新建 `TopologyTemplateHeroSection.vue`，承接 template-library 模式顶部 hero grid。
  - `ChallengeTopologyStudioPage.vue` 只保留模式级壳和数据 owner。
- 预期改动：
  - `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
  - `code/frontend/src/components/platform/topology/TopologyTemplateHeroSection.vue`
- 组件边界：
  - 父组件继续拥有 hero 文案、summary 和 status 数据
  - 子组件只接收现成值并组合 `TopologySummaryGrid`、`TopologyStatusNotes`
- 验证：
  - `git diff --check -- code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue code/frontend/src/components/platform/topology/TopologyTemplateHeroSection.vue`
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts -t "template hero"`
- Review focus：
  - 新组件是否只是展示壳，不新增本地业务状态
  - 父组件是否继续持有展示数据 owner

### Slice 2：回写 TD-1 进展

- 目标：
  - 把 template-library hero 切片进展写回主索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "TopologyTemplateHeroSection|template hero|topology-hero-grid" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否清楚说明这是 template-library 顶部展示壳收口

## 风险

- source 测试需要把 hero 区断言迁移到新组件。
- 如果顺手把 `PageHeader` 动作区一起挪走，会超出本轮边界。

## 回退方式

- 如 `TopologyTemplateHeroSection.vue` 抽层引入回归，可回退新组件并恢复父页 template hero 模板。
- 本轮只影响前端组件层、测试护栏和文档，不涉及 API、route 或服务端逻辑。
