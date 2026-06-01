> 状态：Current
> 事实源：student dashboard page model、dashboard route view、前端架构 allowlist
> 替代：无

# Student Dashboard Route Owner Cleanup Plan

## 目标

- 收掉 `features/student-dashboard/model/useStudentDashboardPage.ts -> vue-router`

## 非目标

- 不重做 student dashboard 各 panel 的视觉结构。
- 不把 dashboard panel 的按钮 contract 全量改成 `AppRouteLink`。
- 不改 `useStudentDashboardData.ts` 当前的数据加载 owner。

## 输入依据

- `useStudentDashboardPage.ts`
- `useStudentDashboardData.ts`
- `DashboardView.vue`
- `DashboardView.test.ts`
- `useStudentDashboardPageBoundary.test.ts`
- `architectureAllowlist.ts`

## 当前结论

- `panel` query tab sync 属于 student dashboard page 自己的能力 owner，继续由 page model 负责更合理。
- 题库 / 分类 / 难度 / 画像 / 题目详情的 5 条导航都属于薄导航，适合先收口成本地 route target contract。
- teacher/admin 的 role redirect 也属于薄 route target，不需要继续让 page model 直接碰 `vue-router`。
- 当前 dashboard panel 的交互以按钮回调为主，本轮不把整套 panel surface 硬翻成 link contract，避免引入更重的 UI 中间态。

## 设计边界

### student dashboard page model 本轮负责

- 保留 dashboard 数据加载、query tab owner、panel binding 组装和 mounted 时机
- 不再直接 import `vue-router`
- 通过本地 route builder + 共享 navigation transport 触发薄导航与角色 redirect

### shared transport 本轮负责

- 提供 `push()` / `replace()` transport
- 不承接 dashboard-specific query policy、redirect 判定或 panel binding 组装

### student dashboard routes 本轮负责

- 统一描述 challenge list、skill profile、challenge detail、teacher/admin redirect 的 route target
- 不承接 mounted lifecycle 或 auth role 判定

## 任务切片

- [ ] Slice 1：route target 与 transport 收口
  - 目标：
    - 新增 student dashboard route builder
    - 新增共享 navigation transport
  - 验证：
    - `cd code/frontend && npm run test:run -- src/features/student-dashboard/model/useStudentDashboardPageBoundary.test.ts`

- [ ] Slice 2：page model 去 router
  - 目标：
    - `useStudentDashboardPage.ts` 去掉 `vue-router`
    - query tab 改回直接消费 `useRouteQueryTabs()` 内部 route owner
    - 薄导航与 redirect 改成 route target + transport
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/features/student-dashboard/model/useStudentDashboardPageBoundary.test.ts`

- [ ] Slice 3：allowlist / backlog / review 收尾
  - 目标：
    - 更新 allowlist、todo、review
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/features/student-dashboard/model/useStudentDashboardPageBoundary.test.ts src/__tests__/architectureBoundaries.test.ts`
    - `cd code/frontend && npm run typecheck`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/features/student-dashboard/model/useStudentDashboardPageBoundary.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/student-dashboard-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-student-dashboard-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-student-dashboard-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/composables/routeNavigationTransport.ts code/frontend/src/features/student-dashboard/model/studentDashboardRoutes.ts code/frontend/src/features/student-dashboard/model/useStudentDashboardPage.ts code/frontend/src/features/student-dashboard/model/useStudentDashboardPageBoundary.test.ts code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 新增的 shared navigation transport 必须足够薄，不能顺手把 role 判定或 dashboard-specific route policy 平移过去。
- 本轮保留 panel button callback contract，如果后续 student dashboard 要继续走 link-first surface，再另开一刀处理，不在本轮和 allowlist cleanup 混做。
