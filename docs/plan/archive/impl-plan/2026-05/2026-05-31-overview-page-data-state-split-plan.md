# overview page data state split 计划

## Objective

- 在 `teacher/dashboard` 内新增 `useTeacherOverviewData`，承接教师概览的数据 owner。
- 在 `platform/overview` 内新增 `usePlatformOverviewData`，承接平台概览的数据 owner。
- 让两个 page model 都退回页面壳的 route / target 编排。

## Non-goals

- 不改 teacher / platform 概览页 UI。
- 不把 teacher / platform 概览页抽成 shared page owner。
- 不调整 dashboard / overview 展示组件、文案和布局。

## Source Inputs

- `code/frontend/src/features/teacher/dashboard/model/useDashboardPage.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts`
- `code/frontend/src/features/platform/overview/model/usePlatformOverviewPage.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformOverview.test.ts`
- `code/frontend/src/features/teacher/class-management/model/useTeacherClassDirectory.ts`
- `code/frontend/src/features/platform/class-management/model/usePlatformClassDirectory.ts`

## Plan Review Result

- 两个概览页都适合做 `page + data` 拆分。
- 数据 owner 负责请求、loading/error 和初始化。
- page model 只保留 route 派生与页面壳对外接口。

## Task Slices

### Slice 1: 新建 useTeacherOverviewData

- 目标：收口教师概览请求和错误状态。
- 风险：
  - 如果把班级管理 route 一起搬走，会重新模糊 page owner。

### Slice 2: 新建 usePlatformOverviewData

- 目标：收口平台概览请求、loading/error 状态。
- 风险：
  - 如果把审计或风险路由一起搬走，会重新模糊 page owner。

### Slice 3: 更新源码级和行为测试

- 目标：给两个新 data owner 补直测，并更新 teacher/platform 源码断言。
- 风险：
  - 不补失败态测试，后续概览异步状态还会回流到 page model。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision overview-page-data-state-split`
- `npm run test:run -- src/features/teacher/dashboard/model/useTeacherOverviewData.test.ts src/pages/teacher/__tests__/TeacherDashboard.test.ts src/features/platform/overview/model/usePlatformOverviewData.test.ts src/pages/platform/__tests__/PlatformOverview.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `npm run typecheck`
- `git diff --check`
- `bash scripts/check-reuse-first.sh`

## Review Focus

- `useTeacherOverviewData`、`usePlatformOverviewData` 是否只承接概览数据 owner。
- `useDashboardPage`、`usePlatformOverviewPage` 是否只剩 route / target 页面壳编排。
- route page 是否继续只做组合。

## Rollback / Recovery

- 如果新 data owner 的返回接口不顺手，可以调整字段组织，但概览数据加载 owner 仍必须留在新 composable。
