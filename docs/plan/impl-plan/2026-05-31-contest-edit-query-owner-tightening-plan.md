# Contest Edit Query Owner Tightening 计划

## Objective

- 收紧 `useContestEditPage.ts` 的 query owner，让 `ContestEdit` 不再直接读取 `window.location.search`。
- 继续保持 route props + route target contract 已有边界，不扩大到别的 contest 页面。

## Non-goals

- 不改 `useUrlSyncedTabs()` 的共享实现。
- 不调整 AWD workbench 的数据加载、stage 可见性或保存逻辑。
- 不同时处理 `ContestManage`、`ContestOperationsHub`、`ContestProjector`。

## Source Inputs

- `code/frontend/src/features/platform/contests/model/useContestEditPage.ts`
- `code/frontend/src/shared/model/navigation/useRouteQueryTransport.ts`
- `code/frontend/src/features/platform/user-management/model/usePlatformUserManagePage.ts`
- `code/frontend/src/features/teaching/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestEdit.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- 最小正确改动是复用已有 `useRouteQueryTransport()`，只替换 `ContestEdit` 里残留的直接 location query 读取。
- `useUrlSyncedTabs()` 继续负责 tab 的本地状态与 query 写回；本轮不重开共享 tab transport 设计。

## Task Slices

### Slice 1: 收紧 ContestEdit query owner

- 目标：让 `useContestEditPage.ts` 改为通过共享 route query transport 读取当前 stage query。
- 风险：
  - 如果 mock route query 处理不对，会影响测试里旧 `panel=operations` 的兼容行为。

### Slice 2: 补源码级护栏

- 目标：更新 `ContestEdit.test.ts`，锁住 `useContestEditPage.ts` 不再直接引用 `window.location.search`。
- 风险：
  - 如果只看行为不看源码 owner，后续容易再次回流到浏览器全局对象。

### Slice 3: 同步迁移台账

- 目标：更新 backlog 进展，明确 `ContestEdit` 这条 query owner 已进一步收口。
- 风险：
  - 如果台账不更新，后续容易重复判断这条残余债务是否还存在。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision contest-edit-query-owner-tightening`
- `cd code/frontend && npm run test:run -- src/pages/platform/contests/__tests__/ContestEdit.test.ts`
- `git diff --check -- .harness/reuse-decisions/contest-edit-query-owner-tightening.md docs/plan/impl-plan/2026-05-31-contest-edit-query-owner-tightening-plan.md code/frontend/src/features/platform/contests/model/useContestEditPage.ts code/frontend/src/pages/platform/contests/__tests__/ContestEdit.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Review Focus

- `useContestEditPage.ts` 是否只通过共享 transport 读取 route query，而不是重新碰 `window.location.search`。
- `ContestEdit` 对旧 `panel=operations` 的降级行为是否保持不变。
- 本轮是否保持了原有 route props 与 route target contract，不把别的 owner 混进来。

## Rollback / Recovery

- 如果共享 transport 接入导致测试或行为不稳定，可以回退到更薄的 query normalize helper，但不能回退到 feature model 直接读取 `window.location.search`。
