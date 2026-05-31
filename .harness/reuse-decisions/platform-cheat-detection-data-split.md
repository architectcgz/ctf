# Reuse Decision

## Change type
frontend refactor / platform cheat detection data split

## Existing code searched
- `code/frontend/src/features/platform/cheat-detection/model/useCheatDetectionPage.ts`
- `code/frontend/src/pages/platform/__tests__/CheatDetection.test.ts`
- `code/frontend/src/features/platform/overview/model/usePlatformOverviewData.ts`
- `code/frontend/src/features/platform/overview/model/usePlatformOverviewPage.ts`

## Similar implementations found
- `useCheatDetectionPage.ts` 当前同时持有风险检测请求、loading/error 状态和审计 route target。
- `usePlatformOverviewData.ts` 刚完成同类拆分，已经证明 platform 管理页适合把概览/检测数据 owner 单独抽成 data composable。

## Decision
refactor_existing

## Reason
当前最小正确切片是把平台作弊检测拆成：

- `useCheatDetectionData`：承接风险检测请求、loading/error 状态。
- `useCheatDetectionPage`：保留审计 route target、快捷操作和格式化函数。

这样可以：

- 去掉 `useCheatDetectionPage` 里的异步数据 owner
- 保持页面壳只承接 route 和展示编排语义

本轮不做：

- 不调整作弊检测工作台 UI
- 不改审计联动 route contract
- 不引入 shared cheat detection page owner

## Files to modify
- `.harness/reuse-decisions/platform-cheat-detection-data-split.md`
- `docs/plan/impl-plan/2026-05-31-platform-cheat-detection-data-split-plan.md`
- `code/frontend/src/features/platform/cheat-detection/model/useCheatDetectionData.ts`
- `code/frontend/src/features/platform/cheat-detection/model/useCheatDetectionData.test.ts`
- `code/frontend/src/features/platform/cheat-detection/model/useCheatDetectionPage.ts`
- `code/frontend/src/features/platform/cheat-detection/model/index.ts`
- `code/frontend/src/pages/platform/__tests__/CheatDetection.test.ts`

## After implementation
- 平台作弊检测页的数据加载 owner 会集中到 `useCheatDetectionData`。
- `useCheatDetectionPage` 只保留 route target、快捷操作和格式化逻辑。
