# composables owner cleanup 第一批计划

> 状态：Current
> 事实源：`code/frontend/src/composables/*`、`shared/ui/layout/*`、`shared/ui/common/*`
> 替代：无

## 目标

- 把第一批 owner 已经明确的历史 composables 从 `src/composables/` 收口到 `shared/model/*`
- 让 layout 共享壳和共享 UI 原语的 state / composition owner 与其 UI 落点对齐

## 范围

- `useWorkspaceShellNavigation`
- `useBackofficeBreadcrumbDetail`
- `useToast`
- `useDestructiveConfirm`

## 非目标

- 本轮不处理 `routeNavigationTransport`、`routeQueryTransport`、`useRouteQueryTabs`
- 本轮不处理 `useWebSocket`、`useTheme`、`usePagination` 这类第二批共享能力
- 本轮不新增业务逻辑，只做 owner 与引用入口调整

## 方案

1. `useWorkspaceShellNavigation`、`useBackofficeBreadcrumbDetail` -> `shared/model/layout/*`
2. `useToast`、`useDestructiveConfirm` -> `shared/model/common/*`
3. 更新 `shared/ui/layout/*`、`shared/ui/common/*`、`features/**`、`pages/**`、`router/**` 的 import
4. 同步迁移相关测试或调整其 import
5. 更新前端架构文档里对 `composables/` 与 `shared/model/*` 的事实描述

## 验证

- `bash scripts/check-task-intake.sh --reuse-decision composables-owner-cleanup-batch1`
- `cd code/frontend && timeout 180s npm run test:run -- src/shared/ui/common/__tests__/AppToast.test.ts src/shared/ui/layout/__tests__/TopNav.test.ts src/shared/ui/layout/__tests__/Sidebar.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && timeout 180s npm run typecheck`
- `git diff --check`
