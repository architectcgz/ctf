> 状态：Current
> 事实源：contest announcements page model、platform contest announcements route view、前端架构 allowlist
> 替代：无

# Contest Announcements Route Target Cleanup Plan

## 目标

- 把 `useContestAnnouncementsPage.ts` 的“返回竞赛工作台”导航收口成显式 route target contract。
- 删除 `useContestAnnouncementsPage.ts -> vue-router` 这条 allowlist。

## 非目标

- 不改公告列表加载、发布和删除 workflow owner。
- 不重做单场公告页的视觉结构。
- 不处理 `useContestEditPage.ts` 自身的 route owner。

## 输入依据

- `code/frontend/src/features/platform-contests/model/useContestAnnouncementsPage.ts`
- `code/frontend/src/views/platform/ContestAnnouncements.vue`
- `code/frontend/src/features/platform-contests/ui/ContestAnnouncementsTopbarPanel.vue`
- `code/frontend/src/views/platform/__tests__/ContestAnnouncements.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- `useContestAnnouncementsPage.ts` 的主 owner 是竞赛详情加载、公告列表 workflow、发布和删除操作。
- “返回竞赛工作台”只是单条薄导航，适合改成 route target helper，而不是继续停留在 `useRouter()`。
- `ContestAnnouncements.vue` 仍然是薄 route shell，适合直接消费 route target，不应重新变成 router owner。

## 设计边界

### `useContestAnnouncementsPage.ts` 本轮负责

- 竞赛详情加载
- 公告列表 workflow
- 发布和删除操作
- 页面错误态与时间格式化

### 本轮 route target contract 负责

- 构建返回竞赛编辑页 route target

### `ContestAnnouncements.vue` / `ContestAnnouncementsTopbarPanel.vue` 本轮负责

- 直接通过 `AppRouteLink` 消费返回路由
- 保留现有公告页组合和错误态渲染

## 任务切片

### Slice 1：page model 去掉 router

- 目标：
  - 删除 `useContestAnnouncementsPage.ts` 里的 `useRouter()` 和 `goBackToStudio()`
  - 复用现有 `buildContestEditRoute()`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAnnouncements.test.ts`
- Review focus：
  - page model 是否不再直接持有 `router.push`

### Slice 2：view / feature UI 直接消费 route target

- 目标：
  - topbar 返回按钮改用 `AppRouteLink`
  - 错误空态返回按钮改用 `AppRouteLink`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAnnouncements.test.ts`
- Review focus：
  - route target contract 是否清楚，没有把公告 workflow 一起迁走

### Slice 3：allowlist / backlog / review 收尾

- 目标：
  - 更新 allowlist、backlog 和 review
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAnnouncements.test.ts src/__tests__/architectureBoundaries.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - `useContestAnnouncementsPage.ts` 是否已经从 feature router allowlist 移除

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAnnouncements.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-announcements-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-contest-announcements-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-contest-announcements-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-contests/model/useContestAnnouncementsPage.ts code/frontend/src/features/platform-contests/ui/ContestAnnouncementsTopbarPanel.vue code/frontend/src/views/platform/ContestAnnouncements.vue code/frontend/src/views/platform/__tests__/ContestAnnouncements.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `useContestEditPage.ts` 自身的 route owner 不在这轮范围。
- 这轮只处理返回入口，不继续改公告页的数据和 mutation workflow。
- review 仍默认按同上下文 self-review 记录，独立 reviewer gate 仍未满足。
