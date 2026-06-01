> 状态：Current
> 事实源：contest manage page model、platform contest manage route view、前端架构 allowlist
> 替代：无

# Contest Manage Route Target Cleanup Plan

## 目标

- 把 `useContestManagePage.ts` 的编辑页 / 运维台 / 公告完整页导航收口成显式 route target contract。
- 删除 `useContestManagePage.ts -> vue-router` 这条 allowlist。

## 非目标

- 不处理 `useContestEditPage.ts`、`useContestOperationsPage.ts`、`useContestAnnouncementsPage.ts` 自身的 route owner。
- 不改竞赛管理页的数据加载、状态筛选、分页和创建流程 owner。
- 不重做公告抽屉交互，只收口“进入完整管理页”这条薄导航。

## 输入依据

- `code/frontend/src/features/platform-contests/model/useContestManagePage.ts`
- `code/frontend/src/views/platform/ContestManage.vue`
- `code/frontend/src/features/platform-contests/ui/ContestOrchestrationPage.vue`
- `code/frontend/src/features/platform-contests/ui/PlatformContestTable.vue`
- `code/frontend/src/features/contest-announcements/ui/ContestAnnouncementManageDrawer.vue`
- `code/frontend/src/views/platform/__tests__/ContestManage.test.ts`
- `code/frontend/src/components/platform/__tests__/PlatformContestTable.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- `useContestManagePage.ts` 的主 owner 是竞赛目录加载、筛选、创建流程、公告抽屉开关和 AWD readiness override dialog。
- 编辑 / 运维 / 公告完整页都只是单条薄导航，适合改成 route target helper，而不是继续停留在 `useRouter()`。
- `ContestManage.vue` 仍然是薄 route shell，适合只透传 route builder，不应重新变成 router owner。

## 设计边界

### `useContestManagePage.ts` 本轮负责

- 竞赛目录加载、筛选和分页
- 创建 / 编辑 dialog workflow owner
- 公告抽屉开关与 active contest owner
- AWD readiness override dialog owner

### `contestManageRoutes.ts` 本轮负责

- 构建竞赛编辑页 route target
- 构建竞赛运维台 route target
- 构建竞赛公告完整页 route target

### `ContestManage.vue` / `ContestOrchestrationPage.vue` 本轮负责

- 继续作为薄 route shell / feature shell 组合 page model
- 把 route builder 传给表格和公告抽屉

### `PlatformContestTable.vue` / `ContestAnnouncementManageDrawer.vue` 本轮负责

- 直接通过 `AppRouteLink` 消费 route target
- 保留现有公告抽屉打开、关闭和菜单 workflow

## 任务切片

### Slice 1：page model 去掉 router

- 目标：
  - 删除 `useContestManagePage.ts` 里的 `useRouter()` 和三条 `open*Page`
  - 新增 `contestManageRoutes.ts`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts`
- Review focus：
  - page model 是否不再直接持有 `router.push`

### Slice 2：view / feature UI 直接消费 route target

- 目标：
  - `ContestManage.vue` / `ContestOrchestrationPage.vue` 改为透传 route builder
  - `PlatformContestTable.vue` 与 `ContestAnnouncementManageDrawer.vue` 改用 `AppRouteLink`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts src/components/platform/__tests__/PlatformContestTable.test.ts`
- Review focus：
  - route target contract 是否清楚，没有把公告抽屉关闭或创建竞赛 workflow 一起迁走

### Slice 3：allowlist / backlog / review 收尾

- 目标：
  - 更新 allowlist、backlog 和 review
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts src/components/platform/__tests__/PlatformContestTable.test.ts src/__tests__/architectureBoundaries.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - `useContestManagePage.ts` 是否已经从 feature router allowlist 移除

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts src/components/platform/__tests__/PlatformContestTable.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-manage-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-contest-manage-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-contest-manage-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-contests/model/useContestManagePage.ts code/frontend/src/features/platform-contests/model/contestManageRoutes.ts code/frontend/src/features/platform-contests/model/index.ts code/frontend/src/features/platform-contests/ui/ContestOrchestrationPage.vue code/frontend/src/features/platform-contests/ui/PlatformContestTable.vue code/frontend/src/features/contest-announcements/ui/ContestAnnouncementManageDrawer.vue code/frontend/src/views/platform/ContestManage.vue code/frontend/src/views/platform/__tests__/ContestManage.test.ts code/frontend/src/components/platform/__tests__/PlatformContestTable.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收 contest manage 的三条薄导航，不继续处理 `useContestEditPage.ts`、`useContestOperationsPage.ts`、`useContestAnnouncementsPage.ts` 的 route owner。
- 公告抽屉点击“进入完整管理页”后仍需维持 close 行为，测试要显式覆盖。
- review 仍默认按同上下文 self-review 记录，独立 reviewer gate 仍未满足。
