> 状态：Current
> 事实源：第一批 feature route-aware page models、对应 route views、前端架构 allowlist
> 替代：无

# Feature Router First Batch Cleanup Plan

## 目标

- 收掉第一批低复杂度 `featureRouterImportAllowlist`：
  - `useScoreboardDetailPage.ts`
  - `useAwdChallengeLibraryPage.ts`
  - `useRegisterPage.ts`
  - `useChallengeImportManagePage.ts`
  - `useChallengeImportPreviewPage.ts`

## 非目标

- 不处理 `useRouteQueryTabs` / `useUrlSyncedTabs` 这类 query-tab owner。
- 不进入第二批中等复杂度页面。
- 不重做上述页面的视觉结构。

## 输入依据

- 五个 feature model 源文件
- 对应五个 route view
- 对应测试文件
- `authRoutes.ts`、`studentRoutes.ts`、`platformRoutes.ts`
- `architectureAllowlist.ts`

## 当前结论

- `scoreboard detail` 只是 route param 输入 owner。
- `awd challenge library` 只是导入页入口薄导航。
- `register` 是 mutation 成功后的跳转。
- `challenge import manage / preview` 主要是导入页、预览页、队列与题库目录之间的薄导航。

这五条都适合按统一模式收口：

- route props owner
- route target contract
- redirect transport

## 设计边界

### route config 本轮负责

- `scoreboard/:contestId` 与 `platform/challenges/imports/:importId` 显式下传 props

### feature model 本轮负责

- 保留数据加载、表单提交、上传与导入 workflow
- 不再直接 import `vue-router`
- 暴露 route target contract 或 redirect signal

### route view / feature UI 本轮负责

- 直接消费 route target
- 通过 `AppRouteRedirect` 承接注册成功、导入成功这类跳转

## 任务切片

### Slice 1：route props owner

- 目标：
  - `ScoreboardDetail`、`ChallengeImportPreview` route param 下沉为 props
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/scoreboard/__tests__/ScoreboardView.test.ts src/views/platform/__tests__/ChallengeImportPreview.test.ts`

### Slice 2：薄导航 route target owner

- 目标：
  - `AWDChallengeLibrary`、`ChallengeImportManage`、`ChallengeImportPreview` 的薄导航改成 route target
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/AWDChallengeLibrary.test.ts src/views/platform/__tests__/ChallengeImportManage.test.ts src/views/platform/__tests__/ChallengeImportPreview.test.ts`

### Slice 3：success redirect transport

- 目标：
  - 注册成功、导入成功跳转改成 `AppRouteRedirect`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/auth/__tests__/RegisterView.test.ts src/views/platform/__tests__/ChallengeImportPreview.test.ts`

### Slice 4：allowlist / backlog / review 收尾

- 目标：
  - 更新 allowlist、backlog 和 review
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/scoreboard/__tests__/ScoreboardView.test.ts src/views/platform/__tests__/AWDChallengeLibrary.test.ts src/views/auth/__tests__/RegisterView.test.ts src/views/platform/__tests__/ChallengeImportManage.test.ts src/views/platform/__tests__/ChallengeImportPreview.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
  - `cd code/frontend && npm run typecheck`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/scoreboard/__tests__/ScoreboardView.test.ts src/views/platform/__tests__/AWDChallengeLibrary.test.ts src/views/auth/__tests__/RegisterView.test.ts src/views/platform/__tests__/ChallengeImportManage.test.ts src/views/platform/__tests__/ChallengeImportPreview.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/feature-router-first-batch-cleanup.md docs/plan/impl-plan/2026-05-29-feature-router-first-batch-cleanup-plan.md docs/reviews/frontend/2026-05-29-feature-router-first-batch-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/router/routes/authRoutes.ts code/frontend/src/router/routes/studentRoutes.ts code/frontend/src/router/routes/platformRoutes.ts code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts code/frontend/src/features/platform-awd-challenges/model/useAwdChallengeLibraryPage.ts code/frontend/src/features/auth/model/useRegisterPage.ts code/frontend/src/features/challenge-package-import/model/index.ts code/frontend/src/features/challenge-package-import/model/useChallengeImportManagePage.ts code/frontend/src/features/challenge-package-import/model/useChallengeImportPreviewPage.ts code/frontend/src/views/scoreboard/ScoreboardDetail.vue code/frontend/src/views/platform/AWDChallengeLibrary.vue code/frontend/src/views/auth/RegisterView.vue code/frontend/src/views/platform/ChallengeImportManage.vue code/frontend/src/views/platform/ChallengeImportPreview.vue code/frontend/src/components/platform/challenge/ChallengeImportHeroPanel.vue code/frontend/src/components/platform/challenge/ChallengeImportQueuePanel.vue code/frontend/src/components/platform/challenge/ChallengeImportPreviewWorkspacePanel.vue code/frontend/src/features/platform-awd-challenges/ui/AWDChallengeLibraryPage.vue code/frontend/src/views/scoreboard/__tests__/ScoreboardView.test.ts code/frontend/src/views/platform/__tests__/AWDChallengeLibrary.test.ts code/frontend/src/views/auth/__tests__/RegisterView.test.ts code/frontend/src/views/platform/__tests__/ChallengeImportManage.test.ts code/frontend/src/views/platform/__tests__/ChallengeImportPreview.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 第二批和第三批 allowlist 不在本轮范围。
- 注册成功和导入成功跳转改为 redirect transport 后，要确认测试里的时序断言没有依赖旧的同步 `push()`。
