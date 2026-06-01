> 状态：Current
> 事实源：contest list page model、student contest list route view、前端架构 allowlist
> 替代：无

# Contest List Route Target Cleanup Plan

## 目标

- 把 `useContestListPage.ts` 的竞赛详情导航收口成显式 route target contract。
- 删除 `useContestListPage.ts -> vue-router` 这条 allowlist。

## 非目标

- 不处理 `ContestDetail` 的 query / tab route owner。
- 不继续处理 `challenge-list`、`contest-manage` 或其它 feature router allowlist。
- 不改 contest list 的视觉结构、筛选行为和分页 owner。

## 输入依据

- `code/frontend/src/features/contest-detail/model/useContestListPage.ts`
- `code/frontend/src/views/contests/ContestList.vue`
- `code/frontend/src/views/contests/__tests__/ContestList.test.ts`
- `code/frontend/src/router/routes/studentRoutes.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- `useContestListPage.ts` 的主 owner 是竞赛目录数据加载、筛选、分页和展示格式化。
- “进入竞赛详情”只是单条薄导航，适合改成 route target helper，而不是继续停留在 `useRouter()`。
- `ContestList.vue` 仍然是薄 route shell，直接消费 `AppRouteLink` 不会引入新的 owner 漂移。

## 设计边界

### `useContestListPage.ts` 本轮负责

- 竞赛目录加载、筛选、分页
- 摘要指标、时间格式化、状态/模式文案
- 错误态与 refresh owner

### `contestListRoutes.ts` 本轮负责

- 构建竞赛详情 route target

### `ContestList.vue` 本轮负责

- 继续作为薄 route shell 组合 page model
- 目录行直接消费竞赛详情 route target

## 任务切片

### Slice 1：page model 去掉 router

- 目标：
  - 删除 `useContestListPage.ts` 里的 `useRouter()` 和 `openContest()`
  - 新增 `contestListRoutes.ts`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/contests/__tests__/ContestList.test.ts`
- Review focus：
  - page model 是否不再直接持有 `router.push`

### Slice 2：route view 直接消费 route target

- 目标：
  - `ContestList.vue` 改用 `AppRouteLink`
  - 行级点击行为保持不变
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/contests/__tests__/ContestList.test.ts`
- Review focus：
  - 详情入口是否已经通过显式 route target 渲染

### Slice 3：allowlist / backlog / review 收尾

- 目标：
  - 更新 allowlist、backlog 和 review
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/contests/__tests__/ContestList.test.ts src/__tests__/architectureBoundaries.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - `useContestListPage.ts` 是否已经从 feature router allowlist 移除

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/contests/__tests__/ContestList.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-list-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-contest-list-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-contest-list-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/contest-detail/model/useContestListPage.ts code/frontend/src/features/contest-detail/model/contestListRoutes.ts code/frontend/src/features/contest-detail/model/index.ts code/frontend/src/features/contest-detail/index.ts code/frontend/src/views/contests/ContestList.vue code/frontend/src/views/contests/__tests__/ContestList.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收 contest list 的单条详情路由，不继续处理 `ContestDetail` 的 route owner。
- review 仍默认按同上下文 self-review 记录，独立 reviewer gate 仍未满足。
