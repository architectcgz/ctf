# Contest AWD Config Panel Cluster Feature UI Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-contest-awd-config-panel-cluster-feature-ui-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/contest-awd-config-panel-cluster-feature-ui-normalization.md`
    - `docs/plan/impl-plan/2026-05-28-contest-awd-config-panel-cluster-feature-ui-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-28-contest-awd-config-panel-cluster-feature-ui-normalization-review.md`
    - `code/frontend/src/features/contest-awd-config/ui/ContestAwdConfigWorkspaceShell.vue`
    - `code/frontend/src/features/contest-awd-config/ui/ContestAwdConfigTopbar.vue`
    - `code/frontend/src/features/contest-awd-config/ui/ContestAwdConfigFooter.vue`
    - `code/frontend/src/features/contest-awd-config/ui/ContestAwdDebugStation.vue`
    - `code/frontend/src/features/contest-awd-config/ui/ContestAwdEditorHeader.vue`
    - `code/frontend/src/features/contest-awd-config/ui/ContestAwdScoreWeights.vue`
    - `code/frontend/src/features/contest-awd-config/ui/ContestAwdServiceDirectory.vue`
    - `code/frontend/src/features/contest-awd-config/ui/ContestAwdCheckerConfigSection.vue`
    - `code/frontend/src/features/contest-awd-config/ui/ContestAwdHttpStandardFields.vue`
    - `code/frontend/src/features/contest-awd-config/ui/ContestAwdLegacyProbeFields.vue`
    - `code/frontend/src/features/contest-awd-config/ui/ContestAwdScriptCheckerFields.vue`
    - `code/frontend/src/features/contest-awd-config/ui/ContestAwdTcpStandardFields.vue`
    - `code/frontend/src/features/contest-awd-config/ui/contestAwdConfigTypes.ts`
    - `code/frontend/src/features/contest-awd-config/ui/index.ts`
    - `code/frontend/src/components.d.ts`
    - `code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 `contest-awd-config` 单一 feature UI cluster 的继续收口。
- Gate verdict：Implemented and re-validated

## Findings

- 无新的未收口 findings。

## Material findings

- 无。

## Senior implementation assessment

- AWD 配置页当前依赖的 panel / fields / UI types cluster 已整体迁入 `features/contest-awd-config/ui`，不再保持 “workspace shell 在 feature，主要编辑 UI 还在 legacy component 目录” 的半迁移状态。
- `ContestAwdConfigWorkspaceShell.vue` 继续只组合 feature 内部 UI 和 model 输出，没有把 save / preview / load owner 回流到视图层。
- `ContestAwdCheckerConfigSection.vue` 与四个 checker fields、`contestAwdConfigTypes.ts` 一起迁位，避免只迁外层 panel 却把内部相对依赖留在旧目录，改动边界更干净。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这次只处理 AWD config panel / fields / UI types cluster 的 owner 迁位，不继续拆 shell 布局职责；如果后续这页继续增长，需要再单独拆 shell owner。

## Touched known-debt status

- `contest-awd-config` 这条 feature 的主要编辑 UI cluster 已从 legacy component 目录退场；本次 touched surface 内没有留下新的旧 `components/platform/contest/*` AWD config panel 引用。
