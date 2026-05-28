# 前端技术债优先级清单

- Project: `ctf 仓库根目录`
- Created: `2026-05-26T23:09+08:00`
- Status: `Open`

## Context

基于当前前端事实源和代码现状整理：

- `docs/reviews/frontend/README.md`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- 当前组件体量扫描结果

排序原则：

- 先做收益高、风险相对可控、能直接复用最近 page shell 拆分模式的项
- 再做收益高但跨模块耦合更重的结构债
- 性能监控、i18n 这类前置决策未定的项保留在后面单独跟踪

## Open Items

- [x] P1：继续拆 `PlatformOverviewPage.vue` 与 `TeacherInstanceManagementPage.vue`
  - 依据：`PlatformOverviewPage.vue` 当前约 `728` 行，`TeacherInstanceManagementPage.vue` 当前约 `597` 行，仍属于明显过宽的 legacy component page。
  - 收益：最容易复用最近 `TeacherDashboardPage` / `ClassStudentsPage` / `UserGovernancePage` 的 page owner 收口模式，能继续压缩 `legacyComponentPageAllowlist` 的负担。
  - 风险：中低。主要是展示区块切分和源码护栏适配，业务 owner 相对清楚。
  - `2026-05-26` 进展：已拆为 `PlatformOverviewHeroPanel` / `PlatformOverviewAlertsSection` / `PlatformOverviewHotspotsSection` 与 `TeacherInstanceHeroPanel` / `TeacherInstanceDirectorySection`，父页继续只保留 page shell、事件桥接和派发展示数据。

- [x] P1：处理 `platform-users` 过宽职责，把用户治理、班级/学生目录、实例管理继续拆回更清晰的 feature owner
  - 依据：原始架构 review 把 `platform-users` 视为过宽 bucket；当前代码事实里，这个桶已经拆成 `platform-user-management`、`platform-class-management`、`platform-student-management`、`platform-instance-management`。
  - 收益：完成后，admin 侧的用户治理、班级目录、学生目录、实例目录已经不再堆在同一个 feature 名下。
  - 风险：原始拆桶风险已消化；当前残留风险已转移到下面这条“admin / teacher 结构耦合”。
  - `2026-05-27` 进展：进一步补上 `api/admin/teaching.ts` owner，platform class / student / instance feature 与 admin 通知发布不再直接引用 `@/api/teaching`。

- [ ] P1：收口 admin / teacher 结构耦合，优先停止让 `/platform/*` 直接依赖 teacher 视图或 teacher 语义 owner
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

