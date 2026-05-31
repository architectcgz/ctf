# Reuse Decision

## Change type
refactor_existing / shared-common / test / docs

## Existing code searched
- `code/frontend/src/composables/useProbeEasterEggs.ts`
- `code/frontend/src/composables/__tests__/useProbeEasterEggs.test.ts`
- `code/frontend/src/features/auth/model/useLoginPage.ts`
- `code/frontend/src/features/notifications/model/useNotificationListPage.ts`
- `code/frontend/src/features/notifications/model/useNotificationDetailPage.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts`
- `code/frontend/src/shared/ui/errors/ErrorStatusShell.vue`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`

## Similar implementations found
- `shared/model/common/useClipboard.ts` 与 `shared/model/common/useToast.ts` 都属于跨 feature 复用、带浏览器交互或局部状态 owner 的共享 common model
- `useProbeEasterEggs` 依赖 `sessionStorage` 与内存回退，明显是一个共享交互状态 owner，而不是纯函数型 `shared/lib`

## Decision
refactor_existing

## Reason
- `useProbeEasterEggs` 没有 challenge / notification / auth 等业务 owner，只承接共享 probe 计数与激活状态
- 它带 `sessionStorage` 读写与降级逻辑，属于共享 common model，不适合继续留在历史 `composables/`
- 这批只迁 owner，不改阈值、key、存储格式或对外 API

## Files to modify
- `.harness/reuse-decisions/composables-owner-cleanup-batch7-probe-easter-eggs.md`
- `docs/plan/impl-plan/2026-05-31-composables-owner-cleanup-batch7-probe-easter-eggs-plan.md`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `code/frontend/src/composables/useProbeEasterEggs.ts`
- `code/frontend/src/composables/__tests__/useProbeEasterEggs.test.ts`
- `code/frontend/src/shared/model/common/useProbeEasterEggs.ts`
- `code/frontend/src/features/auth/model/useLoginPage.ts`
- `code/frontend/src/features/auth/model/useLoginPage.test.ts`
- `code/frontend/src/features/notifications/model/useNotificationListPage.ts`
- `code/frontend/src/features/notifications/model/useNotificationDetailPage.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts`
- `code/frontend/src/shared/ui/errors/ErrorStatusShell.vue`

## After implementation
- `useProbeEasterEggs` 从历史 `src/composables` 收口到 `shared/model/common`
- 消费方与测试断言统一切到共享 common owner
- `src/composables` 继续只剩 realtime owner `useWebSocket.ts`
