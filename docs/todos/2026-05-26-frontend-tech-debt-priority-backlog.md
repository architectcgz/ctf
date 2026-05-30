# 前端技术债优先级清单

- Project: `ctf 仓库根目录`
- Created: `2026-05-26T23:09+08:00`
- Status: `Open`

## Context

基于当前前端事实源和代码现状整理：

- `docs/reviews/frontend/README.md`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/__tests__/frontendArchitecturePolicy.ts`
- 当前组件体量扫描结果

排序原则：

- 先做收益高、风险相对可控、能直接复用最近 page shell 拆分模式的项
- 再做收益高但跨模块耦合更重的结构债
- 性能监控、i18n 这类前置决策未定的项保留在后面单独跟踪

## Components 存量清理清单

当前前端已经不是“中间迁移态”，但仍保留大量历史业务组件目录。后续如果要继续朝更纯的 `views -> widgets -> features -> entities -> common` 形态收口，优先按下面三类推进。

### A. 保留在 `components/*`

- `components/common`
  - 原因：共享原语、模板、表格、空状态、菜单等跨业务复用块。
- `components/layout`
  - 原因：全局布局壳；`AppLayout`、`Sidebar`、`TopNav`、`NotificationDrawer` 当前已收口为稳定 layout owner。
- `components/navigation`
  - 原因：`AppRouteLink`、`AppRouteRedirect`、`routeTarget.ts` 属于横切导航原语。
- `components/errors`
  - 原因：错误状态页壳体是全局展示原语，不绑定单一 feature owner。
- `components/charts`
  - 原因：图表展示块偏共享 primitive，而不是单一业务 workflow。
- `components/reports`
  - 原因：当前主要承接 teacher / platform 共用的中立导出对话框入口，仍属于跨 feature 复用展示块。

### B. 优先继续迁入 `features/*/ui`

- `components/platform`
  - 现状：仍是最大存量目录，约 `63` 个文件；大量 hero / panel / modal 明显只服务单一 platform feature。
  - 代表：`platform/challenge/*`、`platform/class/*`、`platform/instance/*`、`platform/student/*`、`platform/user/*`、`platform/audit/*`
- `components/teacher`
  - 现状：约 `34` 个文件；仍有不少 teacher 目录 / 洞察 / dashboard 展示块属于单一 feature family。
  - 代表：`teacher/class-management/*`、`teacher/dashboard/*`、`teacher/student-insight/*`、`teacher/instance-management/*`
- `components/contests`
  - 现状：大量竞赛详情 / AWD workspace 展示块仍留在历史组件目录。
  - 代表：`ContestOverviewPanel.vue`、`ContestTeamWorkspaceSection.vue`、`contests/awd/*`
- `components/challenge`
  - 现状：挑战目录 / 题面 / 实例 / 题解等块仍有较多单一 challenge capability UI。
  - 代表：`ChallengeDirectoryPanel.vue`、`ChallengeQuestionPanel.vue`、`ChallengeWriteupPanel.vue`
- `components/dashboard`
  - 现状：student dashboard 相关变体和样式编辑块仍留在旧目录。
  - 代表：`dashboard/student/*`
- `components/profile`
  - 现状：`UserProfileWorkspaceShell.vue`、`SkillProfileWorkspaceShell.vue`、`SecuritySettingsWorkspaceShell.vue` 已基本只服务 `features/profile`
- `components/auth`
  - 现状：`AuthEntryShell.vue` 只服务 auth route，长期更适合并入 `features/auth/ui`
- `components/notifications`
  - 现状：`NotificationCategoryFilter.vue` 目前只服务通知能力，本质更像 feature-owned UI
- `components/instance`
  - 现状：`InstanceListWorkspaceShell.vue` 只服务实例目录能力，更适合并回 `features/instance-list/ui`
- `components/scoreboard`
  - 现状：`ScoreboardWorkspaceShell.vue` 只服务 scoreboard capability，更适合并回 `features/scoreboard/ui`
- `components/training`
  - 现状：训练时间线 panel 已经是中立共享展示，但如果后续 consumer 继续局限在 student dashboard / teacher insight，仍可评估并回 owning feature

### C. 候选收进 `widgets/*`

- `components/awd-review`
  - 现状：当前只剩 `index.ts` 中立入口，真正的 route-level 组合已经在 `widgets/awd-review-workspace`
  - 后续判断：如果确认只服务 AWD review workspace，适合并入 `widgets/awd-review-workspace`
- `components/review-archive`
  - 现状：当前只剩 `index.ts` 中立入口，真正的 route-level 组合已经在 `widgets/review-archive-workspace`
  - 后续判断：如果确认只服务 review archive workspace，适合并入 `widgets/review-archive-workspace`

### D. 优先确认是否可删除的残片

- `components/class-management`
  - 现状：当前目录下没有活动源码文件，也没有 consumer
  - 后续动作：优先确认是否已完全退出主路径；若无残余引用，直接从 backlog 进入删除清理而不是继续视作迁移目标

### 执行顺序建议

1. 先按目录批量清 `components/platform`、`components/teacher`、`components/contests` 里只服务单一 feature 的 UI。
2. 再收 `components/profile`、`components/auth`、`components/instance`、`components/scoreboard` 这类“单 capability workspace shell”。
3. 最后评估 `components/awd-review`、`components/review-archive` 是否并入对应 `widgets/*`，以及 `components/class-management` 是否直接删除。

### 第 1 批子目录迁移表

以下清单只覆盖当前收益最高的三组历史业务目录：`components/platform`、`components/teacher`、`components/contests`。

`2026-05-30` 进展：

- 已完成 `platform/dashboard/*`、`platform/user/*`、`platform/class/*`、`platform/student/*`、`platform/instance/*`、`platform/audit/*`、`platform/images/*` 迁移。
- 对应 UI 已收口到：
  - `features/platform/overview/ui`
  - `features/platform/user-management/ui`
  - `features/platform/class-management/ui`
  - `features/platform/student-management/ui`
  - `features/platform/instance-management/ui`
  - `features/audit-log/ui`
  - `features/image-management/ui`
- route view 已改为通过 feature public API 组合组件，不再直接引用这些历史 `components/platform/*` 目录。
- 相关页面回归测试与前端架构边界测试已通过；下一批入口顺延到 `components/teacher`、`components/contests` 与 `platform/challenge|awd-service|awd-review`。

#### `components/platform`

- `platform/dashboard/*`
  - 当前 consumer：`features/platform/overview/ui/PlatformOverviewPage.vue`
  - 目标 owner：`features/platform/overview/ui`
  - 优先级：`P1`
- `platform/user/*`
  - 当前 consumer：`features/platform/user-management/ui/UserGovernancePage.vue`
  - 目标 owner：`features/platform/user-management/ui`
  - 优先级：`P1`
- `platform/class/*`
  - 当前 consumer：`pages/platform/ClassManageRoutePage.vue`
  - 目标 owner：`features/platform/class-management/ui`
  - 优先级：`P1`
- `platform/student/*`
  - 当前 consumer：`pages/platform/StudentManageRoutePage.vue`
  - 目标 owner：`features/platform/student-management/ui`
  - 优先级：`P1`
- `platform/instance/*`
  - 当前 consumer：`pages/platform/InstanceManageRoutePage.vue`
  - 目标 owner：`features/platform/instance-management/ui`
  - 优先级：`P1`
- `platform/audit/*`
  - 当前 consumer：`pages/platform/AuditLogRoutePage.vue`
  - 目标 owner：`features/audit-log/ui`
  - 优先级：`P1`
- `platform/images/*`
  - 当前 consumer：`pages/platform/ImageManageRoutePage.vue`
  - 目标 owner：`features/image-management/ui`
  - 优先级：`P1`
- `platform/awd-review/*`
  - 当前 consumer：`pages/awd-review/PlatformAwdReviewIndexRoutePage.vue`、`pages/awd-review/PlatformAwdReviewDetailRoutePage.vue`
  - 目标 owner：`widgets/awd-review-workspace`、`features/awd-review-detail-workspace`
  - 优先级：`P1`
- `platform/awd-service/*`
  - 当前 consumer：`features/platform/awd-challenges/ui/AWDChallengeLibraryPage.vue`
  - 目标 owner：`features/platform/awd-challenges/ui`
  - 优先级：`P1`
- `platform/challenge/*`
  - 当前 consumer：`pages/platform/challenges/ChallengeImportManageRoutePage.vue`、`pages/platform/challenges/ChallengeImportPreviewRoutePage.vue`、`features/platform/challenges/ui/ChallengeManagePage.vue`、`pages/platform/challenges/ChallengeDetailRoutePage.vue`、`pages/platform/challenges/ChallengeWriteupViewRoutePage.vue`
  - 目标 owner：按能力分拆到 `features/challenge-package-import/ui`、`features/platform/challenges/ui`、`features/platform/challenge-detail/ui`，避免继续留在一个大目录
  - 优先级：`P1`
- `platform/cheat/*`
  - 当前 consumer：`pages/platform/CheatDetectionRoutePage.vue`
  - 目标 owner：当前先并入 `features/platform/overview/ui`；若后续 capability 继续增长，再独立成 `platform-cheat-detection`
  - 优先级：`P2`
- `platform/contest/*`
  - 当前 consumer：主要是测试、以及少量 platform contest route shell 残片
  - 目标 owner：按职责分到 `features/platform/contests/ui`、`features/contest-awd-config/ui`
  - 优先级：`P2`
- `PlatformPaginationControls.vue`
  - 当前形态：平台后台目录分页壳
  - 目标 owner：先判定是否能中性化后沉到 `components/common`；若仍明显平台专属，再并入 `features/platform-*` 本地目录
  - 优先级：`P2`

#### `components/teacher`

- `teacher/dashboard/*`
  - 当前 consumer：`features/teacher-dashboard/ui/TeacherDashboardPage.vue`
  - 目标 owner：`features/teacher-dashboard/ui`
  - 优先级：`P1`
- `teacher/instance-management/*`
  - 当前 consumer：`features/teacher-instances/ui/TeacherInstanceManagementPage.vue`
  - 目标 owner：`features/teacher-instances/ui`
  - 优先级：`P1`
- `teacher/class-management/*`
  - 当前 consumer：`features/class-students-workspace/ui/ClassStudentsPage.vue`、`features/student-analysis-workspace/ui/StudentAnalysisPage.vue`、`features/teacher-class-management/ui/ClassManagementPage.vue`
  - 目标 owner：拆到 `features/class-students-workspace/ui`、`features/student-analysis-workspace/ui`、`features/teacher-class-management/ui`
  - 优先级：`P1`
- `teacher/student-insight/*`
  - 当前 consumer：teacher 学员洞察 / student analysis 相关 route 和 feature
  - 目标 owner：`features/student-analysis-workspace/ui`
  - 优先级：`P1`
- `teacher/ClassInsightsPanel.vue`、`ClassReviewPanel.vue`、`ClassTrendPanel.vue`
  - 当前 consumer：`features/class-students-workspace/ui/ClassStudentsPage.vue`、`features/teacher-class-report-export/ui/ClassReportExportPreviewSection.vue`
  - 目标 owner：优先中性化后并入 `features/class-students-workspace/ui` 或提炼为实体/共享展示块；不要继续顶层挂在 `components/teacher`
  - 优先级：`P1`
- `teacher/StudentInsightPanel.vue`
  - 当前 consumer：`features/student-analysis-workspace/ui/StudentAnalysisPage.vue`
  - 目标 owner：`features/student-analysis-workspace/ui`
  - 优先级：`P1`
- `teacher/awd-review/*`
  - 当前 consumer：主要由 AWD review workspace 和测试消费
  - 目标 owner：优先并入 `widgets/awd-review-workspace` 或 `features/awd-review-workspace/ui`，避免继续挂在 teacher 目录
  - 优先级：`P2`
- `teacher/review-archive/*`
  - 当前 consumer：主要由 review archive workspace 和测试消费
  - 目标 owner：优先并入 `widgets/review-archive-workspace` 或中立 `components/review-archive`
  - 优先级：`P2`

#### `components/contests`

- `ContestOverviewPanel.vue`
  - 当前 consumer：`views/contests/ContestDetail.vue`
  - 目标 owner：`features/contest-detail/ui`
  - 优先级：`P1`
- `ContestChallengeWorkspacePanel.vue`
  - 当前 consumer：`views/contests/ContestDetail.vue`
  - 目标 owner：`features/contest-detail/ui`
  - 优先级：`P1`
- `ContestAnnouncementsWorkspaceSection.vue`、`ContestAnnouncementsPanel.vue`
  - 当前 consumer：`views/contests/ContestDetail.vue`
  - 目标 owner：优先并入 `features/contest-detail/ui`；若后续 student 侧公告能力继续扩张，再拆单独 feature
  - 优先级：`P1`
- `ContestTeamWorkspaceSection.vue`、`ContestTeamPanel.vue`、`ContestTeamDialogs.vue`
  - 当前 consumer：`views/contests/ContestDetail.vue`
  - 目标 owner：`features/contest-detail/ui`
  - 优先级：`P1`
- `contests/awd/*`
  - 当前 consumer：`features/contest-awd-workspace/ui/ContestAWDWorkspacePanel.vue`
  - 目标 owner：`features/contest-awd-workspace/ui`
  - 优先级：`P1`

#### 推荐切片顺序

1. `platform/dashboard + user + class + student + instance + audit + images`
2. `teacher/dashboard + instance-management + class-management + student-insight + 顶层 teacher panels`
3. `contests/*` 与 `contests/awd/*`
4. `platform/challenge + platform/awd-service + platform/awd-review`
5. `teacher/awd-review + teacher/review-archive + platform/cheat + platform/contest`

## Open Items

- [x] P1：清空运行时 `views/*.vue` 入口层
  - 依据：当前前端不应再用 `views/*.vue` 承载运行时页面入口；同时也不再让 `features/*RoutePage.vue` 或 `widgets/*RoutePage.vue` 兼任页面层语义。
  - 目标：router 运行时入口统一收敛到 `pages/**`；`features/**` 回到能力 owner，`widgets/**` 回到页面区块组合，`views/` 仅保留 `__tests__/` 与必要测试支撑文件。
  - `2026-05-30` 进展：`ChallengeManage`、`PlatformOverview`、`TeacherDashboard`、`ChallengeTopologyStudio`、`ChallengeWriteup`、`ChallengeWriteupView`、`AWDChallengeImport`、`ChallengeImportPreview`、`PlatformStudentReviewArchive`、`AWDReviewIndex`、`TeacherAWDReviewIndex`、`ImageManage`、`CheatDetection`、`PlatformAwdReviewDetail`、`TeacherAWDReviewDetail`、`ChallengePackageFormat`、`ChallengeImportManage` 已切到 `pages/**` 运行时入口；对应 `views/*.vue` 当前只保留测试桥接壳。
  - `2026-05-30` 学生侧进一步进展：`dashboard`、`challenges`、`contests`、`scoreboard`、`instances`、`profile`、`notifications` 已全部切到 `pages/**`，对应学生旧 `views/*.vue` 已删除；当前 `views/` 在这组能力下只剩 `__tests__/`。
  - `2026-05-30` 最终进展：`auth`、`errors`、`UILab` 也已统一迁入 `pages/**`，对应旧 `views/*.vue` 已删除。
  - `2026-05-30` platform 清场进展：`views/platform/*.vue` 旧页面残片已删除，相关 raw-source 测试也已切到 `pages/**` 或 feature / widget owner。
  - `2026-05-30` teacher 清场进展：`views/teacher/*.vue` 旧页面残片也已删除，教师侧相关测试入口已切到 `pages/teacher/**` 或 `pages/awd-review/**`。
  - 当前结论：运行时路由入口已经完全退出 `views/`；`views/` 现在只保留 `__tests__/` 邻近测试支撑。

- [x] P1：继续拆 `PlatformOverviewPage.vue` 与 `TeacherInstanceManagementPage.vue`
  - 依据：`PlatformOverviewPage.vue` 当前约 `728` 行，`TeacherInstanceManagementPage.vue` 当前约 `597` 行，仍属于明显过宽的 legacy component page。
  - 收益：最容易复用最近 `TeacherDashboardPage` / `ClassStudentsPage` / `UserGovernancePage` 的 page owner 收口模式，能继续压缩 `legacyComponentPageAllowlist` 的负担。
  - 风险：中低。主要是展示区块切分和源码护栏适配，业务 owner 相对清楚。
  - `2026-05-26` 进展：已拆为 `PlatformOverviewHeroPanel` / `PlatformOverviewAlertsSection` / `PlatformOverviewHotspotsSection` 与 `TeacherInstanceHeroPanel` / `TeacherInstanceDirectorySection`，父页继续只保留 page shell、事件桥接和派发展示数据。

- [x] P1：处理 `platform-users` 过宽职责，把用户治理、班级/学生目录、实例管理继续拆回更清晰的 feature owner
  - 依据：原始架构 review 把 `platform-users` 视为过宽 bucket；当前代码事实里，这个桶已经拆成 `platform/user-management`、`platform/class-management`、`platform/student-management`、`platform/instance-management`。
  - 收益：完成后，admin 侧的用户治理、班级目录、学生目录、实例目录已经不再堆在同一个 feature 名下。
  - 风险：原始拆桶风险已消化；当前残留风险已转移到下面这条“admin / teacher 结构耦合”。
  - `2026-05-27` 进展：进一步补上 `api/admin/teaching.ts` owner，platform class / student / instance feature 与 admin 通知发布不再直接引用 `@/api/teaching`。

- [x] P1：收口 admin / teacher 结构耦合，优先停止让 `/platform/*` 直接依赖 teacher 视图或 teacher 语义 owner
  - 依据：`docs/reviews/architecture/2026-05-24-frontend-architecture-review.md` 仍把这条列为当前 P1 finding。
  - 收益：能减少权限面和页面 owner 的隐式耦合，避免 teacher 改动静默影响 platform。
  - 风险：高。这里不是简单拆模板，涉及 route view、共享 workspace feature 的重新归位。
  - `2026-05-27` 进展：已先收口 API owner，platform class / student / instance feature 和 admin 通知发布改为通过 `api/admin/teaching.ts` 取教学目录能力；后续仍需继续处理共享页面与 contract 命名里的 teacher 语义。
  - `2026-05-27` 进一步进展：`class-insight-window`、`student-analysis-review`、`student-review-archive` 与 `review-archive-workspace` 已补中立 public owner，`class-students-workspace`、`student-analysis-workspace`、`student-review-archive-workspace` 以及 teacher / platform 的复盘归档 route view 已切到中立入口；剩余重点收口面转到 `PlatformClassWorkspaceSection`、`PlatformAwdReviewDetail` 和更深层 contract 命名。
  - `2026-05-27` AWD 进展：`awd-review-workspace` 与 `awd-review-detail-workspace` 已按角色切换到 `api/admin` / `api/teacher` owner，`PlatformAwdReviewIndex`、`PlatformAwdReviewDetail` 不再通过共享 feature 间接依赖 `@/api/teaching` 的 teacher 命名函数；当前剩余重点收口面收敛到 `PlatformClassWorkspaceSection`、`ChallengeWriteupManagePanel` 和更深层 contract 命名。
  - `2026-05-27` class redirect 进展：`class-workspace-redirect` 已改成只解析 alias route 对应的 `panel`，最终 canonical target route 由 `PlatformClassWorkspaceSection` / `TeacherClassWorkspaceSection` 显式传入；`PlatformClassWorkspaceSection` 的 redirect owner 命名残留已收口，当前剩余重点收口面进一步收敛到 `ChallengeWriteupManagePanel` 和更深层 contract 命名。
  - `2026-05-27` writeup 进展：`ChallengeWriteupManagePanel` 对应的 `useChallengeWriteupManagement` 已切到 `api/admin/authoring.ts` 下的 platform writeup submissions owner；教师侧 `TeacherStudentAnalysis` / `useSubmissionReviewFlows` 的题解查看与评阅链路保持不变，当前剩余重点收口面进一步收敛到更深层 contract 命名。
  - `2026-05-27` contract naming 进展：题解投稿目录的共享 DTO 已从 `TeacherSubmissionWriteupItemData` 收口到 `WriteupSubmissionItemData`，manual review 共享 DTO 已进一步从 `TeacherManualReviewSubmissionItemData` / `TeacherManualReviewSubmissionDetailData` 收口到 `ManualReviewSubmissionItemData` / `ManualReviewSubmissionDetailData`，writeup detail DTO 已从 `TeacherSubmissionWriteupDetailData` 收口到 `WriteupSubmissionDetailData`，班级目录 DTO 已从 `TeacherClassItem` 收口到 `ClassDirectoryItem`，攻击会话筛选 query 已从 `TeacherAttackSessionQuery` 收口到 `AttackSessionQuery`，AWD review 赛事目录项 DTO 已从 `TeacherAWDReviewContestItemData` 收口到 `AwdReviewContestItemData`；platform 题解管理、teacher / platform 学员分析以及共享班级目录 / 复盘筛选 / AWD review index 消费面不再继续共用这组 teacher 前缀 contract，当前更深层命名残片已从这组共享 contract 面清掉。
  - `2026-05-27` report dialog owner 进展：`ClassReportExportDialog` 已补中立 `components/reports` public owner，teacher / platform 的班级管理、班级学员页和学员分析页不再直接从 `components/teacher/reports` 引共享导出对话框；当前剩余重点收口面回到 review archive / AWD review 这类 shared widget 仍直连 teacher 组件入口的存量。
  - `2026-05-27` review archive owner 进展：`ReviewArchiveWorkspace` 已改从中立 `components/review-archive` 入口读取 hero / observation / evidence / reflection 面板，对 `components/teacher/review-archive/*` 的四条直连 import 收敛为一条 barrel；当前剩余重点收口面进一步缩到 AWD review shared widget 的 teacher 组件入口。
  - `2026-05-27` AWD review owner 进展：`AwdReviewWorkspace` 已改从中立 `components/awd-review` 入口读取 round selector / analysis / evidence / team drawer，对 `components/teacher/awd-review/*` 的四条直连 import 完成收口；当前这条 P1 剩余重点开始回到更大颗粒度的页面 / 壳体拆分。
  - `2026-05-28` AWD review access owner 进展：已新增 `api/awd-reviews.ts` 作为 role-aware facade，`useAwdReviewIndex.ts`、`useAwdReviewExportFlow.ts`、`useAwdReviewDetailPage.ts` 不再各自直接双引 `@/api/admin` 与 `@/api/teacher`；当前 AWD review 这条 admin / teacher 耦合残余重点已从共享 feature model 的 access owner 漂移，进一步收敛到更深层的 DTO / contract naming 残片。
  - `2026-05-28` AWD review shared contract naming 进展：`AwdReviewArchiveData`、`AwdReviewRoundItemData`、`AwdReviewTeamItemData`、`AwdReviewServiceItemData`、`AwdReviewAttackItemData`、`AwdReviewTrafficItemData`、`AwdReviewSelectedRoundData` 与 `AwdReviewContestPageData` 已完成中性化，`AwdReviewWorkspace` / `AwdReviewIndexWorkspace` 的 `AwdReview*` summary types、builder 和 `AWD_REVIEW_*` copy owner 也已同步收口。当前 AWD review 这条线上残留的 teacher 语义，已主要收敛到 route name 和 teacher endpoint function / admin alias 这层显式 transport owner。
  - `2026-05-28` AWD review API implementation owner 进展：`api/teaching/awd-reviews.ts` 的共享实现函数已从 `listTeacherAWDReviews()` / `getTeacherAWDReview()` 这类 teacher 命名收口到中性 `AwdReview*` owner，`api/admin/contests.ts` 也改成显式 `listPlatformAWDReviews()` / `getPlatformAWDReview()` 包装，不再 alias teacher 命名函数。当前 AWD review 这条线前端本地残留的 teacher 语义，已进一步收敛到后端 `/api/v1/teacher/awd/reviews*` 路径和 teacher route name。
  - `2026-05-28` instance directory contract naming 进展：实例目录共享 DTO / page data / summary / status filter 已从 `TeacherInstance*` 收口到 `InstanceDirectory*`，teacher / platform 实例目录消费面不再继续共用这组 teacher 前缀 contract。当前这条 P1 在实例目录 contract 层的 teacher 语义，已进一步收敛到 `getTeacherInstances()` / `destroyTeacherInstance()` 这类 public wrapper 名和后端 `/api/v1/teacher/instances` transport path。
  - `2026-05-28` instance role-aware access owner 进展：已新增 `api/instances.ts` 作为实例目录共享 workflow 的中立 facade，`usePlatformInstanceManagementPage.ts` 与 `useInstances.ts` 不再各自直连 `@/api/admin` / `@/api/teacher` 的实例目录函数。当前这条 P1 在实例 access owner 层的 teacher 语义，已进一步收敛到 public wrapper 名和 transport path，不再停留在 shared feature model 的 API 选择面。
  - `2026-05-28` instance API implementation owner 进展：`api/teaching/instances.ts` 的共享实现函数已从 `getTeacherInstances()` / `destroyTeacherInstance()` 收口到中性 `getInstanceDirectory()` / `destroyManagedInstance()`，`api/admin/teaching.ts` 也改成显式 `getPlatformInstances()` / `destroyPlatformInstance()` 包装，不再 alias teacher 命名函数。当前实例目录这条线前端本地残留的 teacher 语义，已进一步收敛到后端 `/api/v1/teacher/instances*` 路径和 teacher public wrapper 命名。
  - `2026-05-29` contest route target cleanup 进展：`ContestList.vue` 的详情入口已收口到 `contestListRoutes.ts`；当前继续推进 `ContestManage.vue` 这一条，把 `useContestManagePage.ts` 里剩余的编辑 / 运维 / 公告完整页薄导航改成显式 route target contract，目标是继续压缩 `platform/contests` 的 feature router allowlist，而不触碰竞赛创建、公告抽屉和 readiness override workflow owner。
  - `2026-05-29` contest ops route target cleanup 进展：`ContestManage.vue` 已收口编辑 / 运维 / 公告完整页 route target；当前继续推进 `ContestOperationsHub.vue` 这一条，把 `useContestOperationsHubPage.ts` 里剩余的“进入单场运维台 / 返回竞赛目录”薄导航改成显式 route target contract，目标是继续压缩 `platform/contests` 的 feature router allowlist，而不触碰 AWD 赛事目录加载、preferred contest 和分页 owner。
  - `2026-05-29` contest announcements route target cleanup 进展：`ContestOperationsHub.vue` 已收口返回目录与进入运维台 route target；当前继续推进 `ContestAnnouncements.vue` 这一条，把 `useContestAnnouncementsPage.ts` 里剩余的“返回竞赛工作台”薄导航改成显式 route target contract，目标是继续压缩 `platform/contests` 的 feature router allowlist，而不触碰公告列表加载、发布和删除 workflow owner。
  - `2026-05-29` contest operations route target cleanup 进展：`ContestAnnouncements.vue` 已收口返回竞赛工作台 route target；当前继续推进 `ContestOperations.vue` 这一条，把 `useContestOperationsPage.ts` 里剩余的 `contestId` route param 输入 owner 改成显式 route props contract，目标是继续压缩 `platform/contests` 的 feature router allowlist，而不触碰赛事详情加载、breadcrumb workflow 和 runtime/readiness 判定 owner。
  - `2026-05-29` contest edit route target cleanup 进展：`ContestOperations.vue` 已收口 `contestId` route param owner；当前继续推进 `ContestEdit.vue` 这一条，把 `useContestEditPage.ts` 里剩余的 `contestId`、返回目录、公告页、AWD 配置页与保存成功后的目录跳转统一改成 route props + route target contract，目标是清空 `platform/contests` 最后一条 feature router allowlist，而不触碰 AWD workbench 的数据加载和 query-tab owner。
  - `2026-05-29` 第一批 route target cleanup 进展：`platform/contests` 已清空后，当前开始收第一批低复杂度残余条目：`scoreboard detail`、`awd challenge library`、`register`、`challenge import manage`、`challenge import preview`。目标是优先拿掉“只读 route param / 单条薄导航 / success redirect”这批 allowlist，而不提前进入 `challenge-list`、`student-dashboard`、`audit-log` 这类更重的 query/workflow owner。
  - `2026-05-29` 第一批 route target cleanup 收口：`useScoreboardDetailPage.ts` 与 `useChallengeImportPreviewPage.ts` 已把 `contestId / importId` 下沉为 route props owner；`useAwdChallengeLibraryPage.ts`、`useChallengeImportManagePage.ts`、`useChallengeImportPreviewPage.ts` 的薄导航已改成显式 route target contract；`useRegisterPage.ts` 与 challenge import 两页的成功跳转已改由 `AppRouteRedirect` 承接。因此 `featureRouterImportAllowlist` 这一批 5 条已清空，下一轮可以回到第二批中等复杂度 route-aware page owner。
  - `2026-05-29` 第二批 route owner cleanup 进展：`useNotificationDetailPage.ts` 已把 `id` 改为 route props 输入，并把“返回通知列表 / 查看关联对象”收口成 route target 或外链 transport；`useRouteQueryTabs.ts` 已改成共享 query-tab route owner，`useScoreboardRoutePage.ts` 与 `usePlatformChallengeDetailRoutePage.ts` 不再各自透传 `vue-router`。因此 `featureRouterImportAllowlist` 这一轮又净收掉 3 条，下一批继续回到 `useChallengeListPage.ts` 这类仍混 query/filter sync 的 page owner。
  - `2026-05-29` challenge list route owner cleanup 进展：`useChallengeListPage.ts` 已去掉 `vue-router`，筛选 query sync 改为消费薄的 `routeQueryTransport.ts` transport；返回仪表盘、能力画像和题目详情也已收口成显式 route target，`ChallengeList.vue` 与 `ChallengeDirectoryRow.vue` 现在通过 `AppRouteLink` 直接消费这些目标路由。这样 `featureRouterImportAllowlist` 又再减少 1 条，下一轮可继续看 `student-dashboard`、`contest-detail` 或 `audit-log` 这类仍混 route/query owner 的条目。
  - `2026-05-29` student dashboard route owner cleanup 进展：`useStudentDashboardPage.ts` 已去掉 `vue-router`，`panel` query tab 改回直接复用 `useRouteQueryTabs()` 内部的共享 route transport；新增 `studentDashboardRoutes.ts` 统一描述题库 / 分类 / 难度 / 能力画像 / 题目详情与 teacher/admin redirect 的 route target，同时新增薄的 `routeNavigationTransport.ts` 承接 `push / replace`。这样 student dashboard page owner 保留了 mounted、data load 和 panel binding，不再直接碰 router，`featureRouterImportAllowlist` 再收掉 `useStudentDashboardPage.ts` 这一条。下一轮继续看 `contest-detail` 或 `audit-log` 这类仍同时混 query 与 workflow 的条目。
  - `2026-05-29` audit log route owner cleanup 进展：`useAuditLogPage.ts` 已去掉 `vue-router`，筛选 hydrate、分页和 auto-apply 的 query sync 改为直接消费已有的 `routeQueryTransport.ts`；audit log page owner 继续保留 query normalize、节流和请求取消，不再自己拿 `useRoute/useRouter`。`AuditLog.test.ts` 和 `auditLogPageStateExtraction.test.ts` 也已补上“page model 不再 import vue-router、shared query transport 仍持有 router”护栏，因此 `featureRouterImportAllowlist` 再收掉 `useAuditLogPage.ts` 这一条。当前下一批优先回到 `contest-detail` 这类仍同时混 route param、workspace state 与 query sync 的条目。
  - `2026-05-29` contest detail route owner cleanup 进展：`useContestDetailRoutePage.ts` 已去掉 `vue-router` 与 `useUrlSyncedTabs()`，改为复用 `useRouteQueryTabs()` 和扩过的 `routeQueryTransport.ts`；其中 `contestId` route param、`challenge / panel` query 读取与写回、AWD 默认页签和 contest 派生状态仍留在 page owner，自定义业务规则没有继续下沉到 shared transport。`ContestDetail.test.ts` 也已补上“route page 不再 import vue-router / useUrlSyncedTabs、shared transport 持有 params/query”护栏，因此 `featureRouterImportAllowlist` 再收掉 `useContestDetailRoutePage.ts` 这一条。当前下一批可以回到 `platform/user-management`、`class-students-workspace` 或 `student-analysis-workspace` 这类仍含更重 query / redirect owner 的条目。
  - `2026-05-29` platform user management route owner cleanup 进展：`usePlatformUserManagePage.ts` 已去掉 `vue-router`，overview / import 的 `panel` query 读取与切换改为直接消费已有的 `routeQueryTransport.ts`；`useUserGovernancePanelRoute.ts` 继续只保留 panel 解析 / query 构建 helper，用户治理页的 mounted refresh、筛选、分页、删除确认和表单 workflow owner 都保持在 page model。`UserManage.test.ts` 也已补上“page model 不再 import vue-router、改为复用共享 query transport”护栏，因此 `featureRouterImportAllowlist` 再收掉 `usePlatformUserManagePage.ts` 这一条。当前下一批继续回到 `class-students-workspace`、`student-analysis-workspace` 或 `platform/challenges` 这类仍含 query / route-aware owner 的条目。
  - `2026-05-29` class students route owner cleanup 进展：`useClassStudentsPage.ts` 已去掉 `vue-router`，route `name / params / query` 改为直接消费扩过的 `routeQueryTransport.ts`，班级管理 / 教学概览 / 学员分析 3 条薄导航改为通过本地 `classStudentsRoutes.ts` 生成 route target，并交给 `routeNavigationTransport.ts` 执行；alias route 的 canonical redirect、insight window query owner、班级工作区加载和 stale request 语义都继续留在 page model。`TeacherClassStudents.test.ts`、`PlatformClassStudents.test.ts` 与 `TeacherClassWorkspaceSection.test.ts` 也已补上“page model 不再 import vue-router、redirect 改走 shared navigation transport、本地 route helper 统一承接薄导航”护栏，因此 `featureRouterImportAllowlist` 再收掉 `useClassStudentsPage.ts` 这一条。当前下一批继续回到 `student-analysis-workspace`、`platform/challenges` 或 `challenge-detail` 这类仍含 route-aware owner 的条目。
  - `2026-05-29` student analysis page route owner cleanup 进展：`useStudentAnalysisPage.ts` 已去掉 `vue-router`，route `params / query` 读取改为直接消费 `routeQueryTransport.ts`，review workspace query 写回改为通过 `replaceQuery()` 合并当前 query；班级学生页 / 题目详情 / 复盘归档 3 条薄导航则统一落到本地 `studentAnalysisRoutes.ts`，并交给 `routeNavigationTransport.ts` 执行。`useStudentAnalysisNavigation.ts` 现在只消费 route target callback，不再让 page model 手写 route path / route name 拼装；`TeacherStudentAnalysis.test.ts` 与 `PlatformStudentAnalysis.test.ts` 也已补上“page model 不再 import vue-router、挑战详情改走命名 route target”护栏，因此 `featureRouterImportAllowlist` 再收掉 `useStudentAnalysisPage.ts` 这一条。当前下一批继续回到 `platform/challenges`、`platform/challenge-detail` 或 `contest-awd-config` 这类仍含 route-aware owner 的条目。
  - `2026-05-28` class / student analysis contract naming 进展：班级洞察共享 query / summary / trend / review DTO、学员目录 DTO、证据链 DTO、攻击会话 DTO、证据筛选 query 与班级报告导出 payload 已继续从 `Teacher*` 前缀收口到中性 contract；`api/teaching/classes.ts` / `api/teaching/students.ts` 的局部 normalize / raw response 命名、teacher wrapper 对应 raw type，以及 `class-insight-window` query helper 名称也已同步收口。当前这条 P1 在 shared class / student analysis contract 层的 teacher 语义，已进一步收敛到显式 teacher wrapper / route name / transport path 这类保留边界。
  - `2026-05-28` 结项说明：再次核查 `views/platform`、shared features 与 shared widgets 后，`/platform/*` 已不再直接依赖 teacher route view，也不再在共享页面 / widget / query sync 里直连 `@/api/teacher` 或 teacher 前缀 shared contract。当前剩余的 teacher 命名主要位于 teacher 自有 feature、显式 public wrapper、teacher route name 与后端 `/api/v1/teacher/*` transport path，属于角色边界而不是 admin / teacher 结构耦合；因此这条 `P1` 结项，后续残留工作转入各自的 `P2/P3` debt。
  - `2026-05-28` 状态重估：按当前 backlog 范围，这已经是剩余唯一实质未完成的 `P1`。前端本地 admin / teacher 耦合已基本收敛到显式 transport / public wrapper / route 语义；下一步要么继续切最后一刀，要么把这些显式保留边界写成例外说明后降级出 `P1`。

- [x] P1：把应属于单一 feature 的 page-sized UI 从 `components/**` 继续收口到 `features/*/ui`
  - 依据：`docs/reviews/architecture/2026-05-24-frontend-architecture-review.md` 指出 allowlist 仍在冻结历史例外；当前题解管理三件套就是典型的 `components/*Page.vue -> @/features/*` 例外。
  - 收益：可以把“单一 feature 的 UI 壳”从 legacy component page 通道里迁走，减少 `componentFeatureImportAllowlist` 和 `legacyComponentPageAllowlist`，也给后续其它切片提供明确落点。
  - 风险：中。主要是 import 路径、raw-source 测试和 public API 更新，需要避免把 route owner 或 API owner 从 feature model 重新打散。
  - `2026-05-27` 进展：题解管理三件套 `ChallengeWriteupManagePanel`、`ChallengeWriteupEditorPage`、`ChallengeWriteupViewPage` 已完成迁入 `features/challenge-writeup-editor/ui`，前端架构文档同步补上 `feature-owned UI` 判定规则，题解这组对应的 `componentFeatureImportAllowlist` 与 `legacyComponentPageAllowlist` 已收掉；后续继续优先处理仍然直接依赖单一 feature model 的 legacy component page / panel。
  - `2026-05-27` platform overview 进展：`PlatformOverviewPage.vue` 已迁入 `features/platform/overview/ui`，`PlatformOverview` route 改为直接从 `features/platform/overview` public API 组合 page model 与 page shell；平台总览对应的一条 `componentFeatureImportAllowlist` 和一条 `legacyComponentPageAllowlist` 已收掉，下一批低风险候选可继续看 `TeacherDashboardPage.vue`。
  - `2026-05-27` teacher dashboard 进展：`TeacherDashboardPage.vue` 已迁入 `features/teacher-dashboard/ui`，`TeacherDashboard` route 改为直接从 `features/teacher-dashboard` public API 组合 page model 与 page shell；教师总览对应的一条 `componentFeatureImportAllowlist` 和一条 `legacyComponentPageAllowlist` 已收掉，下一批同模式候选可以继续看 `ContestOrchestrationPage.vue`。
  - `2026-05-27` contest orchestration 进展：`ContestOrchestrationPage.vue` 已迁入 `features/platform/contests/ui`，竞赛编辑页 / 运维页跳转 owner 已从 page shell 回收到 `useContestManagePage()`；竞赛目录 route 改为直接从 `features/platform/contests` public API 组合 page model 与 page shell，对应的一条 `componentFeatureImportAllowlist` 和一条 `legacyComponentPageAllowlist` 已收掉。
  - `2026-05-27` topology studio 进展：`ChallengeTopologyStudioPage.vue` 已迁入 `features/challenge-topology-studio/ui`，拓扑编辑 route 改为直接从 `features/challenge-topology-studio` public API 组合 page model 与 page shell；拓扑页对应的一条 `componentFeatureImportAllowlist` 和一条 `legacyComponentPageAllowlist` 已收掉，下一批同模式候选可继续看 `UserGovernancePage.vue`。
  - `2026-05-27` user governance 进展：`UserGovernancePage.vue` 已迁入 `features/platform/user-management/ui`，`panel=overview/import` 的 query owner 也一并从 page shell 收口到 `useUserGovernancePanelRoute()`；`UserManage` route 改为直接从 `features/platform/user-management` public API 组合 page model 与 page shell，用户治理页对应的一条 `legacyComponentPageAllowlist` 已收掉，并补上 feature model 的 router owner 白名单。
  - `2026-05-29` user governance panel route owner 进展：`useUserGovernancePanelRoute.ts` 已降为纯 panel 解析 / query 构建 helper，不再直接 import `vue-router`；`panel=overview/import` 的真正 route/query owner 已并回 `usePlatformUserManagePage.ts`，`UserGovernancePage.vue` 也改回纯 props / emits 壳。这个收口主要是把 allowlist owner 从非 page helper 挪回 page owner，所以总量未下降，但落点更合理。
  - `2026-05-29` auth router owner 进展：`useAuth.ts` 已退回纯 auth side-effect owner，只保留 API / store / toast，不再直接 import `vue-router`；登录 redirect 与 sanitize owner 已并回 `useLoginPage.ts`，注册成功跳转回到 `useRegisterPage.ts`，退出登录后的 `/login` 跳转回到 `widgets/layout-shell/useLayoutSessionActionsBridge.ts`。这一轮同时把 `useLoginViewPage.ts` 降成纯 redirect target helper，因此 auth 线上的 allowlist owner 已从通用 helper 收拢到更明确的 page / layout workflow owner，`featureRouterImportAllowlist` 总量保持不变但落点更合理。
  - `2026-05-29` AWD review index router owner 进展：`useAwdReviewIndex.ts` 已去掉 `vue-router`，不再直接持有“返回教师总览 / 平台概览”和“进入 AWD 复盘详情”的导航；由于 route view 本身也不允许直接拿 `useRouter`，这轮额外补了 `useAwdReviewIndexPage(scope)` 作为显式 route-aware page wrapper，由它按 `teacher / platform` scope 持有导航 owner，`TeacherAWDReviewIndex.vue` 和 `AWDReviewIndex.vue` 继续保持薄壳。这样 `featureRouterImportAllowlist` 收掉了 `useAwdReviewIndex.ts` 这一条，只留下更明确的 AWD review index page wrapper 入口。
  - `2026-05-29` AWD review index route target 进展：`useAwdReviewIndexPage.ts` 也已去掉 `vue-router`，新增 `awdReviewIndexRoutes.ts` 统一提供角色感知的“返回概览 / 进入复盘详情” route target contract；教师目录头部、教师目录行、平台 hero 返回入口和平台目录“进入复盘”入口现在都直接通过 `AppRouteLink` 消费这些目标路由，目录加载、筛选、分页和刷新 owner 保持不变。`TeacherAWDReviewIndex.test.ts`、`AWDReviewIndex.test.ts`、`AwdReviewIndexWorkspace.test.ts` 与 `AwdReviewContestDirectory.test.ts` 也已补上“page wrapper 不再 import vue-router、workspace / panel 直接消费 route target”护栏，因此 `featureRouterImportAllowlist` 再收掉 `useAwdReviewIndexPage.ts` 这一条。
  - `2026-05-29` contest list route target 进展：`useContestListPage.ts` 已去掉 `vue-router`，新增 `contestListRoutes.ts` 统一提供竞赛详情 route target contract；`ContestList.vue` 的竞赛目录行现在直接通过 `AppRouteLink` 进入详情页，竞赛列表加载、筛选、分页、时间格式化和状态/模式文案 owner 保持不变。`ContestList.test.ts` 也已补上“page model 不再 import vue-router、目录行直接消费 route target、点击后命中 `ContestDetail`”护栏，因此 `featureRouterImportAllowlist` 再收掉 `useContestListPage.ts` 这一条。
  - `2026-05-29` challenge package format router target 进展：`useChallengePackageFormatPage.ts` 已从只做一次 `router.push()` 的薄 wrapper 收口为纯 route target contract，`ChallengePackageFormat.vue` 现在直接通过 `RouterLink` 消费返回导入页的目标路由；`featureRouterImportAllowlist` 又收掉了 `useChallengePackageFormatPage.ts` 这一条。当前 challenge package import 线上剩余的 router owner，开始只剩真正承接导入管理 / 预览 workflow 的 page owner。
  - `2026-05-27` teacher instance 进展：`TeacherInstanceManagementPage.vue` 已迁入 `features/teacher-instances/ui`，`InstanceManagement` route 改为直接从 `features/teacher-instances` public API 组合 page model 与 page shell；教师实例页对应的一条 `legacyComponentPageAllowlist` 已收掉，教师端 raw-source surface 测试也已切到新 owner。
  - `2026-05-27` student management 进展：`StudentManagementPage.vue` 已迁入 `features/teacher-student-management/ui`，`TeacherStudentManagement` route 改为直接从 `features/teacher-student-management` public API 组合 page model 与 page shell，并继续在 route 壳里组合 `ClassReportExportDialog`；教师学生管理页对应的一条 `legacyComponentPageAllowlist` 已收掉，教师端 raw-source surface / header / pagination 测试也已切到新 owner。
  - `2026-05-27` remaining legacy pages 进展：`ClassManagementPage.vue`、`AWDChallengeLibraryPage.vue`、`ClassStudentsPage.vue`、`StudentAnalysisPage.vue` 已分别迁入 `features/teacher-class-management/ui`、`features/platform/awd-challenges/ui`、`features/class-students-workspace/ui`、`features/student-analysis-workspace/ui`；对应 teacher / platform route view 统一改为直接从 feature public API 组合 page model 与 page shell，`components/class-management/index.ts` 这个只服务旧 page path 的 barrel 也已退场。当前 `legacyComponentPageAllowlist` 已只剩 student dashboard 的 5 个页面，`AWDChallengeLibraryPage.vue` 对应的一条 `componentFeatureImportAllowlist` 也已收掉。
  - `2026-05-27` student dashboard 进展：`StudentCategoryProgressPage.vue`、`StudentDifficultyPage.vue`、`StudentOverviewPage.vue`、`StudentRecommendationPage.vue` 与 `dashboardPanelRegistry.ts` 已迁入 `features/student-dashboard/ui`，`DashboardView` 改为直接从 `features/student-dashboard` public API 读取 page model 与 panel registry；对应的 4 条 `legacyComponentPageAllowlist` 和 1 条 `componentFeatureImportAllowlist` 已收掉。当前这一条 P1 剩余的 student dashboard 存量主要只剩 `StudentTimelinePage.vue`，因为它仍被 teacher 学员洞察复用，需要下一刀按共享 panel owner 单独收口。
  - `2026-05-27` timeline panel 进展：`StudentTimelinePage.vue` 已收口为中立 `components/training/TrainingTimelinePanel.vue`，student dashboard registry 与 `StudentInsightPanel` 都改为消费共享训练时间线 panel；student dashboard 这条存量 page-sized 组件已清空，`legacyComponentPageAllowlist` 不再保留 dashboard 历史 page 例外。
  - `2026-05-27` platform AWD challenge feature UI 进展：`AWDChallengeEditorDialog.vue` 与 `AwdChallengeImportSection.vue` 已迁入 `features/platform/awd-challenges/ui`，`AWDChallengeLibrary.vue` route view 与 `AWDChallengeLibraryPage.vue` 不再直连旧 `components/platform/awd-service` 路径；对应的 2 条 `componentFeatureImportAllowlist` 已收掉。当前这一条后续更适合继续瞄准 `platform/challenge-detail`、`contest-workbench` 这类仍有单一 feature UI 滞留在 `components/**` 的残片。
  - `2026-05-27` platform challenge detail feature UI 进展：`AdminChallengeTopbarPanel.vue`、`AdminChallengeWorkspaceTabs.vue`、`AdminChallengeProfilePanel.vue` 已迁入 `features/platform/challenge-detail/ui`，`PlatformChallengeDetailWorkspace.vue` 改为通过 `features/platform/challenge-detail` public API 组合题目详情 topbar / tabs / profile；对应的 2 条 `componentFeatureImportAllowlist` 与 2 条 `widgetLegacyComponentImportAllowlist` 已收掉。当前这一条后续更适合继续瞄准 `contest-workbench`、`contest-announcements` 这类仍有单一 feature UI 滞留在 `components/**` 的残片。
  - `2026-05-27` contest announcements feature UI 进展：`ContestAnnouncementRealtimeBridge.vue` 与 `ContestAnnouncementManageDrawer.vue` 已迁入 `features/contest-announcements/ui`，竞赛详情页 / 竞赛目录页改为从 `features/contest-announcements` public API 取 bridge / drawer；“进入完整管理页”的导航 owner 也已从 drawer 收回 `useContestManagePage()`，对应的 2 条 `componentFeatureImportAllowlist` 已收掉。当前这一条后续更适合继续瞄准 `contest-workbench` 这类仍有单一 feature UI 滞留在 `components/**` 的残片。
  - `2026-05-27` contest workbench feature UI 进展：`ContestWorkbenchStageTabs.vue`、`ContestWorkbenchSummaryStrip.vue`、`ContestChallengeFilterStrip.vue`、`ContestChallengeSummaryStrip.vue`、`ContestChallengeOrchestrationPanel.vue` 已迁入 `features/contest-workbench/ui`，`ContestEditWorkspacePanel.vue` 也已按 route shell owner 迁入 `features/platform/contests/ui`；`ContestEdit.vue` 现在通过 `features/contest-workbench` 与 `features/platform/contests` public API 组合工作台，对应的 6 条 `componentFeatureImportAllowlist` 已收掉。
  - `2026-05-28` contest edit feature UI 进展：`ContestEditTopbarPanel.vue` 已迁入 `features/platform/contests/ui`，`ContestChallengeEditorDialog.vue`、`ContestAwdChallengeSelectorSection.vue`、`ContestChallengeSettingsSection.vue` 已整体迁入 `features/contest-workbench/ui`；`ContestEdit.vue` 与 `ContestChallengeOrchestrationPanel.vue` 不再回头引用旧 `components/platform/contest/*` 路径。当前这一条在 contest edit 线上的剩余低风险候选已经收敛到 `AWDChallengeConfigPanel.vue`。
  - `2026-05-28` AWD challenge config panel 进展：`AWDChallengeConfigPanel.vue` 已迁入 `features/platform/contests/ui`，`ContestEditWorkspacePanel.vue` 改为 feature 内部相对 import，`componentFeatureImportAllowlist` 中这条 `AWDChallengeConfigPanel.vue -> @/features/awd-inspector` 历史例外已收掉。当前这一条在 contest edit / platform contest 线上的低风险剩余候选，进一步收敛到 `PlatformContestFormPanel.vue`、`PlatformContestFormDialog.vue` 这类仍留在 `components/platform/contest/*` 的单一 feature UI。
  - `2026-05-28` platform contest form feature UI 进展：`PlatformContestFormPanel.vue` 与 `PlatformContestFormDialog.vue` 已迁入 `features/platform/contests/ui`，`ContestManage.vue` 改为通过 `features/platform/contests` public API 组合 dialog，`ContestOrchestrationPage.vue` 与 `ContestEditWorkspacePanel.vue` 也已切到 feature 内部 panel import；对应的 2 条 `componentFeatureImportAllowlist` 已收掉。当前这一条在 `platform/contests` 线上剩余的低风险候选，开始进一步收敛到 `PlatformContestTable.vue` 这类仍留在旧 contest 组件目录的展示 UI。
  - `2026-05-28` platform contest table feature UI 进展：`PlatformContestTable.vue` 已迁入 `features/platform/contests/ui`，`ContestOrchestrationPage.vue` 改为 feature 内部相对 import，相关 raw-source / typography / surface alignment 测试与组件声明也已切到新 owner。当前这一条在 `platform/contests` 线上的低风险遗留展示 UI 继续缩小，下一批更适合回到其它 feature 的单一 owner surface 或更大颗粒度的大组件壳体。
  - `2026-05-28` contest AWD preflight panel feature UI 进展：`ContestAwdPreflightPanel.vue` 已迁入 `features/platform/contests/ui`，`ContestEditWorkspacePanel.vue` 改为 feature 内部相对 import，相关 raw-source / theme token 护栏与组件声明也已切到新 owner。这一条在 `platform/contests` 线上的 contest edit 展示残片继续缩小；后续如果继续清 AWD readiness 相关 UI，应单独判断 `AWDReadinessChecklist.vue` / `AWDReadinessDecisionHUD.vue` 的 primitive owner，而不是顺手混入这刀。
  - `2026-05-28` contest announcements panel feature UI 进展：`ContestAnnouncementsTopbarPanel.vue` 与 `ContestAnnouncementsWorkspacePanel.vue` 已迁入 `features/platform/contests/ui`，`ContestAnnouncements.vue` 改为通过 `@/features/platform/contests` public API 组合这组 panel。当前 `platform/contests` 线上与公告路由直接绑定的 page-sized UI 继续从旧 `components/platform/contest/*` 收口到 feature owner。
  - `2026-05-28` contest operations hub panel feature UI 进展：`ContestOperationsHubHeroPanel.vue` 与 `ContestOperationsHubWorkspacePanel.vue` 已迁入 `features/platform/contests/ui`，`ContestOperationsHub.vue` 改为通过 `@/features/platform/contests` public API 组合运维目录 panel。当前 `platform/contests` 线上与赛事运维目录路由直接绑定的 page-sized UI 继续从旧 `components/platform/contest/*` 收口到 feature owner。
  - `2026-05-28` awd inspector ui cluster feature UI 进展：`AWDRoundInspector.vue`、`AWDTrafficPanel.vue`、`AWDServiceStatusPanel.vue`、`AWDScoreboardSummaryPanel.vue`、`AWDRoundHeaderPanel.vue`、`AWDAttackLogPanel.vue`、`AWDServiceAlertBanner.vue` 与 `awdInspector.types.ts` 已整体迁入 `features/awd-inspector/ui`，`AWDOperationsPanel.vue` 与 `ContestOperations.vue` 改为通过 `@/features/awd-inspector` public API 组合 inspector UI；对应的 3 条 `componentFeatureImportAllowlist` 历史例外已收掉。当前这条线上的后续重点开始回到 `AWDReadiness*`、`AWDInstanceOrchestrationPanel` 这类更贴近 `contest-awd-admin` 的 runtime cluster owner。
  - `2026-05-28` awd readiness ui cluster feature 进展：`AWDReadinessChecklist.vue`、`AWDReadinessDecisionHUD.vue`、`AWDReadinessSummary.vue`、`AWDReadinessOverrideDialog.vue` 已整体迁入 `features/awd-readiness/ui`，`ContestAwdPreflightPanel.vue`、`AWDOperationsPanel.vue`、`ContestManage.vue` 改为通过 `@/features/awd-readiness` public API 组合 readiness UI。当前 AWD readiness capability 的 UI owner 已从历史共用组件目录收口出来；下一批更适合回到 `AWDInstanceOrchestrationPanel.vue` 或 shared readiness workflow model 是否继续收口。
  - `2026-05-28` awd instance orchestration panel feature UI 进展：`AWDInstanceOrchestrationPanel.vue` 已迁入 `features/contest-awd-admin/ui`，`AWDOperationsPanel.vue` 改为 feature 内部相对 import，实例编排 panel 不再继续滞留在旧 `components/platform/contest/*` 路径。当前 `contest-awd-admin` runtime cluster 的后续重点进一步收敛到 `AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue`、`AWDRoundCreateDialog.vue`。
  - `2026-05-28` awd runtime dialog cluster feature UI 进展：`AWDRoundCreateDialog.vue`、`AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue` 已迁入 `features/contest-awd-admin/ui`，`AWDOperationsPanel.vue` 改为 feature 内部相对 import，相关 raw-source / duplicate-action / dialog adoption 测试与组件声明也已切到新 owner。当前 `contest-awd-admin` runtime cluster 在 touched surface 上的 legacy dialog 路径已继续缩小。
  - `2026-05-28` awd operations shell primitives feature UI 进展：`AWDContestSelectorField.vue`、`AWDRuntimePendingState.vue` 已迁入 `features/contest-awd-admin/ui`，`AWDOperationsPanel.vue` 改为 feature 内部相对 import，相关 raw-source / theme token 护栏与组件声明也已切到新 owner。当前 `contest-awd-admin` 在 touched surface 上的 AWD operations 子件 legacy 路径继续缩小。
  - `2026-05-28` AWD operations panel feature UI 进展：`AWDOperationsPanel.vue` 已迁入 `features/contest-awd-admin/ui`，`ContestOperations.vue` 改为通过 `@/features/contest-awd-admin` public API 组合运维 panel；对应的 `components/platform/contest/AWDOperationsPanel.vue -> @/features/contest-awd-admin` allowlist 已收掉。当前这条在 AWD 运维线上的后续重点，开始回到更深层的 runtime 子件 owner，而不是继续保留 panel 本体在旧 `components/platform/contest/*`。
  - `2026-05-28` contest projector ui cluster feature UI 进展：`ContestProjectorToolbar.vue`、`ContestProjectorHero.vue`、`ContestProjectorAttackMap.vue`、`ContestProjectorFocusOverlay.vue` 及其 projector cluster 子件 / 样式已迁入 `features/contest-projector/ui`，`ContestProjector.vue` 改为通过 `@/features/contest-projector` public API 组合 page model 与 UI。当前 `contest-projector` 已同时持有 model 与 UI owner，旧 `components/platform/contest/projector/*` 路径开始退出主消费面。
  - `2026-05-28` allowlist A 第一批进展：`AdminNotificationPublishDrawer.vue`、`PlatformUserFormDialog.vue`、`ChallengeManageDirectoryPanel.vue`、`ScoreboardRealtimeBridge.vue`、`InterventionPanel.vue`、`ClassReportExportDialog.vue` 已分别迁入各自 `features/*/ui` 并通过 public API 暴露；由于 `ScoreboardRealtimeBridge` 的唯一 legacy component consumer 是 `ContestAWDWorkspacePanel.vue`，这一轮也把它一并迁入 `features/contest-awd-workspace/ui`，避免留下新的 `components/** -> @/features/*` 违规。相关 route view / raw-source 测试 / 兼容 barrel / 组件声明已同步切到新 owner，`componentFeatureImportAllowlist` 从 `22` 条下降到 `15` 条。
  - `2026-05-28` challenge detail feature UI 进展：`ChallengeWorkspaceShell.vue`、`ChallengeSolutionsPanel.vue`、`ChallengeSubmissionRecordsPanel.vue` 已迁入 `features/challenge-detail/ui`，`ChallengeDetail.vue` 改为通过 `@/features/challenge-detail` public API 组合 workspace shell；对应的 3 条 `componentFeatureImportAllowlist` 已收掉。当前这一条后续主要回到 contest / AWD feature 内更深层 owner 清理，而不再是单纯 feature 私有 UI 落位错误。
  - `2026-05-28` contest AWD workspace defense cluster 进展：`AWDDefenseColumn.vue`、`AWDDefenseOperationsPanel.vue`、`AWDDefenseServiceList.vue` 已迁入 `features/contest-awd-workspace/ui`，`ContestAWDWorkspacePanel.vue` 改为 feature 内部相对 import defense cluster；对应的 3 条 `componentFeatureImportAllowlist` 已收掉。当前这一条后续主要回到 contest / AWD feature 内更深层 owner 清理，而不再是单纯 feature 私有 UI 落位错误。
  - `2026-05-29` topology studio feature UI 进展：`TopologyCanvasBoard.vue`、`TopologyCanvasQuickEditor.vue`、`TopologyCanvasWorkspaceSection.vue`、`TopologyChallengeContextRail.vue`、`TopologyChallengeWorkbench.vue`、`TopologyChallengeWorkspaceHeader.vue`、`TopologyConnectivitySections.vue`、`TopologyEntryNodeSection.vue`、`TopologyNetworkQuickEditor.vue`、`TopologyNetworkSection.vue`、`TopologyNodeEditor.vue`、`TopologyNodeSection.vue`、`TopologyPackageContextPanel.vue`、`TopologyStatusNotes.vue`、`TopologySummaryGrid.vue`、`TopologyTemplateHeroSection.vue`、`TopologyTemplateLibraryHeader.vue`、`TopologyTemplateSidePanel.vue`、`TopologyTemplateWorkbench.vue` 已整体迁入 `features/challenge-topology-studio/ui`，`ChallengeTopologyStudioPage.vue` 与对应 raw-source / theme / async chunk 护栏已切到 feature 内新 owner；剩余的 6 条 `componentFeatureImportAllowlist` 已全部收掉。
  - `2026-05-29` challenge entity contract 进展：`entities/challenge/model/presentation.ts` 与 `ChallengeCategoryDifficultyPills.vue`、`ChallengeCategoryPill.vue`、`ChallengeCategoryText.vue`、`ChallengeDifficultyText.vue`、`ChallengeDirectoryRow.vue`、`ChallengeMetaStrip.vue`、`ChallengeProfileMetaGrid.vue`、`ChallengeProfileSummaryStrip.vue` 已切到 entity 内本地展示类型，不再直接 import `@/api/contracts`；`commonForbiddenImportAllowlist` 已从 10 条收敛到仅剩 `components/common/InstancePanel.vue` 这 1 条。
  - `2026-05-29` instance panel contract 进展：`InstancePanel.vue` 已改为消费 `components/common/instancePanel.types.ts` 内的本地最小展示类型，不再直接 import `@/api/contracts`；对应 raw-source 护栏也已补上 `@/api/contracts` 负向断言。
  - `2026-05-29` utility contract 进展：`utils/contest.ts`、`utils/platformContestAwdChallengeLinks.ts`、`utils/skillProfile.ts` 已切到各自本地最小 contract，不再直接 import `@/api/contracts`；`skillProfile` raw-source 护栏也已补上 `@/api/contracts` 负向断言。
  - `2026-05-29` utility owner 进展：`contest` 展示 helper 已迁入 `entities/contest`，`skill profile / recommendation` normalize 与展示 helper 已迁入 `entities/skill-profile`，`AWD service -> challenge link` mapper 已迁入 `entities/contest-awd-challenge-link`；旧 `utils/contest.ts`、`utils/skillProfile.ts`、`utils/platformContestAwdChallengeLinks.ts` 已退出主路径。
  - `2026-05-29` websocket composable 边界进展：`useWebSocket.ts` 中历史残留的 `useAuthStore` import 已移除，shared composable 边界从 `api+store` 收回到 `api+runtime`；`composableMultiBoundaryAllowlist` 已清空。
  - `2026-05-29` challenge manage router owner 进展：`useChallengeManagePresentation.ts` 已改为消费导航 callback，不再直接 import `vue-router`；平台题目管理的路由动作统一回到 `useChallengeManagePage.ts`，`featureRouterImportAllowlist` 已收掉 `useChallengeManagePresentation.ts` 这一条。
  - `2026-05-29` student dashboard data router owner 进展：`useStudentDashboardData.ts` 已去掉 `Router` 依赖，teacher/admin 的 role redirect signal 改由 `useStudentDashboardPage.ts` 统一接住；`DashboardView` 测试已补“data 层不再 import vue-router”与“非 student 不继续请求 dashboard API”的护栏，`featureRouterImportAllowlist` 又收掉 `useStudentDashboardData.ts` 这一条。
  - `2026-05-29` AWD challenge selection router owner 进展：`useAwdChallengeSelection.ts` 已改为消费 `readServiceQuery` / `replaceServiceQuery` callback，不再直接 import `vue-router`；AWD 配置页的服务选择 query owner 统一回到 `useContestAwdConfigPage.ts`，`ContestAwdConfig` 测试也已补 `useAwdChallengeSelection.ts` 的 raw-source 护栏，`featureRouterImportAllowlist` 再收掉 `useAwdChallengeSelection.ts` 这一条。
  - `2026-05-29` 结项说明：`legacyComponentPageAllowlist`、`widgetLegacyComponentImportAllowlist`、`componentFeatureImportAllowlist`、`commonForbiddenImportAllowlist`、`utilityBoundaryImportAllowlist` 与 `composableMultiBoundaryAllowlist` 已全部清空；对应的 `contest / skillProfile / AWD link mapper` 已从历史 `utils/*` 落位纠偏到明确 owner，`useWebSocket.ts` 也已去掉误挂的 store 边界。`featureRouterImportAllowlist` 则开始进入逐条判定和切片收口阶段，当前已先清掉 `useChallengeManagePresentation.ts`、`useStudentDashboardData.ts` 与 `useAwdChallengeSelection.ts` 这三条非 page owner 例外。
  - `2026-05-29` student analysis helper router owner 进展：`useStudentAnalysisNavigation.ts` 已改为消费显式导航 callback，`useStudentAnalysisReviewQuerySync.ts` 已改为消费本地 `route-like` / `replaceReviewWorkspaceQuery` contract，不再直接 import `vue-router`；student analysis 的 route/query owner 统一回到 `useStudentAnalysisPage.ts`，`TeacherStudentAnalysis` 和 helper 单测也已补上 raw-source 护栏，`featureRouterImportAllowlist` 再收掉 `useStudentAnalysisNavigation.ts` 与 `useStudentAnalysisReviewQuerySync.ts` 这两条 helper 例外。
  - `2026-05-29` class workspace redirect router owner 进展：`useClassWorkspaceSection.ts` 已去掉 `vue-router` 依赖，只保留 alias route 到 canonical workspace target 的解析；班级工作区 alias redirect 现在并回 `useClassStudentsPage.ts` 这个既有 page owner，由它统一持有 `router.replace()`，`PlatformClassWorkspaceSection.vue` 与 `TeacherClassWorkspaceSection.vue` 继续保持薄 route shell，`featureRouterImportAllowlist` 再收掉 `useClassWorkspaceSection.ts` 这一条。
  - `2026-05-29` class workspace redirect feature merge 进展：由于 `useClassWorkspaceSection.ts` 现在只剩 `useClassStudentsPage.ts` 这一处 consumer，`features/class-workspace-redirect` 已并回 `features/class-students-workspace/model`，helper 改为同 feature 本地依赖，不再额外保留只服务单一 page owner 的独立 feature slice。
  - `2026-05-29` platform challenge route wrapper 进展：`useChallengeTopologyStudioRoutePage.ts`、`useChallengeWriteupPage.ts`、`useChallengeWriteupViewPage.ts` 已改为纯委托层；随后 `usePlatformChallengeRoutePage.ts` 也进一步去掉了 `vue-router`，改成通过 `routeQueryTransport.ts` 读取 `challengeId`、通过本地 `platformChallengeRoutes.ts` + `routeNavigationTransport.ts` 处理返回详情和“去题解编辑页”跳转。`ChallengeTopologyStudio.test.ts` 与 `ChallengeWriteup.test.ts` 已同步补上“route page 不再直接 import vue-router、改为复用 shared transports + 本地 route targets”护栏，因此 `featureRouterImportAllowlist` 再从这组净收掉最后 1 条。
  - `2026-05-29` challenge manage route target 进展：`useChallengeManagePage.ts` 也已去掉 `vue-router`，题目导入预览、题目详情、拓扑、题解面板和导入工作区导航统一收口到 `platformChallengeRoutes.ts`，再由 `routeNavigationTransport.ts` 执行；题目目录列表、排序、筛选与发布/删除 workflow owner 保持不变。`ChallengeManage.test.ts` 已补“page model 不再 import vue-router、platform/challenges 本地 route target 统一承接薄导航”护栏，因此 `featureRouterImportAllowlist` 再收掉 `useChallengeManagePage.ts` 这一条。
  - `2026-05-29` platform challenge detail route owner 进展：`usePlatformChallengeDetailPage.ts` 已去掉 `vue-router`，`challengeId` 改由 `routeQueryTransport.ts` 读取，返回题库、去拓扑、去题解查看/编辑以及加载失败后的延迟跳回统一收口到 `platformChallengeDetailRoutes.ts` + `routeNavigationTransport.ts`；题目详情加载、Flag draft、附件下载和延迟重定向计时器 owner 保持不变。`ChallengeDetail.test.ts` 也已补“page model 不再 import vue-router、detail feature 本地 route target 统一承接导航”护栏，因此 `featureRouterImportAllowlist` 再收掉 `usePlatformChallengeDetailPage.ts` 这一条。
  - `2026-05-29` contest awd config route owner 进展：`useContestAwdConfigPage.ts` 已去掉 `vue-router`，`contestId` 与 `service` query 改由 `routeQueryTransport.ts` 读取，`service` 写回直接复用共享 `replaceQuery()`，返回赛事工作台的导航统一收口到 `contestAwdConfigRoutes.ts` + `routeNavigationTransport.ts`；mounted 初始化、breadcrumb、service fallback 与 checker draft / preview / save workflow owner 保持不变。`ContestAwdConfig.test.ts` 也已补“page model 不再 import vue-router、AWD 配置 feature 本地 route target 统一承接 back navigation”护栏，因此 `featureRouterImportAllowlist` 再收掉 `useContestAwdConfigPage.ts` 这一条。
  - `2026-05-29` awd review detail route owner 进展：`useAwdReviewDetailPage.ts` 已去掉 `vue-router`，`contestId` 与 `round` 改由 `routeQueryTransport.ts` 读取，round 切换直接复用共享 `replaceQuery()`，返回索引导航统一收口到 `awdReviewDetailRoutes.ts` + `routeNavigationTransport.ts`；详情加载、summary 聚合、team drawer、export polling 和 breadcrumb owner 保持不变。`PlatformAwdReviewDetail.test.ts` 也已补“page model 不再 import vue-router、AWD review detail feature 本地 route target 统一承接返回导航”护栏，因此 `featureRouterImportAllowlist` 再收掉 `useAwdReviewDetailPage.ts` 这一条。
  - `2026-05-29` challenge detail route owner 进展：`useChallengeDetailPage.ts` 已去掉 `vue-router`，`challengeId` 改由 `routeQueryTransport.ts` 读取，错误态“返回题目列表”导航统一收口到 `challengeDetailRoutes.ts` + `routeNavigationTransport.ts`；challenge 加载、题解 / 提交记录 / writeup 预取、实例 workflow 与 page-level tab owner 保持不变。`ChallengeDetail.test.ts` 也已补“page model 不再 import vue-router、challenge detail feature 本地 route target 统一承接 back navigation”护栏，因此 `featureRouterImportAllowlist` 再收掉 `useChallengeDetailPage.ts` 这一条。
  - `2026-05-29` auth login route owner 进展：`useLoginPage.ts` 已去掉 `vue-router`，redirect query 改由 `routeQueryTransport.ts` 读取，登录成功跳转改由 `routeNavigationTransport.ts` 执行；同时 `sanitizeRedirectPath()` 已提升到中性 `utils/redirectPath.ts`，router guards 与 login page 复用同一实现，不再让 feature 直连 `router/guards.ts`。`useLoginPage.test.ts` 也已补“login page 不再 import vue-router / router guards、redirect 参数先做安全清洗”的护栏，因此 `featureRouterImportAllowlist` 已清空。
  - `2026-05-29` notification list route target 进展：`useNotificationListPage.ts` 已去掉 `vue-router`，通知详情跳转改成纯 `notificationDetailRoute()` helper；`NotificationList.vue` 的通知行现在直接通过 `RouterLink` 进入详情页，相关测试也已补上“page model 不再 import vue-router、route view 直接消费 route target”护栏。因此 `featureRouterImportAllowlist` 再收掉 `useNotificationListPage.ts` 这一条。
  - `2026-05-29` platform overview route target 进展：`usePlatformOverviewPage.ts` 与 `useCheatDetectionPage.ts` 都已去掉 `vue-router`，同 feature 内新增 `platformOverviewRoutes.ts` 统一提供审计日志与作弊检测页 route target contract。`PlatformOverviewPage.vue`、`PlatformOverviewHeroPanel.vue` 以及 `CheatDetectionWorkspacePanel.vue` / `CheatDetectionHeroPanel.vue` / `CheatDetectionReviewPanels.vue` 现在直接通过 `RouterLink` 消费这些 route target，刷新与数据加载 owner 保持不变；对应测试也已补上“page model 不再 import vue-router、overview/cheat surface 直接消费 route target”护栏，因此 `featureRouterImportAllowlist` 再从 `platform/overview` 这一组净收掉 2 条。
  - `2026-05-29` class management route target 进展：`usePlatformClassManagementPage.ts` 与 `useClassManagementPage.ts` 都已去掉 `vue-router`，分别改为暴露本地班级目录 route target contract；平台班级目录的“查看班级”以及教师班级目录的“教学概览 / 进入班级”现在直接通过 `RouterLink` 消费这些目标路由，教师端“导出班级报告”弹窗 owner 保持不变。`ClassManage.test.ts` 与 `ClassManagement.test.ts` 也已补上“page model 不再 import vue-router、目录 UI 直接消费 route target”护栏，因此 `featureRouterImportAllowlist` 再从班级目录这一组净收掉 2 条。
  - `2026-05-29` student management route target 进展：`usePlatformStudentManagementPage.ts` 与 `useStudentManagementPage.ts` 都已去掉 `vue-router`，分别改为暴露平台 / 教师学生目录本地的 route target contract；平台目录的“查看学员”以及教师目录的“班级管理 / 学员分析”现在统一通过 `AppRouteLink` 消费这些目标路由，教师端“导出班级报告”弹窗 owner 保持不变。`StudentManage.test.ts` 与 `TeacherStudentManagement.test.ts` 也已补上“page model 不再 import vue-router、目录 UI 直接消费 route target”护栏，因此 `featureRouterImportAllowlist` 再从学生目录这一组净收掉 2 条。
  - `2026-05-29` instance management route target 进展：`usePlatformInstanceManagementPage.ts` 与 `useInstanceManagementPage.ts` 都已去掉 `vue-router`，分别改为暴露平台 / 教师实例目录本地的 route target contract；平台实例页的“返回概览”“所属用户”以及教师实例页的“返回教学概览”现在统一通过 `AppRouteLink` 消费这些目标路由，实例销毁、刷新、筛选和分页 owner 保持不变。`InstanceManage.test.ts` 与 `InstanceManagement.test.ts` 也已补上“page model 不再 import vue-router、目录 UI 直接消费 route target”护栏，因此 `featureRouterImportAllowlist` 再从实例目录这一组净收掉 2 条。
  - `2026-05-29` teacher dashboard route target 进展：`useDashboardPage.ts` 已去掉 `vue-router`，新增 `teacherDashboardRoutes.ts` 统一提供教师概览“班级管理” route target contract；`TeacherDashboardPage.vue` 现在直接通过 `AppRouteLink` 渲染这条入口，教师概览的数据加载、错误态和 retry owner 保持不变。`TeacherDashboard.test.ts` 也已补上“page model 不再 import vue-router、page shell 直接消费 AppRouteLink、teacher/admin 角色命中正确班级管理路由”的护栏，因此 `featureRouterImportAllowlist` 再收掉 `useDashboardPage.ts` 这一条。
  - `2026-05-29` skill profile route target 进展：`useSkillProfilePage.ts` 已去掉 `vue-router`，新增 `skillProfileRoutes.ts` 统一提供“去做题”和题目详情的 route target contract；`SkillProfileWorkspaceShell.vue` 现在直接通过 `AppRouteLink` 渲染 CTA 与推荐题目跳转，六维画像数据加载、推荐数据加载、教师学员选择和刷新 owner 保持不变。`SkillProfile.test.ts` 也已补上“page model 不再 import vue-router、workspace shell 直接消费 AppRouteLink、做题与推荐题目命中正确路由”的护栏，因此 `featureRouterImportAllowlist` 再收掉 `useSkillProfilePage.ts` 这一条。
  - `2026-05-29` student review archive route target 进展：`useStudentReviewArchivePage.ts` 已去掉 `vue-router`，新增 `studentReviewArchiveRoutes.ts` 统一提供角色感知的“返回学生列表 / 返回学员分析” route target contract；teacher / platform 两个 review archive route view 现在通过 route props 把 `className / studentId` 传给 page model，`ReviewArchiveHero.vue` 则直接通过 `AppRouteLink` 渲染这两条导航，导出、轮询、下载和错误提示 owner 保持不变。`TeacherStudentReviewArchive.test.ts` 也已补上“page model 不再 import vue-router、hero 直接消费 AppRouteLink、teacher/admin 角色命中正确路由”的护栏，因此 `featureRouterImportAllowlist` 再收掉 `useStudentReviewArchivePage.ts` 这一条。

- [x] P1：继续拆 contest / AWD 线上的超大组件壳，优先看 `ContestAwdConfigWorkspaceShell.vue`、`ContestChallengeEditorDialog.vue`、`AWDChallengeLibraryPage.vue`
  - 依据：这三者当前约 `1009` / `899` / `896` 行，是现阶段最肥的一批前端组件壳。
  - 收益：继续收口 `TD-1`，减少单文件模板/样式/局部状态混写。
  - 风险：中高。比赛与 AWD 页面交互密度高，切片要尽量按稳定展示块或编辑分区拆。
  - `2026-05-27` 进展：`AWDChallengeLibraryPage.vue` 已拆成 `AwdChallengeWorkspaceHeader`、`AwdChallengeLibrarySection`、`AwdChallengeImportSection`，共享 page surface 只保留 mode 与 props / emits owner；`ContestChallengeEditorDialog.vue` 已进一步拆成 `ContestAwdChallengeSelectorSection` 与 `ContestChallengeSettingsSection`，父对话框只保留 form / validation / submit / selection owner。当前这条 P1 的剩余重点已经进一步收敛到 `ContestAwdConfigWorkspaceShell.vue`。
  - `2026-05-27` AWD config 进展：`ContestAwdConfigWorkspaceShell.vue` 已继续把 `Checker Parameters` 画布拆成 `ContestAwdCheckerConfigSection` 与按 checker type 划分的字段子组件，父壳只保留服务选择、draft、字段错误、保存、预览和 checker type owner；当前这条 P1 在 touched surface 上已不再由单一 1000 行壳体混放四种 checker 模板。
  - `2026-05-27` AWD config feature ui 进展：`ContestAwdConfigWorkspaceShell.vue` 已迁入 `features/contest-awd-config/ui`，`ContestAwdConfig` route 改为直接从 `features/contest-awd-config` public API 组合 page model 与 workspace shell；当前这条 P1 在 page-sized shell 落位层面的剩余重点已从 AWD 配置页移除。
  - `2026-05-28` AWD config panel cluster feature UI 进展：`ContestAwdConfigTopbar.vue`、`ContestAwdConfigFooter.vue`、`ContestAwdDebugStation.vue`、`ContestAwdEditorHeader.vue`、`ContestAwdScoreWeights.vue`、`ContestAwdServiceDirectory.vue`、`ContestAwdCheckerConfigSection.vue`，以及其依赖的 `ContestAwdHttpStandardFields.vue`、`ContestAwdLegacyProbeFields.vue`、`ContestAwdScriptCheckerFields.vue`、`ContestAwdTcpStandardFields.vue` 与 `contestAwdConfigTypes.ts` 已整体迁入 `features/contest-awd-config/ui`；`ContestAwdConfigWorkspaceShell.vue` 改为 feature 内部相对 import，AWD 配置页的主要编辑 UI cluster 不再散落在旧 `components/platform/contest/*` 路径。
  - `2026-05-28` contest edit follow-up 进展：`ContestChallengeEditorDialog.vue` 在完成 section 拆分后，也已进一步迁入 `features/contest-workbench/ui`，题目编排对话框不再滞留在旧 contest 组件目录。当前这条 P1 的下一批更合适目标已经转到 `AWDChallengeConfigPanel.vue` 这类仍留在 `components/platform/contest/*` 的大尺寸 surface。
  - `2026-05-28` 结项说明：这条 `P1` 原始点名的 3 个目标都已完成拆分与 feature 落位；剩余 contest / AWD 大组件仍然存在，但它们大多已经处在正确的 feature owner 之下，问题从“legacy shell 落位错误”转成“feature 内部仍有超大组件需要继续拆分”。


- [x] P2：继续拆其它 feature 内残余的大组件，优先看 `AWDRoundInspector.vue`
  - 依据：`ClassReportExportDialog.vue`、`ChallengeWriteupEditorPage.vue`、`ChallengeWriteupManagePanel.vue` 这一批 feature 内超大 surface 已完成父壳收口；当前这条线更集中的剩余目标，已经收敛到 `AWDRoundInspector.vue` 这类仍同时承接 workflow owner、稳定 section 和大段样式的 feature 内大组件。
  - 收益：继续降低 feature 内部“dialog shell / local sync watch / 稳定展示 section / 样式大段混写”的回归风险，也能避免新的 feature owner 再长回 page-sized SFC。
  - 风险：中。这里多数 surface 的 owner 已经正确，后续应优先沿 feature 内部局部职责拆分推进，而不是再次引入跨层迁移。
  - `2026-05-28` class report export dialog 进展：`ClassReportExportDialog.vue` 已收口为唯一 `AdminSurfaceModal` shell、`modelValue` contract 与 default context sync watch owner；当前教师上下文 / 导出设置 + preview snapshot、preview 展示区、latest task + guide rail 已分别下沉到 `ClassReportExportContextSection.vue`、`ClassReportExportPreviewSection.vue`、`ClassReportExportTaskRail.vue`，样式迁到 `classReportExportDialog.css`，raw-source 护栏同步改为聚合源码。父组件从约 `880` 行降到 `169` 行，当前这一条的剩余重点开始转向 `challenge-writeup-editor` 和 `awd-inspector` 这类仍未进一步拆开的 feature 内大 surface，而不再是 class report dialog 父壳继续混写 section 与样式。
  - `2026-05-28` challenge writeup editor page 进展：`ChallengeWriteupEditorPage.vue` 已收口为唯一 page shell、`embedded/back` contract 与 `useChallengeWriteupEditorPage()` workflow wiring owner；编辑器表单区、已保存版本 snapshot、题目信息 rail 已分别下沉到 `ChallengeWriteupEditorFormSection.vue`、`ChallengeWriteupSnapshotSection.vue`、`ChallengeWriteupChallengeRail.vue`，样式迁到 `challengeWriteupEditorPage.css`，题解编辑页相关 raw-source 护栏同步改为聚合源码。父组件从约 `670` 行降到 `148` 行，当前这一条的剩余重点进一步收敛到 `ChallengeWriteupManagePanel.vue` 和 `AWDRoundInspector.vue`，而不再是题解编辑页父壳继续混写 form / snapshot / rail 与 page 样式。
  - `2026-05-28` challenge writeup manage panel 进展：`ChallengeWriteupManagePanel.vue` 已收口为唯一 `challengeId/challengeTitle` contract、`openWriteup` emit、`useChallengeWriteupManagement()` workflow wiring 与 delete / action-menu owner；目录页头、summary strip、directory section、directory row 已分别下沉到 `ChallengeWriteupManageHeader.vue`、`ChallengeWriteupSummaryStrip.vue`、`ChallengeWriteupDirectorySection.vue`、`ChallengeWriteupDirectoryRow.vue`，样式迁到 `challengeWriteupManagePanel.css`，相关 raw-source 护栏同步改为聚合源码。父组件从约 `590` 行降到 `83` 行，当前这一条的剩余重点已进一步从题解目录收敛到 `AWDRoundInspector.vue` 和其它仍在 feature 内混写 workflow owner / section / CSS 的 surface。
  - `2026-05-28` awd round inspector 进展：`AWDRoundInspector.vue` 已收口为唯一 props / emits / slots contract、`activeSubTab`、`useAwdInspector*` workflow wiring、导出与 traffic forwarding owner；stats HUD 与 tabbed canvas workspace 已分别下沉到 `AWDInspectorStatsHud.vue`、`AWDInspectorCanvasWorkspace.vue`，样式迁到 `awdRoundInspector.css`，AWD inspector 相关 raw-source 护栏同步改为聚合源码。父组件从约 `545` 行降到 `190` 行，原先点名的这一组 feature 内大 surface 已全部完成父壳收口；当前 residual 重点开始并回上一条 contest / AWD feature 内残余超大 surface 与更深层 workflow handler 清理，而不再停留在这组父壳模板 / CSS 混写。


- [x] P2：把请求层错误导航 owner 继续收回页面 / feature owner，避免 `request.ts` 直接替页面决定可恢复错误的跳转
  - 依据：架构 review 仍把这条列为当前 P1/P2 级结构问题。
  - 收益：失败态、重试、草稿保留和页面内恢复体验会更一致，也更容易测试。
  - 风险：中高。会触及全局请求策略和多个页面的失败路径，需要先定义“哪些状态码是全局错误，哪些必须局部恢复”。
  - `2026-05-28` 结项说明：`request.ts` 已改为只负责 transport error normalization，不再直接 `logout + redirect`；HTTP `401`、WebSocket auth close、Vue runtime error、router runtime error 统一收口到 `runtime/globalErrorRuntime.ts`。可恢复的 `429 / 5xx / 网络错误 / 业务错误` 继续只返回 `ApiError`，由页面 / feature owner 自己决定 toast、inline fallback、retry 和 draft 保留。

- [x] P2：补图片管理页重复提交 owner 收口
  - 依据：前端主索引仍单独提到 `duplicateActionGuardAudit.test.ts` 暴露的图片管理页重复提交缺口。
  - 收益：问题集中、收益直接，能补掉一个实际交互安全缺口。
  - 风险：低到中。范围小，但需要确认按钮态、请求态和错误回退的 owner 不再重复分散。
  - `2026-05-28` 结项说明：`useImageManageMutations.ts` 现在同时持有创建与删除动作的本地 in-flight guard；删除镜像会在确认弹窗阶段和实际删除请求阶段短路同一条记录的重复点击，`ImageDirectoryPanel.vue` 也开始显式消费删除中的状态并禁用对应按钮。`duplicateActionGuardAudit.test.ts` 与 `ImageManage.test.ts` 已补齐这条交互护栏。

- [ ] P3：确定前端性能监控接入方案
  - 依据：主索引里 `TD-3` 仍为未完成项。
  - 收益：能把目前零散的性能体感问题转成可观测指标。
  - 风险：中。不是实现难，而是要先明确指标、隐私边界、上报端点和生产开关。

- [ ] P3：确定 i18n 是否进入当前产品路线
  - 依据：主索引里 `TD-4` 仍为未完成项。
  - 收益：如果产品后续确定多语言，这条需要尽早前置；如果短期不做，可以明确降级为“暂不推进”。
  - 风险：低。当前主要是产品决策问题，不是代码实现问题。

## Notes

- 当前 `oversized route view allowlist` 已经清空，所以“还有没有超大页面”这个问题，答案更准确地说是：
  - route view 层的大页债已大幅收口
  - 但 component / workspace shell / layout 层的大组件债仍然明显存在
- 后续如果按这份清单推进，建议继续沿“一个页面或一个组件族一个切片”的方式做，不要把 `TD-1` 和跨模块结构债混在同一次提交里。