- [ ] P1：把应属于单一 feature 的 page-sized UI 从 `components/**` 继续收口到 `features/*/ui`
  - 依据：`docs/reviews/architecture/2026-05-24-frontend-architecture-review.md` 指出 allowlist 仍在冻结历史例外；当前题解管理三件套就是典型的 `components/*Page.vue -> @/features/*` 例外。
  - 收益：可以把“单一 feature 的 UI 壳”从 legacy component page 通道里迁走，减少 `componentFeatureImportAllowlist` 和 `legacyComponentPageAllowlist`，也给后续其它切片提供明确落点。
  - 风险：中。主要是 import 路径、raw-source 测试和 public API 更新，需要避免把 route owner 或 API owner 从 feature model 重新打散。
  - `2026-05-27` 进展：题解管理三件套 `ChallengeWriteupManagePanel`、`ChallengeWriteupEditorPage`、`ChallengeWriteupViewPage` 已完成迁入 `features/challenge-writeup-editor/ui`，前端架构文档同步补上 `feature-owned UI` 判定规则，题解这组对应的 `componentFeatureImportAllowlist` 与 `legacyComponentPageAllowlist` 已收掉；后续继续优先处理仍然直接依赖单一 feature model 的 legacy component page / panel。
  - `2026-05-27` platform overview 进展：`PlatformOverviewPage.vue` 已迁入 `features/platform-overview/ui`，`PlatformOverview` route 改为直接从 `features/platform-overview` public API 组合 page model 与 page shell；平台总览对应的一条 `componentFeatureImportAllowlist` 和一条 `legacyComponentPageAllowlist` 已收掉，下一批低风险候选可继续看 `TeacherDashboardPage.vue`。
  - `2026-05-27` teacher dashboard 进展：`TeacherDashboardPage.vue` 已迁入 `features/teacher-dashboard/ui`，`TeacherDashboard` route 改为直接从 `features/teacher-dashboard` public API 组合 page model 与 page shell；教师总览对应的一条 `componentFeatureImportAllowlist` 和一条 `legacyComponentPageAllowlist` 已收掉，下一批同模式候选可以继续看 `ContestOrchestrationPage.vue`。
  - `2026-05-27` contest orchestration 进展：`ContestOrchestrationPage.vue` 已迁入 `features/platform-contests/ui`，竞赛编辑页 / 运维页跳转 owner 已从 page shell 回收到 `useContestManagePage()`；竞赛目录 route 改为直接从 `features/platform-contests` public API 组合 page model 与 page shell，对应的一条 `componentFeatureImportAllowlist` 和一条 `legacyComponentPageAllowlist` 已收掉。
  - `2026-05-27` topology studio 进展：`ChallengeTopologyStudioPage.vue` 已迁入 `features/challenge-topology-studio/ui`，拓扑编辑 route 改为直接从 `features/challenge-topology-studio` public API 组合 page model 与 page shell；拓扑页对应的一条 `componentFeatureImportAllowlist` 和一条 `legacyComponentPageAllowlist` 已收掉，下一批同模式候选可继续看 `UserGovernancePage.vue`。
  - `2026-05-27` user governance 进展：`UserGovernancePage.vue` 已迁入 `features/platform-user-management/ui`，`panel=overview/import` 的 query owner 也一并从 page shell 收口到 `useUserGovernancePanelRoute()`；`UserManage` route 改为直接从 `features/platform-user-management` public API 组合 page model 与 page shell，用户治理页对应的一条 `legacyComponentPageAllowlist` 已收掉，并补上 feature model 的 router owner 白名单。
  - `2026-05-27` teacher instance 进展：`TeacherInstanceManagementPage.vue` 已迁入 `features/teacher-instances/ui`，`InstanceManagement` route 改为直接从 `features/teacher-instances` public API 组合 page model 与 page shell；教师实例页对应的一条 `legacyComponentPageAllowlist` 已收掉，教师端 raw-source surface 测试也已切到新 owner。
  - `2026-05-27` student management 进展：`StudentManagementPage.vue` 已迁入 `features/teacher-student-management/ui`，`TeacherStudentManagement` route 改为直接从 `features/teacher-student-management` public API 组合 page model 与 page shell，并继续在 route 壳里组合 `ClassReportExportDialog`；教师学生管理页对应的一条 `legacyComponentPageAllowlist` 已收掉，教师端 raw-source surface / header / pagination 测试也已切到新 owner。
  - `2026-05-27` remaining legacy pages 进展：`ClassManagementPage.vue`、`AWDChallengeLibraryPage.vue`、`ClassStudentsPage.vue`、`StudentAnalysisPage.vue` 已分别迁入 `features/teacher-class-management/ui`、`features/platform-awd-challenges/ui`、`features/class-students-workspace/ui`、`features/student-analysis-workspace/ui`；对应 teacher / platform route view 统一改为直接从 feature public API 组合 page model 与 page shell，`components/class-management/index.ts` 这个只服务旧 page path 的 barrel 也已退场。当前 `legacyComponentPageAllowlist` 已只剩 student dashboard 的 5 个页面，`AWDChallengeLibraryPage.vue` 对应的一条 `componentFeatureImportAllowlist` 也已收掉。
  - `2026-05-27` student dashboard 进展：`StudentCategoryProgressPage.vue`、`StudentDifficultyPage.vue`、`StudentOverviewPage.vue`、`StudentRecommendationPage.vue` 与 `dashboardPanelRegistry.ts` 已迁入 `features/student-dashboard/ui`，`DashboardView` 改为直接从 `features/student-dashboard` public API 读取 page model 与 panel registry；对应的 4 条 `legacyComponentPageAllowlist` 和 1 条 `componentFeatureImportAllowlist` 已收掉。当前这一条 P1 剩余的 student dashboard 存量主要只剩 `StudentTimelinePage.vue`，因为它仍被 teacher 学员洞察复用，需要下一刀按共享 panel owner 单独收口。
  - `2026-05-27` timeline panel 进展：`StudentTimelinePage.vue` 已收口为中立 `components/training/TrainingTimelinePanel.vue`，student dashboard registry 与 `StudentInsightPanel` 都改为消费共享训练时间线 panel；student dashboard 这条存量 page-sized 组件已清空，`legacyComponentPageAllowlist` 不再保留 dashboard 历史 page 例外。
  - `2026-05-27` platform AWD challenge feature UI 进展：`AWDChallengeEditorDialog.vue` 与 `AwdChallengeImportSection.vue` 已迁入 `features/platform-awd-challenges/ui`，`AWDChallengeLibrary.vue` route view 与 `AWDChallengeLibraryPage.vue` 不再直连旧 `components/platform/awd-service` 路径；对应的 2 条 `componentFeatureImportAllowlist` 已收掉。当前这一条后续更适合继续瞄准 `platform-challenge-detail`、`contest-workbench` 这类仍有单一 feature UI 滞留在 `components/**` 的残片。
  - `2026-05-27` platform challenge detail feature UI 进展：`AdminChallengeTopbarPanel.vue`、`AdminChallengeWorkspaceTabs.vue`、`AdminChallengeProfilePanel.vue` 已迁入 `features/platform-challenge-detail/ui`，`PlatformChallengeDetailWorkspace.vue` 改为通过 `features/platform-challenge-detail` public API 组合题目详情 topbar / tabs / profile；对应的 2 条 `componentFeatureImportAllowlist` 与 2 条 `widgetLegacyComponentImportAllowlist` 已收掉。当前这一条后续更适合继续瞄准 `contest-workbench`、`contest-announcements` 这类仍有单一 feature UI 滞留在 `components/**` 的残片。
  - `2026-05-27` contest announcements feature UI 进展：`ContestAnnouncementRealtimeBridge.vue` 与 `ContestAnnouncementManageDrawer.vue` 已迁入 `features/contest-announcements/ui`，竞赛详情页 / 竞赛目录页改为从 `features/contest-announcements` public API 取 bridge / drawer；“进入完整管理页”的导航 owner 也已从 drawer 收回 `useContestManagePage()`，对应的 2 条 `componentFeatureImportAllowlist` 已收掉。当前这一条后续更适合继续瞄准 `contest-workbench` 这类仍有单一 feature UI 滞留在 `components/**` 的残片。
  - `2026-05-27` contest workbench feature UI 进展：`ContestWorkbenchStageTabs.vue`、`ContestWorkbenchSummaryStrip.vue`、`ContestChallengeFilterStrip.vue`、`ContestChallengeSummaryStrip.vue`、`ContestChallengeOrchestrationPanel.vue` 已迁入 `features/contest-workbench/ui`，`ContestEditWorkspacePanel.vue` 也已按 route shell owner 迁入 `features/platform-contests/ui`；`ContestEdit.vue` 现在通过 `features/contest-workbench` 与 `features/platform-contests` public API 组合工作台，对应的 6 条 `componentFeatureImportAllowlist` 已收掉。
  - `2026-05-28` contest edit feature UI 进展：`ContestEditTopbarPanel.vue` 已迁入 `features/platform-contests/ui`，`ContestChallengeEditorDialog.vue`、`ContestAwdChallengeSelectorSection.vue`、`ContestChallengeSettingsSection.vue` 已整体迁入 `features/contest-workbench/ui`；`ContestEdit.vue` 与 `ContestChallengeOrchestrationPanel.vue` 不再回头引用旧 `components/platform/contest/*` 路径。当前这一条在 contest edit 线上的剩余低风险候选已经收敛到 `AWDChallengeConfigPanel.vue`。
  - `2026-05-28` AWD challenge config panel 进展：`AWDChallengeConfigPanel.vue` 已迁入 `features/platform-contests/ui`，`ContestEditWorkspacePanel.vue` 改为 feature 内部相对 import，`componentFeatureImportAllowlist` 中这条 `AWDChallengeConfigPanel.vue -> @/features/awd-inspector` 历史例外已收掉。当前这一条在 contest edit / platform contest 线上的低风险剩余候选，进一步收敛到 `PlatformContestFormPanel.vue`、`PlatformContestFormDialog.vue` 这类仍留在 `components/platform/contest/*` 的单一 feature UI。
  - `2026-05-28` platform contest form feature UI 进展：`PlatformContestFormPanel.vue` 与 `PlatformContestFormDialog.vue` 已迁入 `features/platform-contests/ui`，`ContestManage.vue` 改为通过 `features/platform-contests` public API 组合 dialog，`ContestOrchestrationPage.vue` 与 `ContestEditWorkspacePanel.vue` 也已切到 feature 内部 panel import；对应的 2 条 `componentFeatureImportAllowlist` 已收掉。当前这一条在 `platform-contests` 线上剩余的低风险候选，开始进一步收敛到 `PlatformContestTable.vue` 这类仍留在旧 contest 组件目录的展示 UI。
  - `2026-05-28` platform contest table feature UI 进展：`PlatformContestTable.vue` 已迁入 `features/platform-contests/ui`，`ContestOrchestrationPage.vue` 改为 feature 内部相对 import，相关 raw-source / typography / surface alignment 测试与组件声明也已切到新 owner。当前这一条在 `platform-contests` 线上的低风险遗留展示 UI 继续缩小，下一批更适合回到其它 feature 的单一 owner surface 或更大颗粒度的大组件壳体。
  - `2026-05-28` contest AWD preflight panel feature UI 进展：`ContestAwdPreflightPanel.vue` 已迁入 `features/platform-contests/ui`，`ContestEditWorkspacePanel.vue` 改为 feature 内部相对 import，相关 raw-source / theme token 护栏与组件声明也已切到新 owner。这一条在 `platform-contests` 线上的 contest edit 展示残片继续缩小；后续如果继续清 AWD readiness 相关 UI，应单独判断 `AWDReadinessChecklist.vue` / `AWDReadinessDecisionHUD.vue` 的 primitive owner，而不是顺手混入这刀。

