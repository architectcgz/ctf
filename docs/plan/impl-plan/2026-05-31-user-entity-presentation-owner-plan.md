# 用户实体展示 owner 收口计划

## Objective

- 新建 `entities/user`，统一承接用户显示名、用户名 handle 和上下文 fallback 等稳定展示规则。
- 让 student dashboard、review archive、teacher dashboard 和 teacher student management 直接消费 `entities/user`，停止在各自文件里内联 `name || username` 与 `@username` 逻辑。

## Non-goals

- 不修改登录、权限、角色跳转、头像上传、用户状态或管理员用户治理流程。
- 不改 shared topnav，因为 `shared` 不能反向依赖 `entities`。
- 不处理用户状态 chip、角色文案、实例归属人和审核人文案等其他展示 owner。
- 不在本轮处理实例状态、实例剩余时间、实例访问地址和实例 badge 映射。

## Source Inputs

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
- `TODO/frontend-sliced-architecture.md`

## Brainstorming Conclusion

- 推荐方向：先用 `entities/user/model/presentation.ts` 收 `display name` / `username handle` / `fallback`，不先抽用户专属 UI 组件。
- 原因：当前散点主要是稳定文本派生，不需要先抽 card / row / badge 组件；先把 owner 收口到实体层就能让后续 `instance` 或 `platform user management` 继续复用。
- TDD：需要。这里会改动多个 feature / widget 的稳定展示兜底，适合先用实体层单测和源码级 guardrail 锁住边界。

## Plan Review Result

- `user` 实体层只负责“用户对象如何稳定显示”，不承接任何 route、permission、session 或 mutation owner。
- 本轮直接一起收 student dashboard、review archive、teacher dashboard 和 teacher student management 的重复 owner，不留“先建 helper、后续再迁消费面”的半收口状态。

## Task Slices

### Slice 1: 建立 user entity presentation owner

- 目标：新增 `entities/user` 公共入口，提供显示名和用户名 handle helper，并补单测。
- 变更面：
  - `code/frontend/src/entities/user/index.ts`
  - `code/frontend/src/entities/user/model/index.ts`
  - `code/frontend/src/entities/user/model/presentation.ts`
  - `code/frontend/src/entities/user/model/presentation.test.ts`
- 风险：
  - 如果把角色、状态或权限逻辑也混进实体层，会污染 owner。

### Slice 2: 迁移 student dashboard 与 review archive

- 目标：让 student dashboard、review archive hero 和 review archive breadcrumb 直接消费实体层用户展示 helper。
- 变更面：
  - `code/frontend/src/features/student-dashboard/model/useStudentDashboardData.ts`
  - `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveHero.vue`
  - `code/frontend/src/features/teaching/student-review-archive/model/useStudentReviewArchive.ts`
- 风险：
  - 如果 breadcrumb 标题继续自己兜底，owner 会重新分叉。

### Slice 3: 迁移 teacher dashboard 与 student management

- 目标：让 teacher dashboard builders 和 teacher student directory 停止本地维护显示名 fallback。
- 变更面：
  - `code/frontend/src/features/teacher/dashboard/model/teacherDashboardOverviewBuilders.ts`
  - `code/frontend/src/features/teacher/dashboard/model/teacherDashboardInsightBuilders.ts`
  - `code/frontend/src/features/teacher/student-management/ui/StudentManagementPage.vue`
- 风险：
  - 如果目录页仍保留 `未设置姓名` 和 `name || username` 私有逻辑，会留下双 owner。

### Slice 4: 锁住边界并更新迁移台账

- 目标：补源码级测试和迁移记录，明确 `user` 已经进入实体展示层。
- 变更面：
  - `code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts`
  - `code/frontend/src/pages/review-archive/__tests__/TeacherStudentReviewArchive.test.ts`
  - `code/frontend/src/pages/teacher/__tests__/TeacherStudentManagement.test.ts`
  - `TODO/frontend-sliced-architecture.md`
- 风险：
  - 如果只测渲染结果，不测 owner 回流，后续很容易退化。

### Slice 5: 收口剩余 feature-level 用户展示散点

