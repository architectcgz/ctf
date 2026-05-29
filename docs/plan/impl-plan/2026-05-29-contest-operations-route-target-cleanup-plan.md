> 状态：Current
> 事实源：contest operations page model、platform contest operations route view、前端架构 allowlist
> 替代：无

# Contest Operations Route Target Cleanup Plan

## 目标

- 把 `useContestOperationsPage.ts` 的 `contestId` 输入 owner 收口成显式 route props contract。
- 删除 `useContestOperationsPage.ts -> vue-router` 这条 allowlist。

## 非目标

- 不处理 `useContestEditPage.ts` 自身的 route owner。
- 不改 AWD 运维页里的数据加载策略、breadcrumb workflow 和 runtime/readiness 判定 owner。
- 不重做单场运维页的视觉结构。

## 输入依据

- `code/frontend/src/features/platform-contests/model/useContestOperationsPage.ts`
- `code/frontend/src/views/platform/ContestOperations.vue`
- `code/frontend/src/views/platform/__tests__/ContestOperations.test.ts`
- `code/frontend/src/router/routes/platformRoutes.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- `useContestOperationsPage.ts` 的主 owner 是单场赛事详情加载、breadcrumb detail 标题维护，以及运维面板 runtime/readiness 模式判定。
- `contestId` 只是 route path param 输入，不该继续停留在 `useRoute()`。
- `ContestOperations.vue` 已经是薄 route shell，适合直接消费 route props，不应重新变成 router owner。

## 设计边界

### `platformRoutes.ts` 本轮负责

- 从 `route.params.id` 构建 `contestId` props

### `ContestOperations.vue` 本轮负责

- 定义 `contestId` props
- 把 `contestId` 转成 `Ref` 传给 page model
- 保留现有运维页组合结构

### `useContestOperationsPage.ts` 本轮负责

- 单场赛事详情加载
- breadcrumb detail 标题维护
- runtime/readiness 模式判定

## 任务切片

### Slice 1：route props contract

- 目标：
  - `ContestOperations` 路由显式下传 `contestId`
  - route view 改为 `defineProps`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestOperations.test.ts`
- Review focus：
  - route view 是否仍然保持薄组合，而不是重新拿 router

### Slice 2：page model 去掉 router

- 目标：
  - 删除 `useContestOperationsPage.ts` 里的 `useRoute()`
  - page model 直接接收 `Ref<string> | ComputedRef<string>`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestOperations.test.ts`
- Review focus：
  - feature model 是否不再直接 import `vue-router`

### Slice 3：allowlist / backlog / review 收尾

- 目标：
  - 更新 allowlist、backlog 和 review
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestOperations.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - `useContestOperationsPage.ts` 是否已经从 feature router allowlist 移除

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestOperations.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-operations-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-contest-operations-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-contest-operations-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-contests/model/useContestOperationsPage.ts code/frontend/src/views/platform/ContestOperations.vue code/frontend/src/views/platform/__tests__/ContestOperations.test.ts code/frontend/src/router/routes/platformRoutes.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `useContestEditPage.ts` 仍会是 `platform-contests` 当前剩余的更深 route owner。
- 这轮只处理 route param 输入，不继续改运维页的加载时机和 breadcrumb lifecycle。
- review 仍默认按同上下文 self-review 记录，独立 reviewer gate 仍未满足。
