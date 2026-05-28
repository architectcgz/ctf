# Reuse Decision

## Change type
frontend architecture / request error navigation owner cleanup / runtime error owner extraction

## Existing code searched
- code/frontend/src/api/request.ts
- code/frontend/src/api/__tests__/request.test.ts
- code/frontend/src/utils/errorStatusPage.ts
- code/frontend/src/utils/__tests__/errorStatusPage.test.ts
- code/frontend/src/router/guards.ts
- code/frontend/src/router/__tests__/guards.test.ts
- code/frontend/src/main.ts
- code/frontend/src/composables/useWebSocket.ts
- docs/architecture/frontend/04-api-layer.md
- docs/architecture/frontend/08-build-deploy.md
- docs/reviews/architecture/2026-05-24-frontend-architecture-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- 现有错误状态页映射已经集中在 `utils/errorStatusPage.ts`，说明“状态页导航规则”本身是共享事实，不需要分散回每个页面重新实现。
- `main.ts`、`router/guards.ts`、`useWebSocket.ts` 已经在做真正的全局错误处理，这说明全局 owner 应该落在运行时边界，而不是 transport 层。
- `request.ts` 现有测试已经把 429/500 的跳页收回本地，但 401 仍留在请求层，属于“收了一半”的中间态。

## Decision
refactor_existing

## Reason
这次不应继续保留“请求层里只有一小段特例跳页”的中间迁移状态。最终形态应该是：
- `request.ts` 只负责 envelope 解包、`ApiError` 归一化、取消请求和 transport 元数据；
- 全局错误导航集中到单一 runtime owner；
- 页面 / feature owner 继续决定可恢复错误如何本地展示。

## Files to modify
- .harness/reuse-decisions/request-error-navigation-owner-cleanup.md
- docs/plan/impl-plan/2026-05-28-request-error-navigation-owner-cleanup-implementation-plan.md
- docs/reviews/frontend/2026-05-28-request-error-navigation-owner-cleanup-review.md
- code/frontend/src/api/request.ts
- code/frontend/src/api/__tests__/request.test.ts
- code/frontend/src/runtime/globalErrorRuntime.ts
- code/frontend/src/runtime/__tests__/globalErrorRuntime.test.ts
- code/frontend/src/main.ts
- code/frontend/src/router/guards.ts
- code/frontend/src/composables/useWebSocket.ts
- docs/architecture/frontend/04-api-layer.md
- docs/architecture/frontend/08-build-deploy.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `request.ts` 不再直接 import `auth` store 或错误页导航 helper。
- 401 会话失效、Vue runtime crash、router runtime error、WebSocket auth close 统一通过 runtime owner 处理。
- 429/5xx 在 HTTP 请求路径上继续只返回标准化 `ApiError`，由页面 / feature owner 自己决定本地 UX。