- 目标：把班级学生目录、班级洞察、教学复盘、优先介入列表、平台用户详情和能力画像教师学员选择里的用户展示 fallback 收到 `entities/user`。
- 变更面：
  - `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsDirectoryPanel.vue`
  - `code/frontend/src/features/teaching/class-students-workspace/ui/ClassInsightsPanel.vue`
  - `code/frontend/src/features/teaching/class-students-workspace/ui/ClassReviewPanel.vue`
  - `code/frontend/src/features/teaching/student-analysis-review/ui/InterventionPanel.vue`
  - `code/frontend/src/features/platform/user-management/ui/UserGovernanceDetailModal.vue`
  - `code/frontend/src/features/skill-profile/ui/SkillProfileWorkspaceShell.vue`
- 风险：
  - 如果把“姓名列必须只看实名”和“展示名可回退用户名”混成一个 helper，会让目录类场景再次退化。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision user-entity-presentation-owner`
- `npm run test:run -- src/entities/user/model/presentation.test.ts src/pages/dashboard/__tests__/DashboardView.test.ts src/pages/review-archive/__tests__/TeacherStudentReviewArchive.test.ts src/pages/teacher/__tests__/TeacherStudentManagement.test.ts`
- `npm run test:run -- src/entities/user/model/presentation.test.ts src/pages/teacher/__tests__/TeacherClassStudents.test.ts src/pages/teacher/__tests__/teacherInterventionPanelLayout.test.ts src/pages/platform/__tests__/UserManage.test.ts src/pages/profile/__tests__/SkillProfile.test.ts`
- `git diff --check -- .harness/reuse-decisions/user-entity-presentation-owner.md docs/plan/impl-plan/2026-05-31-user-entity-presentation-owner-plan.md code/frontend/src/entities/user/index.ts code/frontend/src/entities/user/model/index.ts code/frontend/src/entities/user/model/presentation.ts code/frontend/src/entities/user/model/presentation.test.ts code/frontend/src/features/student-dashboard/model/useStudentDashboardData.ts code/frontend/src/widgets/review-archive-workspace/ReviewArchiveHero.vue code/frontend/src/features/teaching/student-review-archive/model/useStudentReviewArchive.ts code/frontend/src/features/teacher/dashboard/model/teacherDashboardOverviewBuilders.ts code/frontend/src/features/teacher/dashboard/model/teacherDashboardInsightBuilders.ts code/frontend/src/features/teacher/student-management/ui/StudentManagementPage.vue code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts code/frontend/src/pages/review-archive/__tests__/TeacherStudentReviewArchive.test.ts code/frontend/src/pages/teacher/__tests__/TeacherStudentManagement.test.ts TODO/frontend-sliced-architecture.md`
- `git diff --check -- .harness/reuse-decisions/user-entity-presentation-owner.md docs/plan/impl-plan/2026-05-31-user-entity-presentation-owner-plan.md code/frontend/src/entities/user/index.ts code/frontend/src/entities/user/model/index.ts code/frontend/src/entities/user/model/presentation.ts code/frontend/src/entities/user/model/presentation.test.ts code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsDirectoryPanel.vue code/frontend/src/features/teaching/class-students-workspace/ui/ClassInsightsPanel.vue code/frontend/src/features/teaching/class-students-workspace/ui/ClassReviewPanel.vue code/frontend/src/features/teaching/student-analysis-review/ui/InterventionPanel.vue code/frontend/src/features/platform/user-management/ui/UserGovernanceDetailModal.vue code/frontend/src/features/skill-profile/ui/SkillProfileWorkspaceShell.vue code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts code/frontend/src/pages/teacher/__tests__/teacherInterventionPanelLayout.test.ts code/frontend/src/pages/platform/__tests__/UserManage.test.ts code/frontend/src/pages/profile/__tests__/SkillProfile.test.ts TODO/frontend-sliced-architecture.md`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-reuse-first.sh`
- `bash scripts/check-workflow-complete.sh`

## Review Focus

- `entities/user` 是否只承接稳定用户展示规则，没有吸入权限、角色跳转或 session owner。
- student dashboard、review archive、teacher dashboard 和 teacher student management 是否已停止本地持有 `name || username` 与 `@username`。
- class students、intervention、platform user detail 和 skill profile teacher select 是否也已经停止本地持有 `name || username`。
- 教师和学生现有页面行为是否保持不变。

## Rollback / Recovery

- 如果 helper 形态让调用点可读性变差，可以回退具体函数命名或返回结构，但不能回退 owner 边界；用户显示规则仍必须停留在 `entities/user` 公共入口。
