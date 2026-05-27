# Sidebar Decomposition 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-sidebar-decomposition-implementation-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/sidebar-decomposition.md`
    - `docs/plan/impl-plan/2026-05-27-sidebar-decomposition-implementation-plan.md`
    - `docs/reviews/frontend/2026-05-27-sidebar-decomposition-review.md`
    - `code/frontend/src/components/layout/Sidebar.vue`
    - `code/frontend/src/components/layout/sidebar/**/*`
    - `code/frontend/src/components/layout/__tests__/Sidebar.test.ts`
    - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于布局层超大组件拆分，保留 `Sidebar.vue` 为 route/navigation owner，只切稳定渲染壳。
- Gate verdict：Self-check pass，独立 reviewer gate 待补

## Findings

- 未发现阻塞性问题。

## Material findings

- None.

## Senior implementation assessment

- `Sidebar.vue` 继续保留 `useWorkspaceShellNavigation()` 装配、active module 判定、expanded menu state 和 `navigate()` owner，没有把 route/navigation 逻辑拆散到子组件。
- 新增的 `components/layout/sidebar/*` 现在只承接移动端壳、桌面端壳、header、workspace label 和 nav tree 的稳定渲染，主要把移动/桌面模板重复压平。
- `Sidebar.test.ts` 与 `sharedThemeTokenAdoption.test.ts` 已切到聚合源码断言，因此布局壳、theme token 和 backoffice shell 结构护栏仍覆盖拆分后的实际代码面。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/components/layout/__tests__/Sidebar.test.ts src/components/layout/__tests__/TopNav.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## Residual risk

- `TopNav.vue` 仍是这组 `P2` 布局层大组件债的下一批重点。
- `Sidebar.vue` 里仍保留若干展示 class 判定 helper；如果后续再扩导航能力，需要继续判断这些 helper 是否要局部收口。

## Touched known-debt status

- `Sidebar.vue` 这条布局层超大组件债已在 touched surface 内完成第一轮结构拆分。