- [ ] P1：继续拆 contest / AWD 线上的超大组件壳，优先看 `ContestAwdConfigWorkspaceShell.vue`、`ContestChallengeEditorDialog.vue`、`AWDChallengeLibraryPage.vue`
  - 依据：这三者当前约 `1009` / `899` / `896` 行，是现阶段最肥的一批前端组件壳。
  - 收益：继续收口 `TD-1`，减少单文件模板/样式/局部状态混写。
  - 风险：中高。比赛与 AWD 页面交互密度高，切片要尽量按稳定展示块或编辑分区拆。
  - `2026-05-27` 进展：`AWDChallengeLibraryPage.vue` 已拆成 `AwdChallengeWorkspaceHeader`、`AwdChallengeLibrarySection`、`AwdChallengeImportSection`，共享 page surface 只保留 mode 与 props / emits owner；`ContestChallengeEditorDialog.vue` 已进一步拆成 `ContestAwdChallengeSelectorSection` 与 `ContestChallengeSettingsSection`，父对话框只保留 form / validation / submit / selection owner。当前这条 P1 的剩余重点已经进一步收敛到 `ContestAwdConfigWorkspaceShell.vue`。
  - `2026-05-27` AWD config 进展：`ContestAwdConfigWorkspaceShell.vue` 已继续把 `Checker Parameters` 画布拆成 `ContestAwdCheckerConfigSection` 与按 checker type 划分的字段子组件，父壳只保留服务选择、draft、字段错误、保存、预览和 checker type owner；当前这条 P1 在 touched surface 上已不再由单一 1000 行壳体混放四种 checker 模板。
  - `2026-05-27` AWD config feature ui 进展：`ContestAwdConfigWorkspaceShell.vue` 已迁入 `features/contest-awd-config/ui`，`ContestAwdConfig` route 改为直接从 `features/contest-awd-config` public API 组合 page model 与 workspace shell；当前这条 P1 在 page-sized shell 落位层面的剩余重点已从 AWD 配置页移除。
  - `2026-05-28` contest edit follow-up 进展：`ContestChallengeEditorDialog.vue` 在完成 section 拆分后，也已进一步迁入 `features/contest-workbench/ui`，题目编排对话框不再滞留在旧 contest 组件目录。当前这条 P1 的下一批更合适目标已经转到 `AWDChallengeConfigPanel.vue` 这类仍留在 `components/platform/contest/*` 的大尺寸 surface。

