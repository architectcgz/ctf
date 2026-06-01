# WebSocket Composable Boundary Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-websocket-composable-boundary-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/websocket-composable-boundary-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-websocket-composable-boundary-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-websocket-composable-boundary-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/composables/useWebSocket.ts`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
- Classification check：同意按 shared composable boundary cleanup 处理；这不是 feature owner 迁移，而是清理历史残留的无效 store 依赖。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useWebSocket.ts` 保留 `api/auth` ticket 与 `runtime/globalErrorRuntime` 的 shared owner 是合理的。
- `useAuthStore` 如果确实未被使用，就不应该继续让 shared composable 背上 `store` 边界。
- 这轮完成后，`composableMultiBoundaryAllowlist` 应该清空，剩余 allowlist 重点会回到 `featureRouterImportAllowlist`。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/websocket-composable-boundary-cleanup.md docs/plan/impl-plan/2026-05-29-websocket-composable-boundary-cleanup-plan.md docs/reviews/frontend/2026-05-29-websocket-composable-boundary-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/composables/useWebSocket.ts code/frontend/src/__tests__/architectureAllowlist.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `featureRouterImportAllowlist` 仍然保留，属于下一轮单独判定范围。
