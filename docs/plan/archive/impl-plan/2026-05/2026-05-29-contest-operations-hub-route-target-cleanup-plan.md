> 状态：Current
> 事实源：contest operations hub page model、platform contest ops route view、前端架构 allowlist
> 替代：无

# Contest Operations Hub Route Target Cleanup Plan

## 目标

- 把 `useContestOperationsHubPage.ts` 的“进入运维台 / 返回竞赛目录”导航收口成显式 route target contract。
- 删除 `useContestOperationsHubPage.ts -> vue-router` 这条 allowlist。

## 非目标

- 不处理 `useContestOperationsPage.ts` 自身的 route owner。
- 不改 AWD 赛事目录加载、preferred contest 选择、分页和错误态 owner。
- 不重做赛事运维目录的视觉结构。

## 输入依据

- `code/frontend/src/features/platform-contests/model/useContestOperationsHubPage.ts`
- `code/frontend/src/views/platform/ContestOperationsHub.vue`
- `code/frontend/src/features/platform-contests/ui/ContestOperationsHubHeroPanel.vue`
- `code/frontend/src/features/platform-contests/ui/ContestOperationsHubWorkspacePanel.vue`
- `code/frontend/src/views/platform/__tests__/ContestOperationsHub.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- `useContestOperationsHubPage.ts` 的主 owner 是 AWD 赛事目录加载、preferred contest 选择、分页和刷新。
- “进入单场运维台 / 返回竞赛目录”都是单条薄导航，适合改成 route target helper，而不是继续停留在 `useRouter()`。
- `ContestOperationsHub.vue` 仍然是薄 route shell，适合只透传 route target，不应重新变成 router owner。

## 设计边界

### `useContestOperationsHubPage.ts` 本轮负责

- AWD 赛事目录加载
- preferred contest 选择
- 分页、刷新与错误态

### `contestOperationsHubRoutes.ts` 本轮负责

- 构建赛事运维台 route target
- 构建返回竞赛目录 route target

### `ContestOperationsHub.vue` 本轮负责

- 继续作为薄 route shell 组合 page model
- 把 route target 传给 hero 与目录 panel

### `ContestOperationsHubHeroPanel.vue` / `ContestOperationsHubWorkspacePanel.vue` 本轮负责

- 直接通过 `AppRouteLink` 消费 route target
- 保留 retry 和分页 emit

## 任务切片

### Slice 1：page model 去掉 router

- 目标：
  - 删除 `useContestOperationsHubPage.ts` 里的 `useRouter()` 和两个导航 handler
  - 新增 `contestOperationsHubRoutes.ts`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestOperationsHub.test.ts`
- Review focus：
  - page model 是否不再直接持有 `router.push`

### Slice 2：view / feature UI 直接消费 route target

- 目标：
  - `ContestOperationsHub.vue` 改为透传 route target
  - hero、空态和目录行改用 `AppRouteLink`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestOperationsHub.test.ts`
- Review focus：
  - route target contract 是否清楚，没有把 retry / 分页 workflow 一起迁走

### Slice 3：allowlist / backlog / review 收尾

- 目标：
  - 更新 allowlist、backlog 和 review
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestOperationsHub.test.ts src/__tests__/architectureBoundaries.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - `useContestOperationsHubPage.ts` 是否已经从 feature router allowlist 移除

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestOperationsHub.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-operations-hub-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-contest-operations-hub-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-contest-operations-hub-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-contests/model/useContestOperationsHubPage.ts code/frontend/src/features/platform-contests/model/contestOperationsHubRoutes.ts code/frontend/src/features/platform-contests/model/index.ts code/frontend/src/features/platform-contests/ui/ContestOperationsHubHeroPanel.vue code/frontend/src/features/platform-contests/ui/ContestOperationsHubWorkspacePanel.vue code/frontend/src/views/platform/ContestOperationsHub.vue code/frontend/src/views/platform/__tests__/ContestOperationsHub.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `useContestOperationsPage.ts` 自身的 route owner 不在这轮范围。
- 这轮只处理返回与进入运维台，不继续改赛事运维目录数据加载策略。
- review 仍默认按同上下文 self-review 记录，独立 reviewer gate 仍未满足。
