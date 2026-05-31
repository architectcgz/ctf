# Reuse Decision

## Change type
refactor_existing / shared-navigation / docs / test

## Existing code searched
- `code/frontend/src/composables/routeNavigationTransport.ts`
- `code/frontend/src/composables/routeQueryTransport.ts`
- `code/frontend/src/composables/useRouteQueryTabs.ts`
- `code/frontend/src/composables/useUrlSyncedTabs.ts`
- `code/frontend/src/shared/lib/navigation/routeTarget.ts`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
- `code/frontend/src/features/**/model/*RoutePage*.ts`
- `code/frontend/src/features/**/ui/*.vue`
- `code/frontend/src/pages/**/__tests__/*.test.ts`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `docs/architecture/frontend/07-pages-dataflow.md`

## Similar implementations found
- `shared/lib/navigation/routeTarget.ts` 已提供共享导航契约入口，适合作为 route-aware target type 的稳定落点
- `architectureBoundaries.test.ts` 已明确 feature 不能直接依赖 router，必须通过 reviewed route-aware composables
- `useRouteQueryTabs` / `useUrlSyncedTabs` 不只是 transport，而是带 active tab、queryKey、默认 tab 与键盘交互编排的共享页面状态 owner

## Decision
refactor_existing

## Reason
- `routeNavigationTransport`、`routeQueryTransport`、`useRouteQueryTabs`、`useUrlSyncedTabs` 都直接承接 router runtime 或 tab/query 页面状态 owner，适合统一进 `shared/model/navigation`
- `shared/lib/navigation` 只保留 `routeTarget.ts` 这类纯契约入口，不直接承接 `vue-router` runtime owner
- 四个文件需要同批迁移，并按 navigation owner 收口，避免 transport 与 tab/query 状态再次散落在历史 `composables`

## Files to modify
- `.harness/reuse-decisions/composables-owner-cleanup-batch6-routing-navigation.md`
- `docs/plan/impl-plan/2026-05-31-composables-owner-cleanup-batch6-routing-navigation-plan.md`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `code/frontend/src/composables/routeNavigationTransport.ts`
- `code/frontend/src/composables/routeQueryTransport.ts`
- `code/frontend/src/composables/useRouteQueryTabs.ts`
- `code/frontend/src/composables/useUrlSyncedTabs.ts`
- `code/frontend/src/shared/model/navigation/useRouteNavigationTransport.ts`
- `code/frontend/src/shared/model/navigation/useRouteQueryTransport.ts`
- `code/frontend/src/shared/model/navigation/useRouteQueryTabs.ts`
- `code/frontend/src/shared/model/navigation/useUrlSyncedTabs.ts`
- `code/frontend/src/features/**`
- `code/frontend/src/pages/**/__tests__/*.test.ts`
- `code/frontend/src/features/**/__tests__/*.test.ts`

## After implementation
- route-aware transport 与 tab/query 同步 owner 从历史 `src/composables` 收口到 `shared/model/navigation`
- `shared/lib/navigation` 继续只保留 route target 契约
- 历史 `src/composables` 继续只剩 realtime 与局部杂项能力
