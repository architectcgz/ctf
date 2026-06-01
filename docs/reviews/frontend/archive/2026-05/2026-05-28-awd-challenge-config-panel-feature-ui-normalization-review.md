# AWD Challenge Config Panel Feature UI Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-awd-challenge-config-panel-feature-ui-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/awd-challenge-config-panel-feature-ui-normalization.md`
    - `docs/plan/impl-plan/2026-05-28-awd-challenge-config-panel-feature-ui-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-28-awd-challenge-config-panel-feature-ui-normalization-review.md`
    - `docs/architecture/features/AWD检查器校验状态设计.md`
    - `code/frontend/src/features/platform-contests/ui/AWDChallengeConfigPanel.vue`
    - `code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue`
    - `code/frontend/src/features/platform-contests/ui/index.ts`
    - `code/frontend/src/components.d.ts`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/components/platform/__tests__/AWDChallengeConfigPanel.test.ts`
    - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase20.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
    - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
    - `code/frontend/src/__tests__/architectureBoundaries.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 contest edit 剩余 feature-owned UI surface 的收口。
- Gate verdict：Implemented and re-validated

## Findings

- 无新的未收口 findings。

## Material findings

- 已修正：`AWDChallengeConfigPanel.vue` 迁入 feature UI 后如果继续显式 `import { RouterLink } from 'vue-router'`，会把 router access 带回 feature UI。最终实现移除了这个 import，保留模板里的全局 `RouterLink` 组件，架构边界测试保持通过。

## Senior implementation assessment

- `AWDChallengeConfigPanel.vue` 已迁入 `features/platform-contests/ui`，contest edit 的 AWD 配置阶段不再依赖旧 `components/platform/contest/*` 路径。
- `ContestEditWorkspacePanel.vue` 改为 feature 内部相对 import，`platform-contests` 现在完整持有这块 contest edit stage UI 的目录 owner。
- `architectureAllowlist.ts` 中 `AWDChallengeConfigPanel.vue -> @/features/awd-inspector` 这条历史例外已经删掉；`useAwdCheckResultPresentation()` 仍继续留在 `features/awd-inspector/model` 作为 presentation helper owner。
- `docs/architecture/features/AWD检查器校验状态设计.md` 已同步到新代码落点，事实源没有继续指向旧路径。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDChallengeConfigPanel.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase20.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这次只收口 `AWDChallengeConfigPanel.vue` 的目录 owner，不会顺手拆它内部 table / summary 结构；如果后续继续在同一 panel 上叠更多显示职责，需要按 header / summary / directory table 再切下一刀。

## Touched known-debt status

- `AWDChallengeConfigPanel.vue` 这条 `components -> feature` allowlist 与单一 feature UI 目录漂移，已在 touched surface 内完成收口。
