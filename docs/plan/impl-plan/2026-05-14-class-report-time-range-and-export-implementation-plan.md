# 班级报告时间段与导出结构实施计划

> 状态：Draft
> 事实源：`docs/reviews/architecture/2026-05-14-teaching-review-thesis-gap-review.md`、`docs/architecture/features/教学复盘优化设计.md`、`docs/architecture/features/教学复盘建议生成架构.md`、`docs/architecture/features/赛事导出与复盘归档架构.md`
> 替代：无

## 1. 目标

- 让班级复盘接口和班级报告导出都支持固定默认窗口与自定义自然日窗口
- 让班级报告导出显式包含与班级复盘一致的 `summary / trend / review` 结构
- 在不改写现有练习型 `score / solved / rank` 口径的前提下，补齐分类分布、难度分布和竞赛迁移摘要

## 2. 非目标

- 不把所有班级指标都改成“严格按时间窗重算”的新统计体系
- 不在本轮改 recommendation difficulty band owner
- 不在本轮重写 `skill_profiles` 或 AWD 队伍级分数口径
- 不新增新的报告任务表或新的教师报告页面路由

## 3. Brainstorming 结论

- 候选方向 A：直接给接口加 `from/to RFC3339`
  - 不采用：教师班级筛选是按日看教学周期，不是按秒级事件查询；RFC3339 容易把时区和自然日边界重新带回前后端。
- 候选方向 B：在前端自己算最近 7 天，再把 RFC3339 传给后端
  - 不采用：默认值、跨度校验和单参数错误会散到前端、handler、service 多层。
- 候选方向 C：共享 `from_date / to_date` 自然日窗口 owner
  - 采用：默认最近 7 天，单点处理 normalize/default/validate，前后端都容易理解。
- 候选方向 D：班级报告继续自己算一套复盘结论
  - 不采用：会继续制造“页面结论和导出结论不是同一个 owner”的 drift。
- 选定方向：共享时间窗 owner + 复用班级复盘事实 owner + 导出层补映射结构

## 4. owner 与边界

### 4.1 时间窗 contract owner

- 新增 `code/backend/internal/teaching/classwindow`
- 唯一负责：
  - 解析 `from_date / to_date`
  - 设定默认最近 7 天窗口
  - 校验“两个参数要么都不传，要么都传”
  - 校验最大跨度 31 天
  - 产出 downstream 使用的 `since / start_of_day / end_of_day / days / display dates`
- `handler`
  - 只负责绑定原始字符串，不解释默认值
- `application`
  - 调用 `classwindow` 产出规范化窗口
- `repository`
  - 只消费已归一化的 `since / days`

### 4.2 班级复盘与导出结构 owner

- `teaching_query` 继续 owning：
  - 班级 `summary`
  - 班级 `trend`
  - 班级 `review`
- `assessment.ReportService` 继续 owning：
  - 班级报告任务生命周期
  - PDF / Excel 导出结构
  - 报告中“复盘结构 -> 文件结构”的显式映射
- 新增共享 `code/backend/internal/teaching/classreview`
  - 统一把 class snapshot + summary + trend 组装成结构化 review items
  - `teaching_query` 页面接口和 `assessment` 导出都复用这层，避免两处拼 review item

## 5. 导出结构约定

### 5.1 新增字段

- `window`
  - `from_date`
  - `to_date`
  - `days`
- `summary`
- `trend`
- `review`
- `category_distribution`
- `difficulty_distribution`
- `contest_migration`

### 5.2 保留字段

- `class_name`
- `total_students`
- `average_score`
- `dimension_averages`
- `top_students`

### 5.3 口径约定

- `summary / trend / review`：严格使用当前窗口
- `average_score / dimension_averages / top_students / category_distribution / difficulty_distribution / contest_migration`：作为当前班级能力快照，不强制改成窗口化历史重算

## 6. 任务切片

### Slice 1：共享时间窗与班级复盘参数化

