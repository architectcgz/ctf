> 状态：Current
> 事实源：contest edit page model、contest edit route view、platform contest route targets、前端架构 allowlist
> 替代：无

# Contest Edit Route Target Cleanup Plan

## 目标

- 把 `useContestEditPage.ts` 的 route param 与薄导航收口成显式 route target contract。
- 删除 `useContestEditPage.ts -> vue-router` 这条 allowlist。

## 非目标

- 不重做竞赛编辑页的 AWD workbench 数据加载策略。
- 不修改 `useUrlSyncedTabs()` 这类 query-tab owner。
- 不重做竞赛编辑页的视觉结构。

## 输入依据

- `code/frontend/src/features/platform-contests/model/useContestEditPage.ts`
- `code/frontend/src/views/platform/ContestEdit.vue`
- `code/frontend/src/features/platform-contests/ui/ContestEditTopbarPanel.vue`
- `code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue`
- `code/frontend/src/features/platform-contests/ui/AWDChallengeConfigPanel.vue`
- `code/frontend/src/features/platform-contests/ui/ContestAwdPreflightPanel.vue`
- `code/frontend/src/views/platform/__tests__/ContestEdit.test.ts`
- `code/frontend/src/router/routes/platformRoutes.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- `useContestEditPage.ts` 的主 owner 仍然应该是竞赛详情加载、AWD 数据刷新、保存 workflow 和 stage 派生。
- 返回竞赛目录、打开公告页、进入 AWD 配置页，都只是薄导航，适合改成 route target contract。
- 保存成功后的“返回目录”仍是 page workflow 的一部分，但真正执行导航不必再由 page model 直接拿 `router.push()`。

## 设计边界

### `platformRoutes.ts` 本轮负责

- 从 `route.params.id` 构建 `contestId` props

### `contestManageRoutes.ts` / `contestOperationsHubRoutes.ts` 本轮负责

- 构建竞赛目录 / 公告页 / 编辑页 / 运维页 / AWD 配置页 route target

### `useContestEditPage.ts` 本轮负责

- 竞赛详情加载
- AWD workbench 数据与 stage 派生
- 保存 workflow
- 解析返回目录、公告页、AWD 配置页 route target

### `ContestEdit.vue` 本轮负责

- route props -> page model 的薄组合
- 承接保存成功后的导航 transport

### `ContestEdit*` / `AWD*` feature UI 本轮负责

- 直接消费薄导航 route target
- 保留现有 form / orchestration / preflight workflow emit

## 任务切片

### Slice 1：route props + page model 去掉 router

- 目标：
  - `ContestEdit` 路由显式下传 `contestId`
  - `useContestEditPage.ts` 删除 `useRoute()` / `useRouter()`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestEdit.test.ts`
- Review focus：
  - page model 是否只保留数据 / workflow owner

### Slice 2：route target contract 下沉到 feature UI

- 目标：
  - topbar、错误空态、AWD 配置目录、赛前检查动作都改成显式 route target
  - 保存成功跳转改由独立 navigation transport 承接
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestEdit.test.ts`
- Review focus：
  - 薄导航是否已经不再依赖 emit -> router.push 的隐式链路

### Slice 3：allowlist / backlog / review 收尾

- 目标：
  - 更新 allowlist、backlog 和 review
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestEdit.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - `platform-contests` 是否清空这组 feature router allowlist

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestEdit.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-edit-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-contest-edit-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-contest-edit-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/router/routes/platformRoutes.ts code/frontend/src/features/platform-contests/model/contestManageRoutes.ts code/frontend/src/features/platform-contests/model/contestOperationsHubRoutes.ts code/frontend/src/features/platform-contests/model/useContestEditPage.ts code/frontend/src/views/platform/ContestEdit.vue code/frontend/src/views/platform/__tests__/ContestEdit.test.ts code/frontend/src/features/platform-contests/ui/ContestEditTopbarPanel.vue code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue code/frontend/src/features/platform-contests/ui/AWDChallengeConfigPanel.vue code/frontend/src/features/platform-contests/ui/AWDChallengeConfigDirectorySection.vue code/frontend/src/features/platform-contests/ui/AWDChallengeConfigDirectoryRow.vue code/frontend/src/features/platform-contests/ui/ContestAwdPreflightPanel.vue code/frontend/src/features/awd-readiness/ui/AWDReadinessChecklist.vue code/frontend/src/components/navigation/AppRouteRedirect.vue`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `useUrlSyncedTabs()` 仍然是现有 query-tab owner，不在本轮范围。
- 这轮不会继续改竞赛编辑页的数据聚合和 stage 逻辑，只处理 route target / navigation transport。
- review 仍默认按同上下文 self-review 记录，独立 reviewer gate 仍未满足。
