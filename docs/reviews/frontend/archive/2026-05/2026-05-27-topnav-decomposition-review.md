# TopNav Decomposition 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-topnav-decomposition-implementation-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/topnav-decomposition.md`
    - `docs/plan/impl-plan/2026-05-27-topnav-decomposition-implementation-plan.md`
    - `docs/reviews/frontend/2026-05-27-topnav-decomposition-review.md`
    - `code/frontend/src/components/layout/TopNav.vue`
    - `code/frontend/src/components/layout/topnav/**/*`
    - `code/frontend/src/components/layout/__tests__/TopNav.test.ts`
    - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于布局层超大组件拆分，保留 `TopNav.vue` 为 route/theme/notification owner，只切稳定视图壳。
- Gate verdict：Pass after targeted verification

## Findings

- None.

## Material findings

- None.

## Senior implementation assessment

- `TopNav.vue` 现在只保留 route-derived breadcrumb detail、`useWorkspaceShellNavigation()` 装配、brand picker open state、theme/brand action、`NotificationDrawer` trigger slot 装配和 logout owner。
- `TopNavMobileToggle`、`TopNavBreadcrumbs`、`TopNavBrandPicker`、`TopNavNotificationTrigger`、`TopNavUserCard` 都只承接稳定视图壳，没有重新直接依赖 router、store 或 composable owner。
- 为了让既有 `topnav-*` 样式继续命中拆出的子组件 DOM，`TopNav.vue` 的样式已改为非 scoped，并保持 `topnav-` 前缀避免布局层样式泄漏到无关组件。
- `TopNav.test.ts` 与 `sharedThemeTokenAdoption.test.ts` 已切到聚合源码断言，结构护栏仍覆盖移动 toggle、breadcrumb、brand picker、通知 trigger 与 theme token surface。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/components/layout/__tests__/TopNav.test.ts src/components/layout/__tests__/Sidebar.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## Residual risk

- 这轮只收口了 topnav 的稳定视图壳。更深层的布局行为 owner，例如通知中心交互策略、侧栏导航判定与 topnav brand/theme 之外的全局壳层责任，仍需要按后续切片继续评估。

## Touched known-debt status

- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 里 `P2` 的 `TopNav.vue` 大组件壳体债已在 touched surface 上完成第一轮收口，不再以单文件 700+ 行模板/样式混写的形式继续增长。
