# Topology Section Card Contract Convergence Plan

> 状态：Current
> 事实源：`SectionCard` shared owner、topology studio consumer、`:deep` guard allowlist

## Objective

- 把 topology studio 里针对 `SectionCard` 的深度样式覆盖收口成公开 contract。
- 删除 `TopologyChallengeWorkbench.vue`、`TopologyTemplateWorkbench.vue`、`TopologyTemplateSidePanel.vue` 中的 `:deep(.section-card*)`。
- 保持 challenge / template 两个拓扑工作台的现有视觉层级和交互不变，只改变样式 owner。

## Non-goals

- 不处理 topology studio 中 `input / select / textarea / data-node-editor / topology-action-btn` 的深度样式。
- 不改 topology studio 的业务交互、tab owner、模板写回流程或包导出逻辑。
- 不引入新的 topology 专属卡片组件或目录迁移。

## Source Inputs

- `code/frontend/src/shared/ui/common/SectionCard.vue`
- `code/frontend/src/features/challenge-topology-studio/ui/TopologyChallengeWorkbench.vue`
- `code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateWorkbench.vue`
- `code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateSidePanel.vue`
- `code/frontend/src/features/challenge-topology-studio/ui/TopologyCanvasWorkspaceSection.vue`
- `code/frontend/src/features/challenge-topology-studio/ui/TopologyEntryNodeSection.vue`
- `code/frontend/src/features/challenge-topology-studio/ui/TopologyNetworkSection.vue`
- `code/frontend/src/features/challenge-topology-studio/ui/TopologyNodeSection.vue`
- `code/frontend/src/features/challenge-topology-studio/ui/TopologyConnectivitySections.vue`
- `code/frontend/src/features/challenge-topology-studio/ui/TopologyPackageContextPanel.vue`
- `code/frontend/scripts/vue-deep-allowlist.json`

## Architecture Fit Check

- 当前 debt 的真实 owner 不是 topology studio 的每个子 section，而是 `SectionCard` 缺少可被跨页面声明的公开样式 contract。
- 这轮不再让 workbench 父壳通过 `.section-card__header`、`.section-card__body` 这类内部类名反向定制共享组件。
- touched surface 上的已知 debt 是 topology studio 对 `.section-card*` 的深度覆盖；本轮要把这组 debt 从 touched surface 收掉，不能只把 allowlist 改小一点。

## File Ownership Map

- `code/frontend/src/shared/ui/common/SectionCard.vue`
  - 负责共享 section 壳结构与公开 CSS variable contract。
- `code/frontend/src/features/challenge-topology-studio/ui/TopologyChallengeWorkbench.vue`
  - 负责 challenge 工作台的 page-level section rhythm，不再负责穿透 `SectionCard` 内部结构。
- `code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateWorkbench.vue`
  - 负责 template 工作台的 page-level section rhythm，不再负责穿透 `SectionCard` 内部结构。
- `code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateSidePanel.vue`
  - 负责模板侧栏的 first-card / directory surface 差异，不再通过 `:deep(.section-card:first-child)` 反向改共享组件。
- `code/frontend/src/features/challenge-topology-studio/ui/*Section.vue`
  - 负责声明自己需要的 `SectionCard` contract 或局部 class。

## Task Breakdown

### Slice 1: 扩大 SectionCard 的公开样式 contract

**Files**
- Modify: `code/frontend/src/shared/ui/common/SectionCard.vue`

- [ ] Step 1: 把当前内部 CSS variables 改成真正可被外层声明的公开 contract，避免 root 默认值把继承变量覆盖掉。
- [ ] Step 2: 为 header 对齐、标题字号、标题颜色、subtitle 颜色、header 边框等 topology 现有需求补足稳定入口。
- [ ] Step 3: 保持 teacher-flat / teacher-surface 和默认样式向后兼容。

### Slice 2: 收口 challenge workbench 的 section card owner

