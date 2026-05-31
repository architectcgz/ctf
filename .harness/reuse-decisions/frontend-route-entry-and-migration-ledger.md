# Reuse Decision

## Change type
page / layout

## Existing code searched
- `code/frontend/src/pages/platform/challenges/*`
- `code/frontend/src/features/platform/challenges/*`
- `code/frontend/src/router/routes/platformRoutes.ts`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `frontend sliced architecture migration ledger`

## Similar implementations found
- `code/frontend/src/pages/platform/challenges/ChallengeDetailRoutePage.vue`
- `code/frontend/src/pages/platform/challenges/ChallengeTopologyStudioRoutePage.vue`
- `code/frontend/src/features/platform/challenges/ui/ChallengeManagePage.vue`
- `code/frontend/src/router/routes/appShellRoute.ts`

## Decision
extend_existing

## Reason
当前平台题目管理页已经有稳定的 feature UI owner：`ChallengeManagePage.vue`。这次不重写页面逻辑，只补一个标准 `pages` 层 route entry 复用现有 feature 页面，并同步扩展现有架构边界测试，避免 router 继续直接挂到 `features/ui`。

## Files to modify
- `frontend sliced architecture migration ledger`
- `code/frontend/src/pages/platform/challenges/ChallengeManageRoutePage.vue`
- `code/frontend/src/router/routes/platformRoutes.ts`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `code/frontend/src/pages/platform/challenges/__tests__/ChallengeManage.test.ts`

## After implementation
- 本轮模式已经能从现有 `pages/platform/challenges/*RoutePage.vue` 找到，不需要额外写入 `.harness/reuse-index/`。
