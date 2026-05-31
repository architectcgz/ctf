# Reuse Decision

## Change type
refactor_existing / shared-foundation / docs / test

## Existing code searched
- `code/frontend/src/composables/useClipboard.ts`
- `code/frontend/src/composables/useCountdown.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeInstance.ts`
- `code/frontend/src/features/challenge-detail/ui/ChallengeInstanceCard.vue`
- `code/frontend/src/features/instance-list/model/useInstanceOperations.ts`
- `code/frontend/src/shared/model/common/*`
- `code/frontend/src/shared/lib/**`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`

## Similar implementations found
- `shared/model/common/useToast.ts`、`shared/model/common/useDestructiveConfirm.ts` 已承接共享反馈型 owner
- `shared/lib/request/useAbortController.ts`、`shared/lib/keyboard/useTabKeyboardNavigation.ts` 已承接纯机制型基础能力
- `useClipboard` 依赖 `useToast` 并直接提供统一复制成功 / 失败反馈，语义更接近共享反馈 owner
- `useCountdown` 只处理时间派生、定时器和生命周期清理，语义更接近纯基础能力

## Decision
refactor_existing

## Reason
- `useClipboard` 与 `useToast` 强绑定，不是单纯浏览器 API 包装，更适合进入 `shared/model/common`
- `useCountdown` 不携带业务状态 owner，只是共享时间派生机制，更适合进入 `shared/lib/time`
- 两者都没有 router / realtime / theme 级复杂依赖，适合组成一批低风险 owner cleanup

## Files to modify
- `.harness/reuse-decisions/composables-owner-cleanup-batch4-clipboard-countdown.md`
- `docs/plan/impl-plan/2026-05-31-composables-owner-cleanup-batch4-clipboard-countdown-plan.md`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `code/frontend/src/composables/useClipboard.ts`
- `code/frontend/src/composables/useCountdown.ts`
- `code/frontend/src/shared/model/common/useClipboard.ts`
- `code/frontend/src/shared/lib/time/useCountdown.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeInstance.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeInstance.test.ts`
- `code/frontend/src/features/challenge-detail/ui/ChallengeInstanceCard.vue`
- `code/frontend/src/features/instance-list/model/useInstanceOperations.ts`

## After implementation
- 历史 `src/composables` 进一步缩小到 router / theme / realtime / easter-eggs 等仍需单独判断 owner 的能力
- 复制反馈与时间倒计时分别进入更匹配的 shared layer
