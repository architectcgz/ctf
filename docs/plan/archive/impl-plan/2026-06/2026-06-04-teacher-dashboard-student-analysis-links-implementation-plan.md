# Teacher Dashboard Student Analysis Links Implementation Plan

## Objective
- 给 `/academy/overview` 的 `学生洞察` 与 `教学复盘` 面板补“查看学生信息”入口。
- 学生粒度摘要先打开学生名单 modal，再从名单进入个人复盘。
- 班级粒度摘要直接进入 class review 页面。
- 保持 dashboard route owner 在 feature/model，UI panel 只触发名单查看动作或消费 class review route target。

## Non-goals
- 不改 overview API 契约。
- 不改 `趋势复盘`、`介入建议` 的点击行为。
- 不新增独立学生推荐/筛选接口。
- 不把摘要标题重新改回直接跳个人详情。

## Source inputs
- `code/frontend/src/features/teacher/dashboard/model/useDashboardPage.ts`
- `code/frontend/src/features/teacher/dashboard/model/useDashboardMetrics.ts`
- `code/frontend/src/features/teacher/dashboard/model/teacherDashboardInsightBuilders.ts`
- `code/frontend/src/features/teacher/dashboard/model/teacherDashboardOverviewBuilders.ts`
- `code/frontend/src/features/teacher/student-management/model/teacherStudentManagementRoutes.ts`
- `code/frontend/src/shared/ui/navigation/AppRouteLink.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts`

## Slice 1: row contract and student list mapping
- 在 dashboard route helper 内补学生详情 route target builder。
- 给 insight/review row contract 增加相关学生列表，再在 `useDashboardMetrics` 收口为 modal 名单项和 route target。
- 验证：
  - `pnpm test:run src/pages/teacher/__tests__/TeacherDashboard.test.ts`
  - `pnpm typecheck`
- Review focus：
  - route owner 是否仍留在 feature/model
  - DTO 到学生名单的映射是否只使用 overview 已有事实
  - 班级 / 风险组条目是否返回整组学生而不是单个代表样本

## Slice 2: modal and class-review flow rendering
- `TeacherDashboardPage.vue` 继续作为 owner，统一维护 modal 开关和当前学生名单。
- `TeacherDashboardStudentListDialog.vue` 渲染学生名单与个人复盘入口。
- `TeacherDashboardStudentInsightPanel.vue` / `TeacherDashboardReviewPanel.vue` 对学生粒度条目触发 modal，对班级粒度条目消费 class review route target。
- 验证：
  - `pnpm test:run src/pages/teacher/__tests__/TeacherDashboard.test.ts`
  - `pnpm typecheck`
- Review focus：
  - panel 是否只发出查看名单事件或消费 route target，不直接 import router
  - modal 内是否只在可定位学生时提供个人复盘入口
  - 班级粒度条目是否复用了现有 class review 页面
  - 无学生名单时是否保持静态展示而不是空 modal

## Expected files
- `code/frontend/src/features/teacher/dashboard/model/teacherDashboardRoutes.ts`
- `code/frontend/src/features/teacher/dashboard/model/teacherDashboardInsightBuilders.ts`
- `code/frontend/src/features/teacher/dashboard/model/teacherDashboardOverviewBuilders.ts`
- `code/frontend/src/features/teacher/dashboard/model/useDashboardMetrics.ts`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardStudentListDialog.vue`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardStudentInsightPanel.vue`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardReviewPanel.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts`

## Risks
- `focus_students` 是 overview 聚合里明确挑出来的风险学生样本，风险组 modal 以它为准。
- 班级粒度入口直接复用 class review 页面，避免在 overview DTO 不完整时前端拼凑“完整班级学生名单”。

## Rollback
- 回退 modal 名单编排和学生名单映射。
- 保持原有纯摘要展示，不影响 overview 数据加载与 tab 行为。
