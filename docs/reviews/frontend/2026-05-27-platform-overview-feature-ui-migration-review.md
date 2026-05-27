# Platform Overview Feature UI Migration 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-platform-overview-feature-ui-migration-implementation-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/platform-overview-feature-ui-migration.md`
    - `docs/plan/impl-plan/2026-05-27-platform-overview-feature-ui-migration-implementation-plan.md`
    - `code/frontend/src/features/platform-overview/index.ts`
    - `code/frontend/src/features/platform-overview/ui/*`
    - `code/frontend/src/views/platform/PlatformOverview.vue`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts`
    - `code/frontend/src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts`
    - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
    - `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
    - `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
    - `code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 allowlist 驱动的前端结构债收口，优先处理低风险 `feature-owned UI` 候选。
- Gate verdict：Pass（本次为同上下文复核；当前回合未使用独立 subagent review）

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- `PlatformOverviewPage.vue` 已从 `components/platform/dashboard/` 收口到 `features/platform-overview/ui/`，route view 不再依赖 legacy component page。
- `views/platform/PlatformOverview.vue` 现在直接从 `features/platform-overview` public API 组合 `usePlatformOverviewPage()` 与 `PlatformOverviewPage`，继续保持 route page 薄壳。
- `PlatformOverviewHeroPanel`、`PlatformOverviewAlertsSection`、`PlatformOverviewHotspotsSection` 仍保留在 `components/platform/dashboard/` 作为稳定子分区，本轮没有把模板拆分和 owner 迁移混在一起。
- `architectureAllowlist.ts` 已移除平台总览对应的一条 `componentFeatureImportAllowlist` 和一条 `legacyComponentPageAllowlist`，相关 raw-source 测试与 `components.d.ts` 也已切到新路径。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/PlatformOverview.test.ts src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/adminRootHeroLayout.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-overview-feature-ui-migration.md docs/plan/impl-plan/2026-05-27-platform-overview-feature-ui-migration-implementation-plan.md docs/reviews/frontend/2026-05-27-platform-overview-feature-ui-migration-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/features/platform-overview code/frontend/src/views/platform/PlatformOverview.vue code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/components.d.ts code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts code/frontend/src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts code/frontend/src/views/__tests__/workspaceShellStyles.test.ts code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 本轮只迁平台总览 page shell，不处理其它 dashboard / contest 候选。
- 子分区仍在 `components/platform/dashboard/`，后续若要继续收口，需要单独判断是否值得一起迁。

## Touched known-debt status

- 本轮 touched 的已知结构债是“应属于单一 feature 的 page-sized UI 仍滞留在 `components/**`，并依赖 allowlist 才能存活”。
- 该债务在平台总览这组 touched surface 上已完成收口：page shell 已迁到 `features/platform-overview/ui`，对应 component->feature 例外和 legacy page 例外已移除，route view 与 raw-source guardrail 已同步到新边界。
