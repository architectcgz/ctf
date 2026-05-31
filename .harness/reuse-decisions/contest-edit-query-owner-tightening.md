# Reuse Decision

## Change type
frontend refactor / contest edit query owner tightening

## Existing code searched
- `code/frontend/src/features`
- `code/frontend/src/features/platform/contests/model/useContestEditPage.ts`
- `code/frontend/src/features/platform/contests/model/useContestOperationsPage.ts`
- `code/frontend/src/features/platform/user-management/model/usePlatformUserManagePage.ts`
- `code/frontend/src/features/teaching/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/shared/model/navigation/useRouteQueryTransport.ts`
- `code/frontend/src/shared/model/navigation/useUrlSyncedTabs.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestEdit.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- 前端 feature-sliced architecture 迁移台账

## Similar implementations found
- `usePlatformUserManagePage.ts` 已经把 `panel` query 读取与切换收口到 `useRouteQueryTransport()`。
- `useClassStudentsPage.ts` 与 `useStudentAnalysisPage.ts` 也都通过共享 route query transport 持有 route-aware query owner。
- `useContestEditPage.ts` 虽然已经不再直接 import `vue-router`，但仍然直接读取 `window.location.search` 判断 stage，和当前 route-aware feature 的 shared transport 模式不一致。

## Decision
refactor_existing

## Reason
当前最小正确切片不是重写 `ContestEdit` 全部 stage 逻辑，而是把这条残留的 query owner 收回共享 transport：

- 保留 `contestId` route props、返回目录 route target、公告页 route target 和保存成功 redirect 的现有 owner。
- 只把 `syncWorkbenchStageSelection()` 里直接读取 `window.location.search` 的残留 query owner 改为消费 `useRouteQueryTransport()`。
- 用源码级测试锁住 `useContestEditPage.ts` 不再自己碰浏览器 location query。

本轮不做：

- 不改 `useUrlSyncedTabs()` 的内部实现。
- 不重写 AWD workbench 数据加载、tab 可见性推导或保存 workflow。
- 不扩到 `ContestManage`、`ContestOperationsHub`、`ContestProjector` 等其他赛事页面。

## Files to modify
- `.harness/reuse-decisions/contest-edit-query-owner-tightening.md`
- `docs/plan/impl-plan/2026-05-31-contest-edit-query-owner-tightening-plan.md`
- `code/frontend/src/features/platform/contests/model/useContestEditPage.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestEdit.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `ContestEdit` 的 stage query owner 会和当前其他 route-aware feature 一样，通过共享 `useRouteQueryTransport()` 读取 route query。
- `useContestEditPage.ts` 不再直接依赖 `window.location.search`。
- `platform/contests` 在 query owner 这条残余结构债上会再收紧一层，避免后续 route cleanup 回流。
