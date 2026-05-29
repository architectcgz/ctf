> 状态：Current
> 事实源：class students page model、class workspace redirect helper、route transports、前端架构 allowlist
> 替代：无

# Class Students Route Owner Cleanup Plan

## 目标

- 收掉 `features/class-students-workspace/model/useClassStudentsPage.ts -> vue-router`

## 非目标

- 不改班级学生工作区的 tabs、报表导出、班级概览和学生目录 UI。
- 不重构 `useStudentListQuery()`、`useStudentFilters()` 或班级洞察 API。
- 不把 alias redirect 和 insight window query policy 下沉到 shared transport。

## 输入依据

- `useClassStudentsPage.ts`
- `useClassWorkspaceSection.ts`
- `routeQueryTransport.ts`
- `routeNavigationTransport.ts`
- `TeacherClassStudents.vue`
- `PlatformClassStudents.vue`
- `TeacherClassStudents.test.ts`
- `PlatformClassStudents.test.ts`
- `TeacherClassWorkspaceSection.test.ts`
- `architectureAllowlist.ts`

## 当前结论

- `class students` 这条的 router 触点同时包含 route `name / params / query` 读取、canonical redirect 和几条薄导航，重于单纯 query replace，但仍属于 page owner 自己的范围。
- 共享 transport 已经能承接 query/params 读取和 `push / replace`，本轮不需要再新建一个 class-students-specific route wrapper。
- 薄导航目标更适合落到本地 `classStudentsRoutes.ts`，而不是继续让 page model 直接拼 route name 和 params。

## 设计边界

### class students page model 本轮负责

- 保留 alias redirect owner、insight window query owner、班级工作区加载和错误处理
- 保留学生筛选、班级概览、报表导出和 stale request owner
- 不再直接 import `vue-router`

### shared transports 本轮负责

- 提供 route `name / params / query` 读取
- 提供 `push()` / `replace()` transport
- 不承接 class-students-specific redirect policy、时间窗口 schema 或 role 语义

### class students routes 本轮负责

- 统一描述班级管理、教学概览、学员分析的 route target
- 不承接 mounted lifecycle、筛选加载或 redirect 判定

## 任务切片

- [ ] Slice 1：page model 改用 shared route transports
  - 目标：
    - `useClassStudentsPage.ts` 去掉 `vue-router`
    - `className` params、时间窗口 query、canonical redirect 改为消费共享 transport
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherClassStudents.test.ts src/views/platform/__tests__/PlatformClassStudents.test.ts`

- [ ] Slice 2：抽本地 route target helper
  - 目标：
    - 新增 `classStudentsRoutes.ts`
    - 学员分析 / 班级管理 / 教学概览走 route target + navigation transport
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherClassStudents.test.ts src/views/platform/__tests__/PlatformClassStudents.test.ts src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts`

- [ ] Slice 3：allowlist / review / backlog 收尾
  - 目标：
    - 更新 allowlist、raw-source 护栏、todo、review
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherClassStudents.test.ts src/views/platform/__tests__/PlatformClassStudents.test.ts src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts src/__tests__/architectureBoundaries.test.ts`
    - `cd code/frontend && npm run typecheck`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherClassStudents.test.ts src/views/platform/__tests__/PlatformClassStudents.test.ts src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/class-students-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-class-students-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-class-students-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/composables/routeQueryTransport.ts code/frontend/src/features/class-students-workspace/model/classStudentsRoutes.ts code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `routeQueryTransport.ts` 如果暴露 `name`，也仍应只保持 route transport 语义；不能继续长成业务 helper。
- 这轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
