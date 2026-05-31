# 前端页面数据流与 owner

> 状态：Current
> 事实源：`code/frontend/src/pages/`、`code/frontend/src/features/**/model/`、`code/frontend/src/stores/`、`code/frontend/src/__tests__/`
> 替代：无

## 定位

本文档只说明“页面数据流现在由谁负责”，以及几个典型页面的调用链和状态边界。

- 覆盖：route view、feature model、局部 composable、共享 store 之间的 owner 切分。
- 不覆盖：所有页面的完整 UI 清单；这里只保留当前最稳定、最能代表结构的页面链路。

## 当前设计

- `code/frontend/src/pages/**`
  - 负责：作为运行时 route entry，挂载页面结构、组合 feature / widget 和共享组件
  - 不负责：直接调用业务 API，或长期持有复杂路由 query 与异步编排

- `code/frontend/src/__tests__`
  - 负责：全局前端架构守卫与跨页面稳定策略测试
  - 不负责：替代 route page、feature 或 widget 的行为测试 owner

- `code/frontend/src/features/**/model`
  - 负责：页面级请求编排、路由参数解析、分页、导出、实时桥接和局部状态机
  - 不负责：重复实现共享登录态、通知列表或竞赛公共状态

- `code/frontend/src/features/**/ui`
  - 负责：承接单一 feature 的 page-sized surface、editor、workspace 壳和目录面板
  - 不负责：绕过 feature model 直接请求 API，或继续承载 `*RoutePage.vue` 运行时入口

- `code/frontend/src/widgets/**`
  - 负责：跨 feature 页面区块组合，例如 workspace、archive、review 这类可被页面入口直接挂载的组合块
  - 不负责：继续承载 `*RoutePage.vue` 运行时入口

- `code/frontend/src/stores/auth.ts`、`notification.ts`、`contest.ts`
  - 负责：跨页共享状态
  - 不负责：替代页面 owner 本身

## 1. 页面 owner 总览

| 页面/能力 | 主要 owner | 关键入口 |
| --- | --- | --- |
| 学生仪表盘 | `features/student-dashboard/model` | `useStudentDashboardPage.ts` |
| 题目详情 | `features/challenge-detail/model` | `useChallengeDetailPage.ts` |
| 通知列表与抽屉 | `features/notifications/model` | `useNotificationListPage.ts`、`useNotificationDrawer.ts` |
| 排行榜目录与详情 | `features/scoreboard/model` | `useScoreboardView.ts`、`useScoreboardDetailPage.ts` |
| 竞赛详情 | `features/contest-detail/model` | `useContestDetailPage.ts`、`useContestDetailRoutePage.ts` |
| 学员分析复盘工作流 | `features/student-analysis-workspace/model`、`features/student-analysis-review/model` | `useStudentAnalysisPage.ts` |
| 题目管理与导入 | `features/platform/challenges/model`、`features/platform/challenge-package-import/model` | `useChallengeManagePage.ts`、`useChallengeImportManagePage.ts`、`useChallengeImportPreviewPage.ts` |
| 平台题解管理 | `features/platform/challenges/model`、`features/challenge-writeup-editor/model`、`features/challenge-writeup-editor/ui` | `useChallengeWriteupPage.ts`、`useChallengeWriteupViewPage.ts`、题解 feature UI |

## 2. 典型数据流

### 2.1 学生仪表盘

`useStudentDashboardPage()` 当前负责两件事：

1. 用 `useRouteQueryTabs()` 管理 `overview / recommendation / category / timeline / difficulty`
2. 把数据加载委托给 `useStudentDashboardData()`，把展示绑定委托给 `useStudentDashboardPanelBindings()`

`useStudentDashboardData()` 当前流程：

1. 先看 `authStore.user.role`
2. 若是 `teacher`，跳 `TeacherDashboard`
3. 若是 `admin`，跳 `PlatformOverview`
4. 学生角色下并行加载：
  - `getMyProgress()`
  - `getMyTimeline()`
  - `getRecommendations()`
  - `getSkillProfile()`
5. 只把仪表盘局部数据保留在 feature model，不写全局 store

### 2.2 题目详情

`useChallengeDetailPage()` 当前是一个“组合器”，而不是把所有逻辑直接写在一个文件里。

主链路：

1. 从 `route.params.id` 取题目 ID
2. route 变化时：
  - 清空当前题目与题解状态
  - 重置交互状态
  - 切回 `question` tab
3. 并行执行：
  - `loadChallenge()`
  - `loadMyWriteupSubmission()`
  - `loadSubmissionRecords()`
4. `useChallengeDetailDataLoader()` 负责题目详情和题解加载
5. `useChallengeInstance()` 负责实例创建、打开、延时、销毁和排队轮询
6. `useChallengeDetailInteractions()` 负责 Flag 提交、附件下载、题解保存和提交记录

关键边界：

- 只有题目已解出时，才加载推荐题解和社区题解
- 实例状态轮询只在 `pending / creating` 时继续
- 页面 tab 与 query 同步走 `useUrlSyncedTabs()`

### 2.3 通知页与通知抽屉

列表页 `useNotificationListPage()`：

1. 通过 `usePagination()` 拉通知列表
2. 页码为 `1` 时，把列表同步进 `notificationStore`
3. 支持“当前页全部标已读”
4. admin 角色额外允许打开通知发布抽屉

抽屉 `useNotificationDrawer()`：

1. 直接消费 `notificationStore.notifications` 和 `notificationStore.unreadCount`
2. 按 `realtimeStatus()` 显示“在线 / 同步中 / 连接异常 / 手动查看”
3. 在抽屉内支持全部标已读和跳详情页

这条链路说明当前通知系统是：