- [ ] P2：收口布局层超大组件，优先看 `NotificationDrawer.vue`、`Sidebar.vue`、`TopNav.vue`
  - 依据：三者当前约 `1071` / `854` / `781` 行，已经超过普通布局组件可维护范围。
  - 收益：能降低全局导航和通知能力的维护摩擦，后续做主题、权限、消息交互时更安全。
  - 风险：中高。属于跨页面共享基础设施，任何切分都需要更谨慎的回归验证。
  - `2026-05-27` 进展：`NotificationDrawer.vue` 已按 layout owner 收口为“trigger / shell / filter state / dismiss owner”，并把 header、summary、tabs、body、footer 稳定视图区块拆到 `components/layout/notification-drawer/*`；相关 raw-source 与主题 token 护栏已切到聚合源码。当前这条 P2 的剩余重点已收敛到 `Sidebar.vue`、`TopNav.vue` 和通知抽屉更深层的行为清理。
  - `2026-05-27` Sidebar 进展：`Sidebar.vue` 已按“route/navigation owner 留父组件，移动壳 / 桌面壳 / nav tree 拆子组件”的方式收口到 `components/layout/sidebar/*`；raw-source 与主题 token 护栏已同步改为聚合源码。当前这条 P2 的剩余重点进一步收敛到 `TopNav.vue` 和 sidebar/nav 更深层的展示判定清理。
  - `2026-05-27` TopNav 进展：`TopNav.vue` 已按“route/theme/notification/logout owner 留父组件，移动 toggle / breadcrumbs / brand picker / notification trigger / user card 拆子组件”的方式收口到 `components/layout/topnav/*`；相关 raw-source 与主题 token 护栏已同步改为聚合源码。当前这条 P2 在大组件壳体层面的剩余重点开始转向通知抽屉、侧栏和 topnav 更深层的行为 owner 清理，而不是继续停留在单文件模板堆叠。

- [ ] P2：把请求层错误导航 owner 继续收回页面 / feature owner，避免 `request.ts` 直接替页面决定可恢复错误的跳转
  - 依据：架构 review 仍把这条列为当前 P1/P2 级结构问题。
  - 收益：失败态、重试、草稿保留和页面内恢复体验会更一致，也更容易测试。
  - 风险：中高。会触及全局请求策略和多个页面的失败路径，需要先定义“哪些状态码是全局错误，哪些必须局部恢复”。

- [ ] P2：补图片管理页重复提交 owner 收口
  - 依据：前端主索引仍单独提到 `duplicateActionGuardAudit.test.ts` 暴露的图片管理页重复提交缺口。
  - 收益：问题集中、收益直接，能补掉一个实际交互安全缺口。
  - 风险：低到中。范围小，但需要确认按钮态、请求态和错误回退的 owner 不再重复分散。

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
