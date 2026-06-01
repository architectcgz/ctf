# Sidebar Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-sidebar-owner-cleanup-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/sidebar-owner-cleanup.md`
    - `docs/plan/impl-plan/2026-05-28-sidebar-owner-cleanup-plan.md`
    - `docs/reviews/frontend/2026-05-28-sidebar-owner-cleanup-review.md`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
    - `code/frontend/src/components/layout/Sidebar.vue`
    - `code/frontend/src/components/layout/sidebar/useSidebarNavigationViewState.ts`
    - `code/frontend/src/components/layout/sidebar/sidebarShell.css`
    - `code/frontend/src/components/layout/__tests__/Sidebar.test.ts`
    - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Classification check：同意按 layout 共享基础设施上的非 trivial frontend refactor 处理，目标是继续收口 `Sidebar.vue` 的本地 nav owner 与样式混写，而不是改 workspace shell bridge contract。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `Sidebar.vue` 现在只保留 `collapsed / mobileOpen` props、desktop/mobile panel 组合和 emits contract，不再继续混放 nav 判定、展开态和壳体样式。
- `useSidebarNavigationViewState.ts` 已明确承接 `brandKicker`、`backofficeItems`、`expandedMenus`、item active / expanded / highlight 判定，以及 `navigate()` 的 same-path / same-query 短路和 mobile close 语义；这批本地 nav view-model 不再滞留在父 SFC。
- `sidebarShell.css` 已承接 sidebar shell、desktop/mobile 侧栏和 light/dark token 样式，raw-source 与 shared-theme 护栏也已切到聚合源码视角，不再假设样式必须留在父 SFC。
- `Sidebar.vue` 文件体量从约 `617` 行降到 `62` 行；本轮 touched surface 上“workspace shell bridge + nav 判定 + navigate + shell CSS”混写债已完成收口。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/components/layout/__tests__/Sidebar.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；按 `development-pipeline` 的独立 gate 要求，依然缺少单独 reviewer 上下文的复核证据。当前流程里没有额外派生 reviewer，所以这条缺口需要在交付说明里明确。
- `TopNav.vue` 仍然是 layout P2 最主要的剩余大组件；本轮只处理侧栏本地 owner，不扩大到 topnav 的 route / breadcrumb / user-card contract。

## Touched known-debt status

- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 中这条 layout P2 对 `Sidebar.vue` 的“更深层 nav owner / shell CSS 清理”已在 touched surface 内完成收口；当前 residual 重点已经进一步收敛到 `TopNav.vue`，而不再是侧栏父组件继续混写本地 nav 判定、navigate 语义与壳体样式。
