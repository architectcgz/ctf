> 状态：Current
> 事实源：student dashboard page-sized UI 迁移边界
> 替代：无

# Student Dashboard Feature UI Migration Plan

## 目标

- 把 `StudentCategoryProgressPage.vue`、`StudentDifficultyPage.vue`、`StudentOverviewPage.vue`、`StudentRecommendationPage.vue` 和 `dashboardPanelRegistry.ts` 从 `components/dashboard/student` 迁到 `features/student-dashboard/ui`。
- 让 `DashboardView.vue` 改为只从 `features/student-dashboard` public API 读取 page model 与 panel registry。

## 非目标

- 本轮不改 `useStudentDashboardPage.ts`、`useStudentDashboardData.ts` 的数据 owner。
- 本轮不迁 `StudentTimelinePage.vue`，因为它仍被 `StudentInsightPanel.vue` 直接复用。
- 本轮不重做学生仪表盘视觉样式或 tab 交互。

## 输入依据

- `code/frontend/src/views/dashboard/DashboardView.vue`
- `code/frontend/src/components/dashboard/student/dashboardPanelRegistry.ts`
- `code/frontend/src/components/dashboard/student/StudentCategoryProgressPage.vue`
- `code/frontend/src/components/dashboard/student/StudentDifficultyPage.vue`
- `code/frontend/src/components/dashboard/student/StudentOverviewPage.vue`
- `code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue`
- `code/frontend/src/components/dashboard/student/StudentTimelinePage.vue`
- `code/frontend/src/features/student-dashboard/model/useStudentDashboardPage.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 这四个 page-sized panel 只服务学生仪表盘 route，本质上属于 `student-dashboard` feature UI。
- `DashboardView.vue` 已经不是业务 owner，只是 route shell + tab 装配，适合继续把 panel registry 也收回 feature。
- `StudentTimelinePage.vue` 仍被 teacher 学员洞察直接消费，本轮暂不和四个 student-only panel 一起迁。

## 设计边界

### `features/student-dashboard` 本轮负责

- 对外暴露 `useStudentDashboardPage`
- 对外暴露 `resolveDashboardPanelComponent`
- 承接 4 个 student-only page-sized UI

### `views/dashboard/DashboardView.vue` 继续负责

- tab shell 渲染
- route level error / loading / panel mount 装配
- 通过 public API 读取 page model 和 panel registry

### `components/dashboard/student` 本轮保留

- `StudentTimelinePage.vue`
- `StudentOverviewStyleEditorial.vue`
- `StudentOverviewVariantSwitcher.vue`，但改成不再依赖旧 `StudentOverviewPage.vue` 路径

## 任务切片

### Slice 1：迁移 feature-owned UI

- 目标：
  - 新增 `features/student-dashboard/ui/*`
  - `DashboardView.vue` 从 feature public API 读取 panel registry
- 验证：
  - `npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/views/__tests__/studentOverviewEntrypoint.test.ts`
- Review focus：
  - page model owner 是否仍留在 feature model
  - DashboardView 是否没有重新耦合旧 `components/dashboard/student/dashboardPanelRegistry.ts`

### Slice 2：同步 allowlist 与 raw-source 护栏

- 目标：
  - 更新 `legacyComponentPageAllowlist`
  - 更新 student dashboard 相关 raw-source 测试到新路径
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Review focus：
  - 只移走 4 个 student-only panel，不误伤仍需共享的 `StudentTimelinePage.vue`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/views/__tests__/studentOverviewEntrypoint.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## 残余风险

- `StudentTimelinePage.vue` 还会继续停留在 `components/dashboard/student`，后续仍需要单独收口它与 teacher 学员洞察之间的共享 panel owner。
