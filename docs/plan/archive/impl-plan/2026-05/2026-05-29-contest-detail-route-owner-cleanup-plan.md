> 状态：Current
> 事实源：contest detail route page model、shared route transport、前端架构 allowlist
> 替代：无

# Contest Detail Route Owner Cleanup Plan

## 目标

- 收掉 `features/contest-detail/model/useContestDetailRoutePage.ts -> vue-router`

## 非目标

- 不重做 contest detail 的视觉结构、队伍流程或 AWD 工作台。
- 不把 `useContestDetailRoutePage.ts` 再平移成新的 wrapper。
- 不改 `useContestDetailPage.ts` 当前的数据加载和业务 workflow owner。

## 输入依据

- `useContestDetailRoutePage.ts`
- `useContestDetailPage.ts`
- `ContestDetail.vue`
- `ContestDetail.test.ts`
- `useRouteQueryTabs.ts`
- `routeQueryTransport.ts`
- `architectureAllowlist.ts`

## 当前结论

- `contestId`、`challenge` query 和 `panel` tab state 都属于 contest detail route-aware page owner。
- `useRouteQueryTabs()` 已经能承接 tab query transport，不需要继续保留 `useUrlSyncedTabs()` 这层 window-history owner。
- `params / query / replaceQuery` 可以下沉到共享 transport；contest-specific normalize、AWD 默认页签和 challenge 选中同步仍留在 page model。

## 设计边界

### contest detail route page model 本轮负责

- 保留 `contestId` / `selectedChallengeId` 派生
- 保留 AWD 默认页签、challenge query sync 和 contest 状态派生
- 不再直接 import `vue-router`

### shared transport 本轮负责

- 提供 `params`、`query` 与 `replaceQuery()`
- 不承接 contest-specific tab policy 或 challenge normalize

### route view 本轮负责

- 继续作为组合壳，不直接持有 route/query owner

## 任务切片

- [ ] Slice 1：page model 改用 shared transport
  - 目标：
    - `useContestDetailRoutePage.ts` 去掉 `vue-router` 和 `useUrlSyncedTabs`
    - 改为复用 `useRouteQueryTabs()` 与共享 route transport
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/contests/__tests__/ContestDetail.test.ts`

- [ ] Slice 2：allowlist / review / backlog 收尾
  - 目标：
    - 更新 allowlist、测试护栏、todo、review
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/contests/__tests__/ContestDetail.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
    - `cd code/frontend && npm run typecheck`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/contests/__tests__/ContestDetail.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-detail-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-contest-detail-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-contest-detail-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/composables/routeQueryTransport.ts code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts code/frontend/src/views/contests/ContestDetail.vue code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `useRouteQueryTabs()` 接入后，AWD 默认页签会把 `panel=challenges` 真正写回 router query；需要确认这属于可接受的 canonical query，而不是回归。
- `routeQueryTransport.ts` 这轮会从“只管 query”扩成“params + query transport”；必须保持它仍然只是 transport，不吞业务规则。
