# 2026-05-31 composables owner cleanup batch4 clipboard countdown plan

> 状态：Draft
> 关联 reuse decision：`.harness/reuse-decisions/composables-owner-cleanup-batch4-clipboard-countdown.md`

## 目标

把 `useClipboard` 和 `useCountdown` 从历史 `code/frontend/src/composables/` 收口到更准确的共享层。

## 非目标

- 不处理 `useTheme`
- 不处理 `routeNavigationTransport`
- 不处理 `routeQueryTransport`
- 不处理 `useRouteQueryTabs`
- 不处理 `useUrlSyncedTabs`
- 不处理 `useWebSocket`
- 不处理 `useProbeEasterEggs`

## 输入事实源

- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `code/frontend/scripts/frontend-architecture-policy.json`
- `code/frontend/src/composables/useClipboard.ts`
- `code/frontend/src/composables/useCountdown.ts`

## 目标归属

- `useClipboard` -> `shared/model/common/useClipboard.ts`
- `useCountdown` -> `shared/lib/time/useCountdown.ts`

理由：

- `useClipboard` 直接承接统一 toast 反馈，是共享反馈 owner 的延伸
- `useCountdown` 只承接时间派生和定时器 cleanup，不依赖业务 contract，也不承接共享 workflow owner

## 任务切片

### Slice 1

迁移基础文件与直接消费方：

- 移动 `useClipboard.ts`
- 移动 `useCountdown.ts`
- 修正 `useChallengeInstance.ts`
- 修正 `useInstanceOperations.ts`
- 修正 `ChallengeInstanceCard.vue`

验证：

- `cd code/frontend && timeout 180s npm run typecheck`

### Slice 2

修正文档事实与测试：

- 更新 `01-architecture-overview.md`
- 更新 `03-state-management.md`
- 修正 `useChallengeInstance.test.ts` 里的 mock 路径

验证：

- `cd code/frontend && timeout 180s npm run test:run -- src/features/challenge-detail/model/useChallengeInstance.test.ts`
- `python3 scripts/check-docs-consistency.py`
- `git diff --check`

## 风险点

- `useClipboard` 当前没有独立测试，迁移时要注意不要顺手改行为
- `useCountdown` 当前被 UI 组件直接消费，迁移只应改变 import，不应改 tick 逻辑
- 文档里现阶段还没有明确记这两个落点，需要和实现一起更新

## Review focus

- `useClipboard` 进 `shared/model/common` 是否比 `shared/lib` 更贴近反馈 owner 语义
- `useCountdown` 进 `shared/lib/time` 是否保持了纯机制层边界
- 是否留下旧 `src/composables` 残余引用
