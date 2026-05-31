# Reuse Decision

## Change type
frontend refactor / overview page data state split

## Existing code searched
- `code/frontend/src/features/teacher/dashboard/model/useDashboardPage.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts`
- `code/frontend/src/features/platform/overview/model/usePlatformOverviewPage.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformOverview.test.ts`
- `code/frontend/src/features/teacher/class-management/model/useTeacherClassDirectory.ts`
- `code/frontend/src/features/platform/class-management/model/usePlatformClassDirectory.ts`

## Similar implementations found
- `useDashboardPage.ts` 当前同时持有教师概览请求、错误状态和班级管理 route 派生。
- `usePlatformOverviewPage.ts` 当前同时持有平台概览请求、loading/error 状态和导航 route target。
- `useTeacherClassDirectory.ts`、`usePlatformClassDirectory.ts` 刚完成同类拆分，已经证明 page model 可以退回页面壳编排，目录/概览数据 owner 独立承接异步状态。

## Decision
refactor_existing

## Reason
当前最小正确切片是把两个概览页都拆成 `page + data` 结构：

- `useTeacherOverviewData`：承接教师概览请求和错误状态。
- `useDashboardPage`：保留班级管理 route 等页面壳编排。
- `usePlatformOverviewData`：承接平台概览请求和 loading/error 状态。
- `usePlatformOverviewPage`：保留审计日志与风险研判 route target。

这样可以：

- 去掉 page model 里混合的概览数据 owner
- 统一当前迁移路线，让 page model 只保留页面壳语义

本轮不做：

- 不改 teacher / platform 概览页 UI
- 不把 teacher / platform 概览抽成 shared page owner
- 不改 dashboard / overview 下游展示组件结构

## Files to modify
- `.harness/reuse-decisions/overview-page-data-state-split.md`
- `docs/plan/impl-plan/2026-05-31-overview-page-data-state-split-plan.md`
- `code/frontend/src/features/teacher/dashboard/model/useTeacherOverviewData.ts`
- `code/frontend/src/features/teacher/dashboard/model/useTeacherOverviewData.test.ts`
- `code/frontend/src/features/teacher/dashboard/model/useDashboardPage.ts`
- `code/frontend/src/features/teacher/dashboard/model/index.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts`
- `code/frontend/src/features/platform/overview/model/usePlatformOverviewData.ts`
- `code/frontend/src/features/platform/overview/model/usePlatformOverviewData.test.ts`
- `code/frontend/src/features/platform/overview/model/usePlatformOverviewPage.ts`
- `code/frontend/src/features/platform/overview/model/index.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformOverview.test.ts`

## After implementation
- 教师概览页的数据加载 owner 会集中到 `useTeacherOverviewData`。
- 平台概览页的数据加载 owner 会集中到 `usePlatformOverviewData`。
- 两个 page model 都只保留页面壳路由与 route target 编排。
