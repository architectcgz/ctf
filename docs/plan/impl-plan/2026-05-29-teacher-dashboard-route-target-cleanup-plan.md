> 状态：Current
> 事实源：teacher dashboard page model、teacher dashboard page shell、前端架构 allowlist
> 替代：无

# Teacher Dashboard Route Target Cleanup Plan

## 目标

- 把 `useDashboardPage.ts` 从单次 `router.push()` 收口成纯 route target contract。
- 保持教师概览的数据加载、错误态和 retry owner 不变，同时再清掉 1 条 `featureRouterImportAllowlist`。

## 非目标

- 不处理 `useStudentDashboardPage.ts` 的 tab query 与 challenge 跳转 owner。
- 不改教师概览 tab 切换、指标构建和 retry 流程。
- 不继续做 dashboard UI 拆分。

## 输入依据

- `code/frontend/src/features/teacher-dashboard/model/useDashboardPage.ts`
- `code/frontend/src/features/teacher-dashboard/ui/TeacherDashboardPage.vue`
- `code/frontend/src/views/teacher/TeacherDashboard.vue`
- `code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/components/navigation/AppRouteLink.vue`

## 当前结论

- `useDashboardPage.ts` 的 router 依赖只剩“进入班级管理”一处。
- 数据 owner、错误态和 retry 已经都在合适位置，不需要再移动。
- 这一条适合直接收口成 route target contract。

## 设计边界

### `teacherDashboardRoutes.ts` 本轮负责

- 生成教师 / 管理员班级管理页 route target

### `useDashboardPage()` 本轮负责

- 教师概览数据加载、错误态和 retry owner
- 暴露 `classManagementRoute`，不再直接导航

### `TeacherDashboardPage.vue` 本轮负责

- 通过共享 `AppRouteLink` 消费 route target
- 保持 retry 仍由 button + emit 驱动

## 任务切片

### Slice 1：page model 去 router 化

- 目标：
  - 新增 teacher dashboard route helper
  - `useDashboardPage.ts` 去掉 `vue-router`
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - allowlist 是否可净减少 1 条

### Slice 2：教师概览 UI 切到 `AppRouteLink`

- 目标：
  - `TeacherDashboardPage.vue` 的“班级管理”入口改成共享 `AppRouteLink`
  - `TeacherDashboard.vue` 继续只做组合
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherDashboard.test.ts`
- Review focus：
  - route target contract 是否清楚，没有把 retry owner 一起迁走

### Slice 3：allowlist / review / backlog 收尾

- 目标：
  - 更新 allowlist、review 和 backlog
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - `featureRouterImportAllowlist` 是否净减少 1 条

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherDashboard.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/teacher-dashboard-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-teacher-dashboard-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-teacher-dashboard-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/teacher-dashboard/model/teacherDashboardRoutes.ts code/frontend/src/features/teacher-dashboard/model/useDashboardPage.ts code/frontend/src/features/teacher-dashboard/ui/TeacherDashboardPage.vue code/frontend/src/views/teacher/TeacherDashboard.vue code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- student dashboard 仍保留 route/query owner，这轮不一并处理。
