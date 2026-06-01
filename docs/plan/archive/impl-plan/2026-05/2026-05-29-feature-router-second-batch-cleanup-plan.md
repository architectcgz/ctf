> 状态：Current
> 事实源：第二批 feature route-aware page models、query-tab composable contract、前端架构 allowlist
> 替代：无

# Feature Router Second Batch Cleanup Plan

## 目标

- 收掉第二批中复杂度 `featureRouterImportAllowlist`：
  - `useNotificationDetailPage.ts`
  - `useScoreboardRoutePage.ts`
  - `usePlatformChallengeDetailRoutePage.ts`

## 非目标

- 不处理 `useChallengeListPage.ts` 的 query/filter sync owner。
- 不重做通知详情与排行榜页的视觉结构。
- 不引入新的 router helper 桶或临时 bridge。

## 输入依据

- 两个 feature model 源文件
- 对应 route view
- 对应测试文件
- `useRouteQueryTabs.ts`
- `studentRoutes.ts`
- `architectureAllowlist.ts`

## 当前结论

- `notification detail` 既有 route param 输入，也有站内 / 站外导航分流，应拆成 route props + route target + external transport。
- `scoreboard route page` 与 `platform challenge detail route page` 本质都是 query-tab owner，但当前都只是把完整 `vue-router` 对象透传给 `useRouteQueryTabs()`。
- `useRouteQueryTabs()` 更适合直接成为共享 query-tab route owner，而不是继续让每个 feature route page 重复承接 router transport。
- `challenge list` 仍混 query/filter sync 与导航，切片风险高于这轮目标，明确留待下一批。

## 设计边界

### route config 本轮负责

- `notifications/:id` 显式下传 `id` props

### feature model 本轮负责

- `useNotificationDetailPage.ts`
  - 保留通知详情加载、标记已读、彩蛋消息和外链打开
  - 不再直接 import `vue-router`
  - 对外暴露返回列表与站内关联对象的 route target
- `useScoreboardRoutePage.ts`
  - 保留 scoreboard panel tabs 的 query-tab owner
  - 不再直接 import `vue-router`
- `usePlatformChallengeDetailRoutePage.ts`
  - 保留 admin challenge detail 的 panel tab owner
  - 不再直接 import `vue-router`

### composable contract 本轮负责

- `useRouteQueryTabs.ts`
  - 直接承接 `useRoute/useRouter`
  - 成为共享 query-tab route owner
  - 不承接 route param、API 或其它 page workflow

### route view 本轮负责

- `NotificationDetail.vue` 只组合 page model，并通过 props 注入 `id`
- `ScoreboardView.vue` 只组合 page model，不再额外透传 router transport

## 任务切片

- [ ] Slice 1：补 route props 与站内 route target
  - 目标：
    - `NotificationDetail` route param 下沉为 props
    - 返回通知列表与站内关联对象改成 route target
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/notifications/__tests__/NotificationDetail.test.ts`

- [ ] Slice 2：query-tab contract 收口
  - 目标：
    - `useRouteQueryTabs()` 改成共享 query-tab route owner
    - `useScoreboardRoutePage.ts`、`usePlatformChallengeDetailRoutePage.ts` 去掉 `vue-router`
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/scoreboard/__tests__/ScoreboardView.test.ts src/views/__tests__/routeQueryTabsAdoption.test.ts`

- [ ] Slice 3：allowlist / backlog / review 收尾
  - 目标：
    - 更新 allowlist、todo、review
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/notifications/__tests__/NotificationDetail.test.ts src/views/scoreboard/__tests__/ScoreboardView.test.ts src/views/__tests__/routeQueryTabsAdoption.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
    - `cd code/frontend && npm run typecheck`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/notifications/__tests__/NotificationDetail.test.ts src/views/scoreboard/__tests__/ScoreboardView.test.ts src/views/__tests__/routeQueryTabsAdoption.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/feature-router-second-batch-cleanup.md docs/plan/impl-plan/2026-05-29-feature-router-second-batch-cleanup-plan.md docs/reviews/frontend/2026-05-29-feature-router-second-batch-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/composables/useRouteQueryTabs.ts code/frontend/src/router/routes/studentRoutes.ts code/frontend/src/features/notifications/model/useNotificationDetailPage.ts code/frontend/src/features/scoreboard/model/useScoreboardRoutePage.ts code/frontend/src/features/platform-challenge-detail/model/usePlatformChallengeDetailRoutePage.ts code/frontend/src/views/notifications/NotificationDetail.vue code/frontend/src/views/notifications/__tests__/NotificationDetail.test.ts code/frontend/src/views/scoreboard/__tests__/ScoreboardView.test.ts code/frontend/src/views/__tests__/routeQueryTabsAdoption.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `NotificationDetail` 的站内关联对象既可能是命名路由，也可能是 path string，需要确认 route target 对 string path 仍直接兼容。
- `useRouteQueryTabs()` 承接 route owner 后，要确认没有影响其它调用点。
- `ChallengeList` 仍留在 allowlist，后续需要单独处理 query/filter owner。