目标：

- 后端收口 `from_date / to_date` contract
- 班级 `summary / trend / review` 三个接口走同一时间窗

改动面：

- `code/backend/internal/teaching/classwindow/`
- `code/backend/internal/dto/teacher.go`
- `code/backend/internal/module/teaching_query/api/http/handler.go`
- `code/backend/internal/module/teaching_query/application/queries/contracts.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service_test.go`

验证：

- `go test ./internal/module/teaching_query/application/queries -count=1`

review focus：

- 时间窗默认值是否只在 `classwindow` 一处解释
- `summary / trend / review` 是否真正共用同一窗口
- 单参数、逆序日期、超 31 天跨度是否被拒绝

### Slice 2：班级报告导出结构补强

目标：

- 班级报告入参加时间窗
- 复用班级复盘 owner，补齐 richer export data

改动面：

- `code/backend/internal/teaching/classreview/`
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

验证：

- `go test ./internal/module/assessment/application/commands ./internal/module/assessment/infrastructure -count=1`

review focus：

- 导出里的 `summary / trend / review` 是否与班级复盘结构同源
- 类别/难度分布和竞赛迁移摘要是否保持练习型/教学型口径，不混成队伍分
- PDF / Excel 是否都能承载新增结构

### Slice 3：前端时间段入口与 query sync

目标：

- 班级详情页支持共享时间段筛选
- 导出弹窗预览和导出动作共用同一时间段

改动面：

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

验证：

- `pnpm vitest run src/components/teacher/reports/__tests__/TeacherClassReportExportDialog.test.ts src/views/teacher/__tests__/TeacherClassStudents.test.ts`

review focus：

- 预览与导出 payload 是否共用同一时间段值
- 路由 query 是否能稳定回放当前班级时间窗
- 页面文案是否不再写死“近 7 天”

### Slice 4：契约与设计文档同步

目标：

- 补 API 契约和 OpenAPI
- 更新教学复盘 / 报告导出事实源

改动面：

- `docs/contracts/api-contract-v1.md`
- `docs/contracts/openapi-v1.yaml`
- `docs/architecture/features/教学复盘优化设计.md`
- `docs/architecture/features/教学复盘建议生成架构.md`
- `docs/architecture/features/赛事导出与复盘归档架构.md`

验证：

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

review focus：

- 契约文档是否把 teacher class detail 接口正式收录
- 事实文档是否明确时间窗 owner 和导出映射 owner

## 7. 风险与回退

- 如果把时间窗默认值分散在前端和后端两边，后续容易再次出现“页面是 7 天、导出不是 7 天”的漂移；必须坚持单点 owner。
- 如果班级报告直接 import 页面 DTO 拼装 PDF，后续会让导出结构跟着页面展示细节漂移；应由 `ReportService` 明确 owning export shape。
- 如果竞赛迁移摘要直接使用队伍积分或比赛排名，会把教学评估和赛事成绩混口径；这轮只导出个人可归因参与与成功覆盖摘要。
- 回退方式：
  - 时间窗 contract 回退到默认最近 7 天时，只需保留 `classwindow` 默认路径，不需要再恢复多处硬编码。
  - 导出结构新增字段可向后兼容保留，旧前端只消费状态接口，不会因导出文件内容增量失效。

## 8. 最终验证

- `go test ./internal/module/teaching_query/application/queries ./internal/module/assessment/application/commands ./internal/module/assessment/infrastructure -count=1`
- `pnpm vitest run src/components/teacher/reports/__tests__/TeacherClassReportExportDialog.test.ts src/views/teacher/__tests__/TeacherClassStudents.test.ts`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

## 9. Checklist

- [x] Slice 1：共享时间窗与班级复盘参数化
- [x] Slice 2：班级报告导出结构补强
- [x] Slice 3：前端时间段入口与 query sync
- [x] Slice 4：契约与设计文档同步
- [x] 最终验证完成并通过
