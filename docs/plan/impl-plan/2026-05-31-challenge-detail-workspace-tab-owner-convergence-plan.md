# Challenge Detail Workspace Tab Owner Convergence 计划

## Objective

- 把 challenge detail workspace 的 `panel` query owner 从 `useUrlSyncedTabs()` 收敛到共享 `useRouteQueryTabs()`。
- 保持 challenge 数据加载、solution tab、本地交互和实例 workflow owner 不变。

## Non-goals

- 不改 `recommended/community` solution tab 的本地切换 owner。
- 不改题目详情、题解、提交记录、我的题解和实例链路的加载 / 保存策略。
- 不扩到 `platform challenge detail` 或 `contest-edit`。

## Source Inputs

- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts`
- `code/frontend/src/pages/challenges/ChallengeDetailRoutePage.vue`
- `code/frontend/src/widgets/challenge-detail-workspace/ChallengeDetailWorkspace.vue`
- `code/frontend/src/shared/model/navigation/useRouteQueryTabs.ts`
- `code/frontend/src/shared/model/navigation/useUrlSyncedTabs.ts`
- `code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- 这条线优先直接复用 `useRouteQueryTabs()`，不再继续维护 `useUrlSyncedTabs()` 在 challenge detail page model 里的特例。
- 如果 route page / widget 仍需要 `setTabButtonRef` 和 `handleWorkspaceTabKeydown`，应该像已收口页面那样把键盘导航收回 UI 壳或 route shell，而不是继续让 page model同时持有 route query owner 和 DOM focus owner。

## Task Slices

### Slice 1: 收敛 page model 的 workspace tab query owner

- 目标：`useChallengeDetailPage.ts` 改成通过 `useRouteQueryTabs()` 持有 `question/solution/records/writeup` 的 `panel` query。
- 风险：
  - 需要确认现有 `workspaceTabs`、默认 `question` tab 和 `?panel=` hydrate 行为不变。

### Slice 2: 让 route page / widget 退回展示桥接

- 目标：把 workspace tab 的 button ref / keydown owner 从 page model 移回更合适的 UI 层，同时保持 solution tab 的本地键盘导航不变。
- 风险：
  - 如果桥接契约没收清，可能会把 route-aware owner 和 DOM focus owner 再次混回 page model。

### Slice 3: 补 challenge detail 护栏测试

- 目标：补 source 护栏以及 `?panel=` 初始恢复、点击 tab 回写 / 切回默认 tab 清理 query 的运行态断言。
- 风险：
  - challenge detail 测试较大，需要尽量局部补断言，避免为了这条 owner 收口重写整套用例。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision challenge-detail-workspace-tab-owner-convergence`
- `cd code/frontend && npm run test:run -- src/pages/challenges/__tests__/ChallengeDetail.test.ts`
- `git diff --check -- .harness/reuse-decisions/challenge-detail-workspace-tab-owner-convergence.md docs/plan/impl-plan/2026-05-31-challenge-detail-workspace-tab-owner-convergence-plan.md code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts code/frontend/src/pages/challenges/ChallengeDetailRoutePage.vue code/frontend/src/widgets/challenge-detail-workspace/ChallengeDetailWorkspace.vue code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Review Focus

- `useChallengeDetailPage.ts` 是否停止直接依赖 `useUrlSyncedTabs()`。
- route page / widget 是否只保留展示桥接和必要的键盘导航契约。
- `?panel=` hydrate、非默认 tab 写回和默认 tab 清理 query 是否都保持正确。

## Rollback / Recovery

- 如果 route page / widget 的 tab 契约命名还需要微调，可以继续改 props / emits，但不能再回退成 page model 直接绑定 `useUrlSyncedTabs()`。
