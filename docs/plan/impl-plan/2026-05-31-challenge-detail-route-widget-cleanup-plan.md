# 题目详情路由页 widget 收口计划

## Objective

- 把 `ChallengeDetailRoutePage.vue` 从厚 route page 收口成标准 `pages` 组合入口。
- 把题目详情页的 loading / 错误态 / `ChallengeWorkspaceShell` 外壳和本地 shell 样式迁到 `widgets/challenge-detail-workspace`。
- 保持 `features/challenge-detail/model/useChallengeDetailPage.ts` 继续作为 route、数据加载、题解、提交记录、writeup 和实例 workflow owner。

## Non-goals

- 不改 challenge detail API、实例工作流、题解分页和 writeup owner。
- 不在本轮重拆 `ChallengeWorkspaceShell.vue` 内部子面板边界。
- 不处理挑战列表页。

## Source Inputs

- `TODO/frontend-sliced-architecture.md`
- `code/frontend/src/pages/challenges/ChallengeDetailRoutePage.vue`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts`
- `code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test.ts`
- `code/frontend/src/features/challenge-detail/ui/ChallengeWorkspaceShell.vue`

## Task Slices

### Slice 1: 新增题目详情 widget

- 目标：把题目详情页 loading / 错误态 / workspace shell 和本地 shell 样式迁到 `ChallengeDetailWorkspace.vue`。
- 变更面：
  - `code/frontend/src/widgets/challenge-detail-workspace/ChallengeDetailWorkspace.vue`
  - `code/frontend/src/widgets/challenge-detail-workspace/index.ts`
- 风险：
  - props 透传不全会影响题解、提交记录、writeup 或实例操作。
- 验证：
  - 题目详情页测试。

### Slice 2: route page 收口为组合入口

- 目标：让 `ChallengeDetailRoutePage.vue` 只负责 `useChallengeDetailPage()` 和 widget 组合。
- 变更面：
  - `code/frontend/src/pages/challenges/ChallengeDetailRoutePage.vue`
- 风险：
  - route page 如果残留 shell 样式和错误态模板，结构收口不完整。
- 验证：
  - 源码级断言 route page 不直接持有 shell 模板和样式壳。

### Slice 3: 同步测试与台账

- 目标：把模板和样式断言切到 widget owner，并在迁移台账里把题目详情从当前优先项中移出。
- 变更面：
  - `code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test.ts`
  - `code/frontend/src/pages/challenges/__tests__/challengePageUiStrategy.test.ts`
  - `code/frontend/src/pages/__tests__/journalUserShellStyles.test.ts`
  - `code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts`
  - `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`
  - `TODO/frontend-sliced-architecture.md`
- 风险：
  - 测试如果继续盯 route page 旧模板，会产生假失败。
- 验证：
  - `ChallengeDetail.test.ts`
  - 受影响的页面源码级边界测试

## Validation Plan

- `npm run test:run -- src/pages/challenges/__tests__/ChallengeDetail.test.ts src/pages/__tests__/workspaceShellStyles.test.ts src/__tests__/routePageArchitectureBoundary.test.ts src/__tests__/architectureBoundaries.test.ts`
- `npm run test:run -- src/pages/challenges/__tests__/ChallengeDetail.test.ts src/pages/challenges/__tests__/challengePageUiStrategy.test.ts src/pages/__tests__/journalUserShellStyles.test.ts src/pages/__tests__/workspaceShellStyles.test.ts src/pages/__tests__/workspacePageHeaderStyles.test.ts src/__tests__/routePageArchitectureBoundary.test.ts src/__tests__/architectureBoundaries.test.ts`

## Review Focus

- route page 是否已经回到纯组合入口。
- widget 是否只承担页面壳和状态编排，没有拿走 challenge detail feature owner。
- 台账是否同步反映题目详情已完成当前这轮 route page -> widget 收口。

## Rollback / Recovery

- 如果 widget 抽取导致 loading / 错误态 / 实例操作回退，可先把 route page 恢复到原模板，再按更小切片拆外层壳和 shell 样式。
