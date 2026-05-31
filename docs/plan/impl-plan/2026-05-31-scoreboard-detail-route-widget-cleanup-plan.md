# 排行详情路由页 widget 收口计划

## Objective

- 把 `ScoreboardDetailRoutePage.vue` 从厚 route page 收口成标准 `pages` 组合入口。
- 把排行详情页的头部、概况卡、目录、空态和分页壳迁到 `widgets/scoreboard-detail-workspace`。
- 保持 `features/scoreboard/model/useScoreboardDetailPage.ts` 继续作为加载、静默刷新、实时刷新和分页 owner。

## Non-goals

- 不改排行榜 API、realtime bridge、toast 或分页 owner。
- 不在本轮重拆 `useScoreboardDetailPage.ts` 的刷新与错误策略。
- 不处理 `ChallengeDetailRoutePage.vue`。

## Source Inputs

- `TODO/frontend-sliced-architecture.md`
- `code/frontend/src/pages/scoreboard/ScoreboardDetailRoutePage.vue`
- `code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts`
- `code/frontend/src/pages/scoreboard/__tests__/ScoreboardView.test.ts`
- `code/frontend/src/features/scoreboard/ui/ScoreboardWorkspaceShell.vue`

## Task Slices

### Slice 1: 新增排行详情 widget

- 目标：把排行详情页头、概况卡、目录、空态和分页壳层迁到 `ScoreboardDetailWorkspace.vue`。
- 变更面：
  - `code/frontend/src/widgets/scoreboard-detail-workspace/ScoreboardDetailWorkspace.vue`
  - `code/frontend/src/widgets/scoreboard-detail-workspace/index.ts`
- 风险：
  - props 透传不全会影响手动刷新、realtime 刷新和分页行为。
- 验证：
  - 排行页测试。

### Slice 2: route page 收口为组合入口

- 目标：让 `ScoreboardDetailRoutePage.vue` 只负责 `useScoreboardDetailPage()` 和 widget 组合。
- 变更面：
  - `code/frontend/src/pages/scoreboard/ScoreboardDetailRoutePage.vue`
- 风险：
  - route page 如果残留概况卡和目录模板，结构收口不完整。
- 验证：
  - 源码级断言 route page 不直接持有排行详情模板和样式壳。

### Slice 3: 同步测试与台账

- 目标：把模板和样式断言切到 widget owner，并在迁移台账里把排行详情从当前优先项中移出。
- 变更面：
  - `code/frontend/src/pages/scoreboard/__tests__/ScoreboardView.test.ts`
  - `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`
  - `code/frontend/src/pages/__tests__/studentDirectoryTypographyBoundary.test.ts`
  - `TODO/frontend-sliced-architecture.md`
- 风险：
  - 源码级测试如果继续盯 route page 旧模板，会产生假失败。
- 验证：
  - `ScoreboardView.test.ts`
  - 受影响的页面源码级边界测试

## Validation Plan

- `npm run test:run -- src/pages/scoreboard/__tests__/ScoreboardView.test.ts src/pages/__tests__/workspaceShellStyles.test.ts src/pages/__tests__/studentDirectoryTypographyBoundary.test.ts src/__tests__/routePageArchitectureBoundary.test.ts src/__tests__/architectureBoundaries.test.ts`

## Review Focus

- route page 是否已经回到纯组合入口。
- widget 是否只承担页面壳和区块编排，没有拿走 scoreboard feature owner。
- 台账是否同步反映排行详情已完成当前这轮 route page -> widget 收口。

## Rollback / Recovery

- 如果 widget 抽取导致刷新、realtime 或分页回退，可先把 route page 恢复到原模板，再按更小切片拆头部区和目录区。
