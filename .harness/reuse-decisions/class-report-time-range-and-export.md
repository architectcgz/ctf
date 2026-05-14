# Reuse Decision

## Change type

api / handler / service / repository / readmodel / frontend page / modal / tests / docs

## Existing code searched

- `code/backend/internal/module/teaching_query/api/http/handler.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service_test.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
- `code/backend/internal/module/assessment/api/http/report_handler.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/frontend/src/api/teacher/classes.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/frontend/src/features/teacher-class-students/model/useTeacherClassStudentsPage.ts`
- `code/frontend/src/features/teacher-class-report-export/model/useTeacherClassReportExport.ts`
- `code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue`
- `code/frontend/src/components/teacher/reports/TeacherClassReportExportDialog.vue`
- `docs/contracts/api-contract-v1.md`
- `docs/contracts/openapi-v1.yaml`
- `docs/architecture/features/教学复盘优化设计.md`
- `docs/architecture/features/教学复盘建议生成架构.md`
- `docs/architecture/features/赛事导出与复盘归档架构.md`
- `docs/reviews/architecture/2026-05-14-teaching-review-thesis-gap-review.md`

## Similar implementations found

- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
  - 已经是班级 `summary / trend / teaching fact snapshot` 的唯一读 owner，底层查询本身支持 `since + days`，只是上层把窗口写死成最近 7 天。
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service.go`
  - 已经把 `summary / trend / review` 装配到教师班级详情接口里，适合作为班级复盘结构的事实源。
- `code/backend/internal/module/assessment/application/commands/report_service.go`
  - 已经负责班级报告异步生命周期、文件渲染和导出入口，适合作为“导出层” owner，而不是新建第二套导出任务链路。
- `code/frontend/src/features/teacher-class-students/model/useTeacherClassStudentsPage.ts`
  - 已经统一加载班级详情页 `summary / trend / review`，适合作为班级时间段筛选入口。
- `code/frontend/src/features/teacher-class-report-export/model/useTeacherClassReportExport.ts`
  - 已经统一加载导出弹窗预览和导出动作，适合作为导出时间段入口。

## Decision

extend_existing

## Reason

这次不是再造一套“班级报告读模型”，而是把现有班级复盘 owner 和报告导出 owner 接起来，并把时间窗 contract 收口成单点解释。

选定方向：

- 继续复用 `teaching_query` 作为班级 `summary / trend / review` 的事实 owner。
- 继续复用 `assessment.ReportService` 作为班级报告导出 owner。
- 新增共享 `classwindow` 小包，统一解释 `from_date / to_date`：
  - 默认最近 7 天
  - 只传一个参数直接判 invalid
  - 两个都传时按自然日闭区间解释
  - 最大跨度 31 天
- 报告导出不重新发明班级复盘结构，而是显式带出：
  - 时间窗元信息
  - 当前窗口的 `summary / trend / review`
  - 班级能力快照的分类分布、难度分布、竞赛迁移摘要
  - 原有 `total_students / average_score / dimension_averages / top_students`
- 前端继续复用班级详情页和导出弹窗现有入口，只补时间段表单、query sync 和 typed payload。

不采用的方向：

- 不在 `handler`、`application`、`repository` 三层各自解释时间窗默认值和校验规则。
- 不在 `assessment` 里复制一套班级复盘建议规则。
- 不用 `from/to RFC3339`，避免教师侧按日筛选时出现时区和半天边界歧义。

## Files to modify

- `.harness/reuse-decisions/class-report-time-range-and-export.md`
- `docs/plan/impl-plan/2026-05-14-class-report-time-range-and-export-implementation-plan.md`
- `code/backend/internal/teaching/classwindow/`
- `code/backend/internal/teaching/classreview/`
- `code/backend/internal/module/teaching_query/api/http/handler.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service_test.go`
- `code/backend/internal/module/teaching_query/application/queries/contracts.go`
- `code/backend/internal/module/assessment/api/http/report_handler.go`
- `code/backend/internal/module/assessment/api/http/request_mapper_gen.go`
- `code/backend/internal/module/assessment/application/commands/report_command_input.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `code/backend/internal/module/assessment/domain/report.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository_test.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/assessment/runtime/module.go`
- `code/backend/internal/dto/report.go`
- `code/backend/internal/dto/teacher.go`
- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teacher/classes.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/frontend/src/features/teacher-class-students/model/useTeacherClassStudentsPage.ts`
- `code/frontend/src/features/teacher-class-report-export/model/useTeacherClassReportExport.ts`
- `code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue`
- `code/frontend/src/components/teacher/TeacherClassTrendPanel.vue`
- `code/frontend/src/components/teacher/reports/TeacherClassReportExportDialog.vue`
- `code/frontend/src/components/teacher/reports/__tests__/TeacherClassReportExportDialog.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/contracts/openapi-v1.yaml`
- `docs/architecture/features/教学复盘优化设计.md`
- `docs/architecture/features/教学复盘建议生成架构.md`
- `docs/architecture/features/赛事导出与复盘归档架构.md`

## After implementation

- 班级详情页和班级报告导出会共享同一套时间窗解释，不再各自写死“近 7 天”。
- 导出结构会显式映射班级复盘结构，页面可见结论可以直接进报告产物。
- 班级报告仍保持练习型 `score / solved / rank` 口径，不把 AWD 队伍分直接改写成个人/班级得分。
