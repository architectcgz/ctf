# 前端统一间距体系（CTF）

> 状态：已落地到共享样式层（workspace/tab/surface/divider）
> 适用范围：`code/frontend/src` 全部路由页面与页面承载组件

---

## 1. 覆盖范围总览

本次按路由入口盘点了 **42 个前端路由页面**（来自 `router/index.ts`）：

- admin（12）
  - `@/views/admin/AdminDashboard.vue`
  - `@/views/admin/AuditLog.vue`
  - `@/views/admin/ChallengeDetail.vue`
  - `@/views/admin/ChallengeManage.vue`
  - `@/views/admin/ChallengePackageFormat.vue`
  - `@/views/admin/ChallengeTopologyStudio.vue`
  - `@/views/admin/ChallengeWriteup.vue`
  - `@/views/admin/CheatDetection.vue`
  - `@/views/admin/ContestManage.vue`
  - `@/views/admin/EnvironmentTemplateLibrary.vue`
  - `@/views/admin/ImageManage.vue`
  - `@/views/admin/UserManage.vue`
- teacher（8）
  - `@/views/teacher/ClassManagement.vue`
  - `@/views/teacher/InstanceManagement.vue`
  - `@/views/teacher/ReportExport.vue`
  - `@/views/teacher/TeacherClassStudents.vue`
  - `@/views/teacher/TeacherDashboard.vue`
  - `@/views/teacher/TeacherStudentAnalysis.vue`
  - `@/views/teacher/TeacherStudentManagement.vue`
  - `@/views/teacher/TeacherStudentReviewArchive.vue`
- errors（8）
  - `@/views/errors/BadGatewayView.vue`
  - `@/views/errors/ForbiddenView.vue`
  - `@/views/errors/GatewayTimeoutView.vue`
  - `@/views/errors/InternalServerErrorView.vue`
  - `@/views/errors/NotFoundView.vue`
  - `@/views/errors/ServiceUnavailableView.vue`
  - `@/views/errors/TooManyRequestsView.vue`
  - `@/views/errors/UnauthorizedView.vue`
- profile（3）
  - `@/views/profile/SecuritySettings.vue`
  - `@/views/profile/SkillProfile.vue`
  - `@/views/profile/UserProfile.vue`
- notifications（2）
  - `@/views/notifications/NotificationDetail.vue`
  - `@/views/notifications/NotificationList.vue`
- contests（2）
  - `@/views/contests/ContestDetail.vue`
  - `@/views/contests/ContestList.vue`
- challenges（2）
  - `@/views/challenges/ChallengeDetail.vue`
  - `@/views/challenges/ChallengeList.vue`
- auth（2）
  - `@/views/auth/LoginView.vue`
  - `@/views/auth/RegisterView.vue`
- 其他（3）
  - `@/views/dashboard/DashboardView.vue`
  - `@/views/instances/InstanceList.vue`
  - `@/views/scoreboard/ScoreboardView.vue`

同时盘点了 **16 个页面承载组件**（`components/**/**Page.vue`），这些组件承担 teacher/admin/student 的主体布局实现，已纳入同一间距体系。

---

## 2. 现状审计结论

对 `views + components` 的 class 间距工具类做统计后，发现：

- 高频值集中在 `px-4 / px-3 / py-3 / gap-3 / py-2 / gap-2 / space-y-2`，说明页面已自然聚集到少量主值。
- 仍有大量局部硬编码和页面私有微调（例如 `pt-*`、`mt-*` 单点修补），导致跨页节奏不稳定。
- `workspace-shell`、`top-tabs`、`teacher-surface`、`journal-divider` 这些共享骨架层之前存在固定数值，页面间很难统一调整。

结论：

- 问题不在“没有间距值”，而在“缺少统一 token + 语义层绑定 + 共享层收敛”。

---

## 3. 统一间距设计体系

### 3.1 基础刻度（Global Scale）

定义在 `theme.css`：

- `--space-0` 到 `--space-12`
- 采用 4px 基础网格，并保留历史兼容步长（如 `--space-2-5`=10px、`--space-5-5`=22px）

目标：

- 新样式优先只用该刻度；避免新增任意 px/rem 常量。

### 3.2 语义间距（Semantic Spacing）

定义在 `theme.css`，面向布局语义而非组件名：

- workspace 壳层
  - `--space-workspace-topbar-gap-x/y`
  - `--space-workspace-topbar-padding-top`
  - `--space-workspace-side-padding`
  - `--space-workspace-tabs-gap`
  - `--space-workspace-tabs-offset-top`
  - `--space-workspace-content-padding`
- 分区与分隔线
  - `--space-section-gap`
  - `--space-section-gap-compact`
  - `--space-divider-gap`
  - `--space-divider-gap-compact`

目标：

- 统一改“语义变量”，而不是全仓逐页改 `mt/pt/gap`。

### 3.3 共享层绑定规则

已绑定到以下共享样式文件：

- `workspace-shell.css`
  - 顶栏、tabs、content-pane 的 gap/padding 改为语义 token
- `page-tabs.css`
  - 通用 tab rail、tab heading/copy 的默认间距改为 token
- `teacher-surface.css`
  - board/section/filter/topbar/actions/summary 改为 token
- `journal-soft-surfaces.css`
  - note/eyebrow/button 的默认内边距与间距改为 token
- `journal-notes.css`
  - `journal-divider`、note helper/card padding 改为 token

---

## 4. 组件与页面使用约束

### 4.1 必须遵守

- 页面优先使用共享骨架（`workspace-shell` / `top-tabs` / `content-pane`）提供的间距。
- 新增间距优先使用 `var(--space-*)` 或语义 token。
- 页面局部需要微调时，先覆盖语义 token，再考虑局部类。

### 4.2 禁止事项

- 在页面 scoped CSS 中新增大量裸值（如 `padding: 22px 28px 0`）。
- 同一层级混用多套间距逻辑（例如同时写 `space-y-*` 和额外 hardcoded `margin-top` 链式覆盖）。
- 为了解决单页问题，新增只在一个页面可见的“伪全局常量”。

---

## 5. 落地与迁移策略

- 第一阶段（已完成）：统一共享层 token 化（本次提交）。
- 第二阶段（按页面批次）：
  - teacher 类页面先迁移（班级管理/学生管理/分析页）
  - admin 类目录页迁移（challenge/image/user/contest）
  - student 类仪表盘子页收敛
- 第三阶段：测试基线固化（style tests 覆盖 token 与共享层绑定）。

---

## 6. 验收清单

- [x] 有全局间距刻度 token。
- [x] 有语义间距 token。
- [x] 共享壳层样式已消费语义 token。
- [x] 通用分隔线样式可复用，且间距可 token 化。
- [x] 文档写明覆盖范围、规则和迁移策略。

