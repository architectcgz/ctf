> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ChallengeTopologyStudioPage.vue` 当前样式分布、template-library 已抽出的 header / hero / workbench 结构
> 替代：无

# Topology Template Library Style Owner Convergence Implementation Plan

## 目标

- 把 template-library 模式下的 header / hero / workbench 样式 owner 从 `ChallengeTopologyStudioPage.vue` 收回到对应子组件。
- 保持父页只保留页面壳、主题变量和极少量 section 容器样式。
- 降低父页对已抽出子组件的跨组件 `:deep` 样式覆盖。

## 非目标

- 本轮不改 challenge 模式样式。
- 本轮不改 `useChallengeTopologyStudioPage`、模板数据流或任何业务逻辑。
- 本轮不继续新增结构组件。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateLibraryHeader.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateHeroSection.vue`
- `code/frontend/src/components/platform/topology/TopologyTemplateWorkbench.vue`
- `.harness/reuse-decisions/topology-template-library-style-owner-convergence.md`

## 当前结论

- template-library 的结构壳已经基本拆完，父页里剩余最重的是一组跨子组件样式覆盖。
- 这些样式多数已经有明确组件归属，继续留在父页只会制造“结构已拆，样式还没跟着走”的 owner 混杂。
- 最小安全切片是：只迁移样式 owner，不碰模板结构和行为 contract。

## 任务切片

### Slice 1：收回 header / hero / workbench 样式 owner

- 目标：
  - `TopologyTemplateLibraryHeader.vue` 持有 header 皮肤和 `PageHeader` 局部 override。
  - `TopologyTemplateHeroSection.vue` 持有 hero 文案区和右侧 status 区的 template-library 样式。
  - `TopologyTemplateWorkbench.vue` 持有 tab strip、按钮皮肤、section card、局部输入 / 画布外观等 template-library 样式。
- 预期改动：
  - `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
  - `code/frontend/src/components/platform/topology/TopologyTemplateLibraryHeader.vue`
  - `code/frontend/src/components/platform/topology/TopologyTemplateHeroSection.vue`
  - `code/frontend/src/components/platform/topology/TopologyTemplateWorkbench.vue`
- 验证：
  - `git diff --check -- code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue code/frontend/src/components/platform/topology/TopologyTemplateLibraryHeader.vue code/frontend/src/components/platform/topology/TopologyTemplateHeroSection.vue code/frontend/src/components/platform/topology/TopologyTemplateWorkbench.vue`
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
  - `cd code/frontend && npm run typecheck`
- Review focus：
  - 样式 owner 是否跟已抽出的组件边界一致
  - 父页是否明显减少对 template-library 子组件的跨组件样式覆盖

### Slice 2：回写 TD-1 进展

- 目标：
  - 把 template-library 样式 owner 收口进展写回主索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "样式 owner|template-library|TopologyTemplateWorkbench" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否明确这是样式 owner 收口，而不是新一轮结构抽层

## 风险

- template-library 依赖父页上的主题变量，迁移时要避免把变量定义也错误下沉。
- `scoped` 样式 owner 迁移后，移动端 tab / dark mode 背景容易漏掉。

## 回退方式

- 如迁移后出现样式回归，可回退对应子组件样式 block，并恢复父页里的 template-library 覆盖规则。
- 本轮只影响前端组件样式层、测试护栏和 review 文档，不涉及 API、route 或服务端逻辑。
