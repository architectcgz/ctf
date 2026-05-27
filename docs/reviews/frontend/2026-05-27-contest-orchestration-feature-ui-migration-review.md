# Contest Orchestration Feature UI Migration 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-contest-orchestration-feature-ui-migration-implementation-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/contest-orchestration-feature-ui-migration.md`
    - `docs/plan/impl-plan/2026-05-27-contest-orchestration-feature-ui-migration-implementation-plan.md`
    - `code/frontend/src/features/platform-contests/index.ts`
    - `code/frontend/src/features/platform-contests/model/useContestManagePage.ts`
    - `code/frontend/src/features/platform-contests/ui/*`
    - `code/frontend/src/views/platform/ContestManage.vue`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/views/platform/__tests__/ContestManage.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestManageExportFlowExtraction.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption*.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestOrchestrationTabsAdoption.test.ts`
    - `code/frontend/src/views/platform/__tests__/journalPlatformShellStyles.test.ts`
    - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
    - `code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts`
    - `code/frontend/src/views/__tests__/journalNoteStyles.test.ts`
    - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
    - `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
    - `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 allowlist 驱动的前端结构债收口，这次同时包含 page shell 迁移和 router owner 回收。
- Gate verdict：Pass（本次为同上下文复核；当前回合未使用独立 subagent review）

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- `ContestOrchestrationPage.vue` 已从 `components/platform/contest/` 收口到 `features/platform-contests/ui/`，route view 不再依赖 legacy component page。
- `useContestManagePage()` 现在明确接管竞赛编辑页与单场运维台跳转 owner，竞赛目录 page shell 自身不再直接依赖 `vue-router`。
- `views/platform/ContestManage.vue` 现在直接从 `features/platform-contests` public API 组合 `useContestManagePage()` 与 `ContestOrchestrationPage`，同时继续保留 `openEditDialog` 事件桥，避免影响现有管理弹窗测试面。
- `architectureAllowlist.ts` 已移除竞赛目录对应的一条 `componentFeatureImportAllowlist` 和一条 `legacyComponentPageAllowlist`，并为新的 route-aware feature model router owner 补上 `useContestManagePage.ts -> vue-router` 守卫。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ContestManage.test.ts src/views/platform/__tests__/contestManageExportFlowExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase21.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase22.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase27.test.ts src/views/platform/__tests__/contestOrchestrationTabsAdoption.test.ts src/views/platform/__tests__/journalPlatformShellStyles.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/__tests__/adminRootHeroLayout.test.ts src/views/__tests__/journalNoteStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/workspaceShellStyles.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-orchestration-feature-ui-migration.md docs/plan/impl-plan/2026-05-27-contest-orchestration-feature-ui-migration-implementation-plan.md docs/reviews/frontend/2026-05-27-contest-orchestration-feature-ui-migration-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/features/platform-contests code/frontend/src/views/platform/ContestManage.vue code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/components.d.ts code/frontend/src/views/platform/__tests__/ContestManage.test.ts code/frontend/src/views/platform/__tests__/contestManageExportFlowExtraction.test.ts code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase21.test.ts code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase22.test.ts code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase27.test.ts code/frontend/src/views/platform/__tests__/contestOrchestrationTabsAdoption.test.ts code/frontend/src/views/platform/__tests__/journalPlatformShellStyles.test.ts code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts code/frontend/src/views/__tests__/journalNoteStyles.test.ts code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 本轮只迁竞赛目录 page shell，不处理 `ContestAwdConfigWorkspaceShell.vue` 这类更肥的 contest / AWD 组件壳。
- page shell 仍然是一个相对大的 SFC，本轮主要收口 owner 和路径，不做进一步模板拆分。

## Touched known-debt status

- 本轮 touched 的已知结构债是“应属于单一 feature 的 page-sized UI 仍滞留在 `components/**`，并依赖 allowlist 才能存活”。
- 该债务在竞赛目录这组 touched surface 上已完成收口：page shell 已迁到 `features/platform-contests/ui`，对应 component->feature 例外和 legacy page 例外已移除，竞赛编辑页 / 运维页跳转 owner 也已从 page shell 回收到 `useContestManagePage()`。
