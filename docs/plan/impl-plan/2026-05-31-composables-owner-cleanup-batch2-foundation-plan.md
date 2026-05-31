# 2026-05-31 composables owner cleanup batch2 foundation plan

> 状态：Draft
> 关联 reuse decision：`.harness/reuse-decisions/composables-owner-cleanup-batch2-foundation.md`

## 目标

把第二批 owner 清晰、无业务语义的基础 composables 从 `code/frontend/src/composables/` 收口到共享基础层，减少历史杂物目录。

本批只处理：

- `useAbortController`
- `usePagination`
- `useSanitize`
- `useTabKeyboardNavigation`

## 非目标

- 不处理 `useWebSocket`
- 不处理 `routeNavigationTransport`、`routeQueryTransport`
- 不处理 `useTheme`
- 不处理 `useReportStatusPolling`
- 不处理 `useRouteQueryTabs`、`useUrlSyncedTabs`

## 输入事实源

- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `code/frontend/scripts/frontend-architecture-policy.json`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`

## 目标归属

- `useAbortController` -> `shared/lib/request/useAbortController.ts`
- `usePagination` -> `shared/model/common/usePagination.ts`
- `useSanitize` -> `shared/lib/sanitize/useSanitize.ts`
- `useTabKeyboardNavigation` -> `shared/lib/keyboard/useTabKeyboardNavigation.ts`

理由：

- `useAbortController`、`useSanitize`、`useTabKeyboardNavigation` 是共享基础能力，不承载共享 UI 状态，也不绑定具体业务 owner。
- `usePagination` 依赖 `PageResult` contract，更适合放进 `shared/model/common`。

## 任务切片

### Slice 1

收口基础文件与测试：

- 移动 4 个 composable 源文件
- 移动对应测试
- 修正直接 import

验证：

- `cd code/frontend && timeout 180s npm run test:run -- src/composables/__tests__/useAbortController.test.ts src/composables/__tests__/usePagination.test.ts src/composables/__tests__/useSanitize.test.ts`
  说明：迁移后按新路径改成对应 shared/lib 测试文件

### Slice 2

修正 features / entities / pages 的消费路径：

- `useAbortController`
- `usePagination`
- `useSanitize`
- `useTabKeyboardNavigation`

验证：

- `cd code/frontend && timeout 180s npm run typecheck`
- `cd code/frontend && timeout 180s npm run test:run -- src/__tests__/architectureBoundaries.test.ts`

### Slice 3

对齐架构文档事实：

- 更新 `01-architecture-overview.md`
- 更新 `03-state-management.md`

验证：

- `python3 scripts/check-docs-consistency.py`
- `git diff --check`

## 风险点

- `useTabKeyboardNavigation` 被 route page / feature ui / page model 多处使用，迁移后要注意 `?raw` 测试引用是否漏改
- `usePagination` 带泛型重载，迁移后需要优先确认 typecheck 是否受路径变化影响
- `usePagination` 从 `shared/lib` 改判到 `shared/model/common` 后，要确认消费路径与文档事实一起收口

## Review focus

- 新路径是否匹配 shared 基础层语义
- 是否只发生路径迁移，没有混入行为变化
- 是否留下旧 `src/composables` 残余引用