**Files**
- Modify: `code/frontend/src/features/challenge-topology-studio/ui/TopologyChallengeWorkbench.vue`
- Modify: `code/frontend/src/features/challenge-topology-studio/ui/TopologyCanvasWorkspaceSection.vue`
- Modify: `code/frontend/src/features/challenge-topology-studio/ui/TopologyEntryNodeSection.vue`
- Modify: `code/frontend/src/features/challenge-topology-studio/ui/TopologyNetworkSection.vue`
- Modify: `code/frontend/src/features/challenge-topology-studio/ui/TopologyNodeSection.vue`
- Modify: `code/frontend/src/features/challenge-topology-studio/ui/TopologyConnectivitySections.vue`
- Modify: `code/frontend/src/features/challenge-topology-studio/ui/TopologyPackageContextPanel.vue`

- [ ] Step 1: 把 challenge workbench 里 `.section-card*` 的深度覆盖迁成 page-level CSS variables 或 consumer class。
- [ ] Step 2: 处理 primary column / context rail 首卡去顶边的差异，不再依赖 `:deep(.section-card:first-child)`。
- [ ] Step 3: 保持 canvas、node、network、policy、package panel 的现有 section rhythm 不变。

### Slice 3: 收口 template workbench 与 library side panel 的 section card owner

**Files**
- Modify: `code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateWorkbench.vue`
- Modify: `code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateSidePanel.vue`
- Modify: `code/frontend/scripts/vue-deep-allowlist.json`

- [ ] Step 1: 把 template workbench 里的 `.section-card*` 深度覆盖迁成公开 contract。
- [ ] Step 2: 把 library side panel 首卡去顶边与 header left-rail 语义迁成显式 root class / variable override。
- [ ] Step 3: 更新 allowlist，确认这批 selector 已退场。

## Recommended Execution Order

1. 先扩 `SectionCard` 的公开 contract。
2. 再迁移 challenge workbench。
3. 最后迁移 template workbench 与 side panel，并同步 allowlist。

## Review Focus

- `SectionCard` 新公开 contract 是否真的覆盖了 topology 当前用到的 header / title / body / border / first-card 需求。
- workbench 父壳是否已经不再依赖 `.section-card__header`、`.section-card__body` 这类内部类名。
- challenge / template 两套拓扑页面是否仍保持当前 section rhythm。
- 这轮是否只收 `SectionCard` owner 链，没有把其他 `:deep` 类型混进同一批次。

## Validation

- `cd code/frontend && npm run check:vue-deep`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- code/frontend/src/shared/ui/common/SectionCard.vue code/frontend/src/features/challenge-topology-studio/ui/TopologyChallengeWorkbench.vue code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateWorkbench.vue code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateSidePanel.vue code/frontend/src/features/challenge-topology-studio/ui/TopologyCanvasWorkspaceSection.vue code/frontend/src/features/challenge-topology-studio/ui/TopologyEntryNodeSection.vue code/frontend/src/features/challenge-topology-studio/ui/TopologyNetworkSection.vue code/frontend/src/features/challenge-topology-studio/ui/TopologyNodeSection.vue code/frontend/src/features/challenge-topology-studio/ui/TopologyConnectivitySections.vue code/frontend/src/features/challenge-topology-studio/ui/TopologyPackageContextPanel.vue code/frontend/scripts/vue-deep-allowlist.json .harness/reuse-decisions/frontend-topology-section-card-contract-convergence.md docs/plan/impl-plan/2026-06-04-topology-section-card-contract-convergence-plan.md`

## Rollback / Recovery

- 如果公开 CSS variable contract 无法覆盖 topology 的 title / header 差异，再补更窄的显式 contract，不回退到新的 `.section-card*` 深度覆盖。
- 如果 challenge / template 两套视觉差异过大，优先保留共享 contract，再在 consumer root class 上做局部 override，不新建一份 topology 专属卡片组件。
