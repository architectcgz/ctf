# Reuse Decision

## Change type
+feature / api / handler / service / repository / component / readmodel

## Existing code searched
- code/frontend/src/views/teacher/TeacherDashboard.vue
- code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue
- code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue
- code/frontend/src/features/teacher-dashboard/model/useTeacherDashboardPage.ts
- code/frontend/src/features/teacher-dashboard/model/useTeacherDashboardMetrics.ts
- code/frontend/src/features/teacher-workspace/model/useTeacherWorkspace.ts
- code/frontend/src/features/teacher-student-analysis/model/useTeacherReviewWorkspace.ts
- code/frontend/src/features/teacher-class-workspace/model/useTeacherClassWorkspaceSection.ts
- code/frontend/src/api/teacher/classes.ts
- code/backend/internal/module/teaching_readmodel/api/http/handler.go
- code/backend/internal/module/teaching_readmodel/application/queries/service.go
- code/backend/internal/module/teaching_readmodel/infrastructure/repository.go
- code/backend/internal/module/teaching_readmodel/ports/query.go
- docs/architecture/features/教学复盘优化设计.md
- docs/architecture/features/教学复盘建议生成架构.md

## Similar implementations found
- `ClassStudentsPage.vue` 已经是班级详情工作区 owner，适合继续承载趋势、复盘、洞察、介入这些完整班级面板。
- `teaching_readmodel.QueryService` 已经是教师侧只读聚合 owner，班级摘要、趋势、复盘都在这里实现，新增教学概览聚合应继续落在这个模块，而不是在前端并发拼多个班级接口。
- `TeacherDashboardPage.vue` 当前虽然复用了 `TeacherClassTrendPanel`、`TeacherClassReviewPanel`、`TeacherInterventionPanel`，但这些组件的 contract 明确是班级详情级，不适合作为教学概览长期 owner。
- `useTeacherWorkspace.ts`、`useTeacherReviewWorkspace.ts`、`useTeacherClassWorkspaceSection.ts` 都是在 feature 内把页面 owner 继续收口到一个语义化 hook 名称。这次新增 `useTeacherOverviewPage.ts` / `useTeacherOverviewWorkspace.ts` 不是复制新逻辑，而是把新的 overview scope owner 和既有 `useTeacherDashboardPage`、`useTeacherDashboardMetrics` 之间建立稳定别名，避免路由页和页面组件继续暴露旧的 dashboard 命名。

## Decision
+refactor_existing

## Reason
教学概览和班级详情当前复用了同一批 class-level query 与 panel，导致 `/academy/overview` 实际上退化成“默认班级详情”。这次不新造第二套权限或评估规则模块，而是在既有 `teaching_readmodel` 与 `teacher-dashboard` feature 上做 owner 重划分：新增 overview 专属 read model 与前端 workspace，保留班级详情面板继续服务 `/academy/classes/:className`。这样可以直接修正当前总览/详情边界错误，并为后续教师多班级权限扩展保留接口空间。

## Files to modify
- .harness/reuse-decisions/teacher-overview-aggregate.md
- docs/architecture/features/教师教学概览聚合架构.md
- docs/architecture/features/专题架构索引.md
- docs/plan/impl-plan/2026-05-12-teacher-overview-aggregate-implementation-plan.md
- docs/architecture/backend/04-api-design.md
- code/backend/internal/dto/teacher.go
- code/backend/internal/app/router_routes.go
- code/backend/internal/module/teaching_readmodel/api/http/handler.go
- code/backend/internal/module/teaching_readmodel/application/queries/contracts.go
- code/backend/internal/module/teaching_readmodel/application/queries/service.go
- code/backend/internal/module/teaching_readmodel/application/queries/service_overview_test.go
- code/backend/internal/module/teaching_readmodel/ports/query.go
- code/backend/internal/module/teaching_readmodel/infrastructure/repository.go
- code/backend/internal/module/teaching_readmodel/runtime/module.go
- code/frontend/src/api/contracts.ts
- code/frontend/src/api/teacher/classes.ts
- code/frontend/src/api/teacher/index.ts
- code/frontend/src/features/teacher-dashboard/model/index.ts
- code/frontend/src/features/teacher-dashboard/model/teacherDashboardInsightBuilders.ts
- code/frontend/src/features/teacher-dashboard/model/teacherDashboardOverviewBuilders.ts
- code/frontend/src/features/teacher-dashboard/model/useTeacherDashboardMetrics.ts
- code/frontend/src/features/teacher-dashboard/model/useTeacherDashboardMetricsBoundary.test.ts
- code/frontend/src/features/teacher-dashboard/model/useTeacherDashboardPage.ts
- code/frontend/src/features/teacher-dashboard/model/useTeacherOverviewPage.ts
- code/frontend/src/features/teacher-dashboard/model/useTeacherOverviewWorkspace.ts
- code/frontend/src/views/teacher/TeacherDashboard.vue
- code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue
- code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts
- code/frontend/src/api/__tests__/teacher.test.ts

## After implementation
- 如果这次“overview 不复用 detail panel，而是新增 scope-level read model”的边界可复用，再同步到 `harness/reuse/history.md` 和 `harness/reuse/index.yaml`。
