# Reuse Decision

## Change type
frontend refactor / entity presentation owner strengthening

## Existing code searched
- `code/frontend/src/features/student-dashboard/model/useStudentDashboardData.ts`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveHero.vue`
- `code/frontend/src/features/teaching/student-review-archive/model/useStudentReviewArchive.ts`
- `code/frontend/src/features/teacher/dashboard/model/teacherDashboardOverviewBuilders.ts`
- `code/frontend/src/features/teacher/dashboard/model/teacherDashboardInsightBuilders.ts`
- `code/frontend/src/features/teacher/student-management/ui/StudentManagementPage.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsDirectoryPanel.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassInsightsPanel.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassReviewPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/InterventionPanel.vue`
- `code/frontend/src/features/platform/user-management/ui/UserGovernanceDetailModal.vue`
- `code/frontend/src/features/skill-profile/ui/SkillProfileWorkspaceShell.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherInterventionPanelLayout.test.ts`
- `code/frontend/src/pages/platform/__tests__/UserManage.test.ts`
- `code/frontend/src/pages/profile/__tests__/SkillProfile.test.ts`
- `code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts`
- `code/frontend/src/pages/review-archive/__tests__/TeacherStudentReviewArchive.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentManagement.test.ts`
- 前端 feature-sliced architecture 迁移台账文档

## Similar implementations found
- `entities/notification` 已经承接通知对象的稳定类型、已读态和 accent 展示规则。
- `entities/contest` 已经承接 contest 状态、模式、CTA、accent 和 badge class 等稳定展示规则。
- `entities/team` 已经承接队长关系、成员数、邀请码文案和成员展示项构建。
- 当前 `user` 还没有实体层入口，`name || username`、`@username`、`未设置姓名`、`选手` 这类稳定展示兜底仍散在 feature model 与 widget / feature ui 中。

## Decision
refactor_existing

## Reason
当前最小正确切片不是统一所有用户管理 UI，而是先把 `user` 对象的稳定显示规则收口：

- `useStudentDashboardData.ts` 仍在本地维护 `name || username || '选手'`
- `ReviewArchiveHero.vue` 仍直接拼 `archive.student.name || archive.student.username` 和 `@username`
- `useStudentReviewArchive.ts` 仍直接用 `name || username` 写 breadcrumb 详情标题
- 教师仪表盘 builders 与学生目录表格仍各自维护 `name || username` 或 `未设置姓名`
- 班级学生目录 / Top 学生 / 教学复盘 chips / 优先介入列表仍在 feature ui 内联 `name || username`
- 平台用户详情和能力画像教师学员选择仍各自维护用户展示 fallback

本轮最小正确改动是：

- 新建 `entities/user`，承接显示名、用户名 handle 和带 fallback 的稳定展示兜底
- 让 student dashboard、review archive、teacher dashboard 和 teacher student management 直接消费实体层公共入口
- 继续让 class students、student analysis review、platform user detail 和 skill profile teacher select 复用同一个实体展示 owner
- 用测试锁住 `name || username` 与 `@username` owner 不再回流到 feature model / widget / feature ui

本轮不做：

- 不改登录、权限、角色跳转、头像、用户状态、管理员用户治理 workflow
- 不改 shared topnav，因为 `shared` 不能反向依赖 `entities`
- 不处理实例、通知、角色 chip、审核人等其他对象的展示 owner

## Files to modify
- `.harness/reuse-decisions/user-entity-presentation-owner.md`
- `docs/plan/impl-plan/2026-05-31-user-entity-presentation-owner-plan.md`
- `code/frontend/src/entities/user/index.ts`
- `code/frontend/src/entities/user/model/index.ts`
- `code/frontend/src/entities/user/model/presentation.ts`
- `code/frontend/src/entities/user/model/presentation.test.ts`
- `code/frontend/src/features/student-dashboard/model/useStudentDashboardData.ts`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveHero.vue`
- `code/frontend/src/features/teaching/student-review-archive/model/useStudentReviewArchive.ts`
- `code/frontend/src/features/teacher/dashboard/model/teacherDashboardOverviewBuilders.ts`
- `code/frontend/src/features/teacher/dashboard/model/teacherDashboardInsightBuilders.ts`
- `code/frontend/src/features/teacher/student-management/ui/StudentManagementPage.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsDirectoryPanel.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassInsightsPanel.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassReviewPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/InterventionPanel.vue`
- `code/frontend/src/features/platform/user-management/ui/UserGovernanceDetailModal.vue`
- `code/frontend/src/features/skill-profile/ui/SkillProfileWorkspaceShell.vue`
- `code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts`
- `code/frontend/src/pages/review-archive/__tests__/TeacherStudentReviewArchive.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentManagement.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherInterventionPanelLayout.test.ts`
- `code/frontend/src/pages/platform/__tests__/UserManage.test.ts`
- `code/frontend/src/pages/profile/__tests__/SkillProfile.test.ts`

## After implementation
- `entities/user` 会成为这批 student / teacher surface 的显示名、用户名 handle 和稳定 fallback 展示 owner。
- student dashboard、review archive、teacher dashboard 和 teacher student management 不再本地维护 `name || username` 和 `@username`。
- class students、student analysis review、platform user detail 和 skill profile teacher select 也会停止散落 `name || username`。
- `user` 仍有后续消费面待继续收口，完成这一批后再评估剩余零散展示位还是切到 `instance`。
