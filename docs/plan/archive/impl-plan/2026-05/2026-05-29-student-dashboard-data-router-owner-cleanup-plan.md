> 状态：Current
> 事实源：student dashboard page / data owner、前端架构 allowlist、DashboardView 测试
> 替代：无

# Student Dashboard Data Router Owner Cleanup Plan

## 目标

- 把 `useStudentDashboardData.ts` 从 route-aware owner 收口回 dashboard data owner。
- 让 `useStudentDashboardPage.ts` 保留唯一 router owner，并承接 role redirect。
- 删除对应 `featureRouterImportAllowlist` 条目。

## 非目标

- 不重构 student dashboard 的 panel registry、展示结构或推荐逻辑。
- 不改 teacher/admin dashboard 的页面路由设计。
- 不处理 `featureRouterImportAllowlist` 其它 feature 条目。

## 输入依据

- `code/frontend/src/features/student-dashboard/model/useStudentDashboardData.ts`
- `code/frontend/src/features/student-dashboard/model/useStudentDashboardPage.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts`

## 当前结论

- `useStudentDashboardData.ts` 当前职责是数据加载与展示派生；teacher/admin redirect 混在这个 composable 中，会把 router owner 漂进 data 层。
- `useStudentDashboardPage.ts` 已经持有 `useRouter()` 和 query tab owner，继续作为唯一 route-aware owner 更合理。

## 设计边界

### `useStudentDashboardPage.ts` 本轮负责

- `useRouter()` 获取
- teacher/admin role redirect
- challenge / category / difficulty / skill profile 的导航动作
- route query tabs owner

### `useStudentDashboardData.ts` 本轮负责

- 根据当前用户角色暴露是否需要 redirect 的 signal
- 仪表盘数据加载
- 展示派生数据，如 highlights、counts、completion rate
- 加载失败时的本地错误态与回退数据

## 任务切片

### Slice 1：data owner 去掉 router 依赖

- 目标：
  - 从 `useStudentDashboardData.ts` 移除 `Router` 依赖与 `router.replace(...)`
  - 暴露 role redirect signal，保持 load 入口仍由 page owner 调用
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts`
- Review focus：
  - data 层是否已经不再 import `vue-router`
  - 非 student 角色是否不再继续触发 dashboard 数据请求

### Slice 2：page owner / allowlist / 护栏收尾

- 目标：
  - 由 `useStudentDashboardPage.ts` 接住 redirect signal 并执行跳转
  - 删除 allowlist 条目
  - 更新 raw-source 护栏与 backlog 记录
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/dashboard/__tests__/DashboardView.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - route owner 是否只回到 page，而没有漂到别的 data/helper
  - redirect 行为断言是否仍然成立

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/dashboard/__tests__/DashboardView.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/student-dashboard-data-router-owner-cleanup.md docs/plan/impl-plan/2026-05-29-student-dashboard-data-router-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-student-dashboard-data-router-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/student-dashboard/model/useStudentDashboardData.ts code/frontend/src/features/student-dashboard/model/useStudentDashboardPage.ts code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收口一条 `featureRouterImportAllowlist`，不代表剩余条目都不合理；仍需逐条判定。
- 这轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
