# Topology Section Card Contract Convergence Review

## Review Target

- Repository: `ctf`
- Branch: `main`
- Diff source: working tree changes for `frontend-topology-section-card-contract-convergence`
- Files reviewed:
  - `.harness/reuse-decisions/frontend-topology-section-card-contract-convergence.md`
  - `docs/plan/impl-plan/2026-06-04-topology-section-card-contract-convergence-plan.md`
  - `code/frontend/src/shared/ui/common/SectionCard.vue`
  - `code/frontend/src/features/challenge-topology-studio/ui/TopologyChallengeWorkbench.vue`
  - `code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateWorkbench.vue`
  - `code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateSidePanel.vue`
  - `code/frontend/scripts/vue-deep-allowlist.json`

## Classification Check

- 结论：同意 `non-trivial frontend refactor` 分类。
- 原因：这次继续扩共享 `SectionCard` 的公开样式 contract，并迁移 topology studio 两套工作台的样式 owner，不是局部样式替换。

## Gate Verdict

- `pass with minor issues`
- 说明：当前结论来自同上下文显式自审归档，不替代独立 reviewer gate。

## Findings

- 无 blocker / major / minor finding。

## Material Findings

- 无。

## Senior Implementation Assessment

- 当前方案比继续给 topology studio 增加父级 `:deep(.section-card*)` 更稳：`SectionCard` 现在直接暴露 title / subtitle / header / body / direct surface 所需的公开变量入口，page shell 只声明 contract，不再依赖内部类名。
- 这次没有把 topology UI 复制成另一套专属卡片组件，也没有把 `SectionCard` 再切成更多变体；对现有 shared owner 来说，公开 CSS variable contract 是更小且可复用的收口方式。
- `TopologyTemplateSidePanel.vue` 的首卡去顶边也已经改成显式 root class，不再需要 `:deep(.section-card:first-child)` 这类局部穿透。

## Required Re-validation

- `cd code/frontend && npm run check:vue-deep`
- `cd code/frontend && npm run test:run -- src/features/challenge-topology-studio/ui/ChallengeTopologyStudioPage.test.ts src/features/challenge-topology-studio/model/topologyStudioBoundary.test.ts`
- `cd code/frontend && npm run typecheck`
- `bash scripts/check-frontend-architecture.sh --quick`
- `git diff --check -- code/frontend/src/shared/ui/common/SectionCard.vue code/frontend/src/features/challenge-topology-studio/ui/TopologyChallengeWorkbench.vue code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateWorkbench.vue code/frontend/src/features/challenge-topology-studio/ui/TopologyTemplateSidePanel.vue code/frontend/scripts/vue-deep-allowlist.json .harness/reuse-decisions/frontend-topology-section-card-contract-convergence.md docs/plan/impl-plan/2026-06-04-topology-section-card-contract-convergence-plan.md docs/reviews/frontend/2026-06-04-topology-section-card-contract-convergence-review.md`

## Residual Risk

- 本轮故意不处理 topology studio 中 `input / select / textarea / data-node-editor / topology-action-btn` 相关 `:deep`，这些 selector 仍然留在 allowlist 中，需要后续再按 owner 链拆批处理。
- `SectionCard` 现在承担了更多公开样式 contract；如果后续还有第三套大面积样式语义继续进入，应该优先评估是否需要把某组 variables 收成更窄的 named variant，而不是继续无界扩张。
- 独立 reviewer gate 还没有满足；如果要把这批当成真正完成态推进提交，仍需要一次脱离当前实现上下文的复审。

## Touched Known-Debt Status

- 本次直接触达的已知 debt 是 topology studio 在 `TopologyChallengeWorkbench.vue`、`TopologyTemplateWorkbench.vue`、`TopologyTemplateSidePanel.vue` 中对 `.section-card*` 的父级穿透。
- 在本批次 touched surface 内，这部分 debt 已收口完成：对应 `.section-card`、`.section-card__header`、`.section-card__body`、`.section-card:first-child` 深度覆盖已退场，并从 allowlist 中移除。
- topology studio 仍有其他类型的 `:deep` 存量，但不再包含这条 `SectionCard` owner 链，不构成本批 diff 的 blocker。