- HTTP 列表页负责完整数据
- store 负责跨页共享
- realtime 负责增量同步

### 2.4 排行榜目录与详情

目录页由 `useScoreboardView()` 和 `useScoreboardContestDirectoryPage()` 配合：

1. `useScoreboardView()` 加载竞赛列表和练习积分榜
2. 只保留 `running / frozen / ended` 的竞赛
3. `useScoreboardContestDirectoryPage()` 负责分页、卡片说明和目录态文案

详情页由 `useScoreboardDetailPage()` 负责：

1. 从路由中拿 `contestId`
2. 拉 `getScoreboard(contestId, { page: 1, page_size: 100 })`
3. 根据竞赛状态计算 `supportsRealtime`
4. 实时桥接由 `ScoreboardRealtimeBridge.vue` 挂载后触发 `loadScoreboard(true)`

### 2.5 教师学生分析

`useTeacherStudentAnalysisPage()` 当前承担的是教师复盘工作区的总组合，但内部已经拆子能力：

- 基础加载：
  - `getClasses()`
  - `getClassStudents()`
  - `getStudentProgress()`
  - `getStudentSkillProfile()`
  - `getStudentRecommendations()`
  - `getStudentTimeline()`
- 复盘工作区：
  - `useTeacherReviewWorkspace()`
- 题解审核和人工评审：
  - `useTeacherSubmissionReviewFlows()`
- 复盘归档导出：
  - `useReviewArchiveExportFlow()`
- 面包屑和导航：
  - `useBackofficeBreadcrumbDetail()`
  - `useTeacherStudentAnalysisNavigation()`

说明：

- 路由 `className / studentId / reviewMode / reviewResult / reviewChallengeId` 都会反驱到 feature state
- 页面本身是“组合器”，不再让所有评审逻辑堆在同一个函数里

### 2.6 平台题目管理与导入

题目目录 `useChallengeManagePage()`：

1. 组合 `usePlatformChallenges()` 拉目录、筛选、发布、删除
2. 本地只负责排序、空状态和打开“导入题目”工作区

题目导入 `useChallengeImportManagePage()` / `useChallengePackageImport()`：

1. 进入导入页先 `refreshQueue()`
2. 选择一个或多个包后，通过 `challengeImportUploadFlow.ts` 走预解析
3. 成功时跳 `PlatformChallengeImportPreview`
4. 预览页根据 `route.params.importId` 调 `loadPreview()`
5. 提交导入时走 `commitPreview()`，成功后回题目目录

关键边界：

- 上传流程、错误归一化和队列刷新已经拆到 `challengeImportUploadFlow.ts`、`challengeImportErrorSupport.ts`
- 主组合器 `useChallengePackageImport()` 不再内联上传细节

### 2.7 平台题解管理

题解 route page 当前分成两层 owner：

1. `features/platform/challenges/model/useChallengeWriteupPage.ts` 与 `useChallengeWriteupViewPage.ts`
   - 负责从 `route.params.id` 解析题目 ID
   - 负责“返回题目详情”“跳转题解编辑页”这类 route navigation
2. `features/challenge-writeup-editor/model/*`
   - 负责题解加载、保存、删除、推荐切换和投稿目录查询
3. `features/challenge-writeup-editor/ui/*`
   - 负责题解管理目录、编辑器和查看页的 page-sized surface
   - 只消费题解 feature model，不直接依赖 route 或 API

这条链路说明：

- route entry 应继续保持薄壳，并统一落到 `pages/**`
- feature model 继续持有真实行为 owner
- page-sized 业务 UI 不一定非要挂在 `components/*Page.vue`，只要它只服务单一 feature，就应跟随 feature 收进 `features/*/ui`

### 2.8 个人安全设置页的 shell 分层

`SecuritySettings` 当前保留了一个典型的 route page -> feature model -> workspace shell 链路：

1. `code/frontend/src/pages/profile/SecuritySettingsRoutePage.vue`
   - 负责：作为路由入口接线 `useSecuritySettingsPage()` 和 `SecuritySettingsWorkspaceShell.vue`
   - 不负责：内联密码校验、提交流程和展示布局细节
2. `code/frontend/src/features/profile/model/useSecuritySettingsPage.ts`
   - 负责：密码表单草稿、字段校验、提交防重、错误提示和成功 toast
   - 不负责：页面布局、分区标题和视觉样式
3. `code/frontend/src/features/profile/ui/SecuritySettingsWorkspaceShell.vue`
   - 负责：安全概况、密码修改区和安全提示区的页面壳布局
   - 不负责：直接调用 `changePassword()` 或自行管理提交流程

辅助图：

- 渲染图：[`shell方案.png`](./shell方案.png)
- Mermaid 源：[`07-pages-dataflow-security-settings-shell.mmd`](./07-pages-dataflow-security-settings-shell.mmd)

## 3. 页面数据流不变量

- route page 默认不直接 import 非 contract API 模块
- 页面 owner 优先以 feature model 组合多个子 composable，而不是新增一个更大的页面文件
- query/tab 同步继续通过 `useRouteQueryTabs()` 或 `useUrlSyncedTabs()` 收口
- 只有跨页面共享的数据进 store；其余数据跟随页面生命周期释放

## 4. Guardrail

- route page 与页面路由边界：`code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
- 前端整体分层：`code/frontend/src/__tests__/architectureBoundaries.test.ts`
- 学生仪表盘组合边界：`code/frontend/src/features/student-dashboard/model/useStudentDashboardPageBoundary.test.ts`
- 竞赛详情组合边界：`code/frontend/src/features/contest-detail/model/useContestDetailPageBoundary.test.ts`
- 题目导入组合边界：`code/frontend/src/features/platform/challenge-package-import/model/useChallengePackageImportBoundary.test.ts`
