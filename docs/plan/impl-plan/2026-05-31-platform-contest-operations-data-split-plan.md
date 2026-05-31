# platform contest operations data split 计划

## Objective

- 在 `platform/contests` 内新增 `useContestOperationsData`，承接单场运维页的数据 owner。
- 让 `useContestOperationsPage` 只保留 breadcrumb、toast 和 runtime 内容派生。

## Non-goals

- 不改 `ContestOperationsRoutePage.vue` UI。
- 不改 `AWDOperationsPanel` contract。
- 不改 breadcrumb / toast hook 行为。

## Source Inputs

- `code/frontend/src/features/platform/contests/model/useContestOperationsPage.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestOperations.test.ts`
- `code/frontend/src/features/platform/contests/model/useContestAnnouncementsData.ts`

## Plan Review Result

- 这页适合做 `page + data` 拆分。
- data owner 负责单场竞赛请求、loading 和错误状态。
- page model 保留 breadcrumb / toast side effect 和 runtime 内容派生。

## Task Slices

### Slice 1: 新建 useContestOperationsData

- 目标：收口单场竞赛请求、loading 和错误状态。
- 风险：
  - 如果把 breadcrumb / toast 一起搬走，会把页面壳 side effect 混进数据 owner。

### Slice 2: useContestOperationsPage 改为消费 data owner

- 目标：保留 breadcrumb、toast 和 runtime 内容派生。
- 风险：
  - 如果 page 继续直接依赖 `getContest`，就没有真正收口数据 owner。

### Slice 3: 更新源码级和行为测试

- 目标：给新 data owner 补直测，并更新平台端源码断言。
- 风险：
  - 不补缺少 contestId / 加载失败测试，页级 loading owner 还会回流进 page model。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision platform-contest-operations-data-split`
- `npm run test:run -- src/features/platform/contests/model/useContestOperationsData.test.ts src/pages/platform/contests/__tests__/ContestOperations.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `npm run typecheck`
- `git diff --check`
- `bash scripts/check-reuse-first.sh`

## Review Focus

- `useContestOperationsData` 是否只承接单场竞赛数据 owner。
- `useContestOperationsPage` 是否只剩 breadcrumb、toast 和 runtime 内容派生。
- route page 是否继续只负责组合。

## Rollback / Recovery

- 如果 `useContestOperationsData` 的接口不顺手，可以调整返回结构，但数据加载 owner 仍必须留在新 composable。
