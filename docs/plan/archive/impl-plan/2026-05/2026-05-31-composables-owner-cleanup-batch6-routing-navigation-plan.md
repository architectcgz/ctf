# 2026-05-31 composables owner cleanup batch6 routing navigation plan

> 状态：Draft
> 关联 reuse decision：`.harness/reuse-decisions/composables-owner-cleanup-batch6-routing-navigation.md`

## 目标

把 router/query 这一组从历史 `code/frontend/src/composables/` 收口出来，并统一收口到 `shared/model/navigation`；`shared/lib/navigation` 只保留 `routeTarget.ts` 这类纯契约入口。

## 非目标

- 不改 `useRouteQueryTabs` / `useUrlSyncedTabs` 的行为
- 不把 `useUrlSyncedTabs` 强行改成 Vue Router 实现
- 不处理 `useWebSocket`
- 不处理 `useProbeEasterEggs`

## 输入事实源

- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
- `code/frontend/src/shared/lib/navigation/routeTarget.ts`

## 目标归属

- `routeNavigationTransport` -> `shared/model/navigation/useRouteNavigationTransport.ts`
- `routeQueryTransport` -> `shared/model/navigation/useRouteQueryTransport.ts`
- `useRouteQueryTabs` -> `shared/model/navigation/useRouteQueryTabs.ts`
- `useUrlSyncedTabs` -> `shared/model/navigation/useUrlSyncedTabs.ts`

理由：

- 四个文件都直接承接 router runtime 或 tab/query 页面状态 owner，属于 reviewed route-aware navigation model
- `shared/lib/navigation` 只保留纯契约，不直接承接 `vue-router` runtime

## 任务切片

### Slice 1

迁移 transport：

- `routeNavigationTransport.ts`
- `routeQueryTransport.ts`
- 修正直接消费方与相关测试断言

验证：

- `cd code/frontend && timeout 180s npm run test:run -- src/__tests__/architectureBoundaries.test.ts`

### Slice 2

迁移 tab/query owner：

- `useRouteQueryTabs.ts`
- `useUrlSyncedTabs.ts`
- 修正 features / pages / 测试断言

验证：

- `cd code/frontend && timeout 180s npm run typecheck`
- `cd code/frontend && timeout 180s npm run test:run -- src/pages/contests/__tests__/ContestDetail.test.ts src/pages/scoreboard/__tests__/ScoreboardView.test.ts src/features/platform/contests/ui/contestAdminUiStrategy.test.ts`

### Slice 3

对齐架构文档事实：

- 更新 `01-architecture-overview.md`
- 更新 `03-state-management.md`
- 更新 `07-pages-dataflow.md`

验证：

- `python3 scripts/check-docs-consistency.py`
- `git diff --check`

## 风险点

- 大量测试直接用 `?raw` 读取源码并断言 import 路径，需要整体同步
- `useUrlSyncedTabs` 不是 router 版实现，迁移时不能误改为 `useRoute/useRouter`
- `useRouteQueryTabs` 的 `routeName/routeParams` 分支多，owner 迁移时不应顺手重构

## Review focus

- route-aware navigation owner 是否都已收口到 `shared/model/navigation`
- `shared/lib/navigation` 是否只剩纯契约入口
- 是否只发生 owner 迁移，没有混入行为变化
- 是否清干净了旧 `src/composables` 路径引用与源码断言
