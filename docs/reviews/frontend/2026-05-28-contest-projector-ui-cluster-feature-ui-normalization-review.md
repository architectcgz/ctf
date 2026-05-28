# Contest Projector UI Cluster Feature UI Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-contest-projector-ui-cluster-feature-ui-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/contest-projector-ui-cluster-feature-ui-normalization.md`
    - `docs/plan/impl-plan/2026-05-28-contest-projector-ui-cluster-feature-ui-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-28-contest-projector-ui-cluster-feature-ui-normalization-review.md`
    - `code/frontend/src/features/contest-projector/index.ts`
    - `code/frontend/src/features/contest-projector/ui/*`
    - `code/frontend/src/views/platform/ContestProjector.vue`
    - `code/frontend/src/views/platform/__tests__/contestProjectorPageStateExtraction.test.ts`
    - `code/frontend/src/features/contest-projector/ui/__tests__/*`
    - `code/frontend/src/components.d.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 `contest-projector` 单一 feature UI cluster owner 的继续收口。
- Gate verdict：Implemented and re-validated

## Findings

- 无新的未收口 findings。

## Material findings

- 无。

## Senior implementation assessment

- `ContestProjectorToolbar.vue`、`ContestProjectorHero.vue`、`ContestProjectorAttackMap.vue`、`ContestProjectorFocusOverlay.vue` 及其 projector cluster 子件 / 样式已整体迁入 `features/contest-projector/ui`，大屏 projector UI 不再继续滞留在旧 `components/platform/contest/projector/*` 目录。
- `ContestProjector.vue` 已切到 `@/features/contest-projector` public API，projector route view 继续只保留 page shell 与空态 / 加载态组合 owner，而不是回退到深层 feature 内部路径或旧 components 路径。
- projector UI 组件内部已统一改为依赖 `features/contest-projector/model/projectorTypes.ts` 与 `projectorFormatters.ts`，旧 `components/platform/contest/projector/contestProjectorTypes.ts`、`contestProjectorFormatters.ts` 已退出主消费面，没有继续留下双事实源。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/features/contest-projector/ui/__tests__/ContestProjectorAttackMap.test.ts src/features/contest-projector/ui/__tests__/ContestProjectorFocusOverlay.test.ts src/features/contest-projector/ui/__tests__/ContestProjectorServiceMatrix.test.ts src/views/platform/__tests__/contestProjectorPageStateExtraction.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这轮只收 projector UI cluster owner，不继续拆 `ContestProjectorAttackMap.vue` 的内部展示和交互分区；如果后续继续下钻，更合适的是按 attack map 内部 panel / overlay / board state 再单独分刀。
- `ContestProjector.vue` 当前仍保留 route view 级模板装配和 focus-panel 组合逻辑，不过 page model / UI owner 已经分清；是否继续下沉 route shell，应等 projector 线再出现新增长后单独判断。

## Touched known-debt status

- `contest-projector` 在 touched surface 内已从“只有 model owner，UI 还留在旧 components 目录”收口到 model + UI 双 owner 同 feature；旧 `components/platform/contest/projector/*` 路径已退出主消费面。
