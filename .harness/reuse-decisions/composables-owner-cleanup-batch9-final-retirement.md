# Reuse Decision

## Change type
refactor_existing / docs / test / tooling

## Existing code searched
- `code/frontend/src/composables/__tests__/useProbeEasterEggs.test.ts`
- `code/frontend/src/shared/model/common/useProbeEasterEggs.ts`
- `code/frontend/src/shared/model/common/__tests__/usePagination.test.ts`
- `code/frontend/scripts/check-theme-tail.mjs`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/02-routing.md`
- `docs/architecture/frontend/03-state-management.md`
- `docs/architecture/frontend/09-spacing-system.md`

## Similar implementations found
- `shared/model/common/__tests__/usePagination.test.ts` 已采用“测试跟随 shared model owner 邻近放置”的形态
- `composables-owner-cleanup-batch7-probe-easter-eggs` 已把 runtime owner 收口到 `shared/model/common/useProbeEasterEggs.ts`
- `composables-owner-cleanup-batch8-websocket-runtime` 已把最后一个 runtime composable 收口到 `shared/model/realtime/useWebSocket.ts`

## Decision
refactor_existing

## Reason
- 当前 `src/composables` 已不再保留运行时代码，只剩一份历史测试和少量过时文档描述
- `useProbeEasterEggs` 测试继续留在旧目录，会让 `src/composables` 假装仍是活动层，和当前 owner 事实冲突
- 主题尾部检查脚本和前端架构文档都应对齐“活动层已经迁到 `pages / features / shared / widgets`”的最新状态

## Files to modify
- `.harness/reuse-decisions/composables-owner-cleanup-batch9-final-retirement.md`
- `docs/plan/impl-plan/2026-05-31-composables-owner-cleanup-batch9-final-retirement-plan.md`
- `code/frontend/src/composables/__tests__/useProbeEasterEggs.test.ts`
- `code/frontend/src/shared/model/common/__tests__/useProbeEasterEggs.test.ts`
- `code/frontend/scripts/check-theme-tail.mjs`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/02-routing.md`
- `docs/architecture/frontend/03-state-management.md`
- `docs/architecture/frontend/09-spacing-system.md`

## After implementation
- `src/composables` 不再承担任何活动测试或运行时 owner
- `useProbeEasterEggs` 的测试与实现统一落在 `shared/model/common`
- 主题检查脚本与当前架构事实源同步到最新目录边界
