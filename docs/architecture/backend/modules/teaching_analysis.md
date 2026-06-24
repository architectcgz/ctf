# teaching_analysis 模块设计

> 状态：Current
> 事实源：`code/backend/internal/module/teaching_analysis/`、`code/backend/internal/app/composition/teaching_analysis_module.go`
> 替代：无

## 定位

`teaching_analysis` 是教师视角只读查询聚合模块，负责教学复盘、班级洞察、学生活动和教师概览等跨 owner 查询拼装。

## 事实来源

- HTTP 入口：`code/backend/internal/module/teaching_analysis/api/http/handler.go`
- 查询用例：`code/backend/internal/module/teaching_analysis/application/queries/`
- 对外响应契约：`code/backend/internal/module/teaching_analysis/contracts/`
- 查询端口：`code/backend/internal/module/teaching_analysis/ports/query.go`
- 查询仓储：`code/backend/internal/module/teaching_analysis/infrastructure/`
- 专题事实：`docs/architecture/features/教学复盘优化设计.md`、`docs/architecture/features/教师教学概览聚合架构.md`

## 当前设计

- `teaching_analysis/api/http`
  - 负责：教师分析、教学概览、班级洞察和学生复盘查询入口。
  - 不负责：写 practice / contest / assessment / identity 状态。
- `teaching_analysis/application/queries`
  - 负责：跨 owner 只读查询编排、分页/切片、响应映射和 recommendation 拼装。
  - 不负责：训练判分、画像更新、报告生成或赛事状态修改。
- `teaching_analysis/ports`
  - 负责：按查询 use case 拆分 repository 能力，例如用户 lookup、班级洞察、学生活动、教学概览。
  - 不负责：声明宽 `Repository` 接口或带 GORM tag 的 persistence model。
- `teaching_analysis/infrastructure`
  - 负责：只读 SQL 查询、共享行结构和学生目录 / 活动 / profile / overview / class insight repository。
  - 不负责：成为业务写模型 owner。

## API 入口设计

| 路由 | Handler | Service / 用例 |
| --- | --- | --- |
| `GET /api/v1/teacher/overview` | `teaching_analysis/api/http.Handler.GetOverview` | `queries.OverviewService` |
| `GET /api/v1/teacher/classes` | `Handler.ListClasses` | class / student directory repository |
| `GET /api/v1/teacher/students` | `Handler.ListStudents` | student directory query |
| `GET /api/v1/teacher/classes/:name/students` | `Handler.ListClassStudents` | class insight repository |
| `GET /api/v1/teacher/classes/:name/summary` | `Handler.GetClassSummary` | `ClassInsightService` |
| `GET /api/v1/teacher/classes/:name/trend` | `Handler.GetClassTrend` | `ClassInsightService` |
| `GET /api/v1/teacher/classes/:name/review` | `Handler.GetClassReview` | class review query |
| `GET /api/v1/teacher/students/:id/progress` | `Handler.GetStudentProgress` | `queries.Service` |
| `GET /api/v1/teacher/students/:id/recommendations` | `Handler.GetStudentRecommendations` | assessment recommendation provider + query service |
| `GET /api/v1/teacher/students/:id/timeline`、`/evidence`、`/attack-sessions` | `Handler` | `StudentReviewService` |

## Application / Service 设计

| Service | 代码路径 | 负责 | 不负责 |
| --- | --- | --- | --- |
| Query service | `teaching_analysis/application/queries/service.go` | 学生进度、证据和基础教师分析查询 | 写训练/竞赛状态 |
| Overview service | `overview_service.go` | 教师概览统计 | dashboard 运维统计 |
| Class insight service | `class_insight_service.go` | 班级 summary、trend、review 和学生列表切片 | 画像更新 |
| Student review service | `student_review_service.go` | timeline、evidence、attack sessions 和复盘数据整合 | 报告文件生成 |
| Response mappers | `response_mapper*.go`、`class_insight_response_mapper.go` | 查询响应 DTO 映射 | persistence row 暴露 |

## 数据设计

`teaching_analysis` 不拥有写表；它只读多个 owner 的当前事实。

| 只读来源 | 用途 | Owner |
| --- | --- | --- |
| `users` | 学生、教师、班级和基础资料 | `identity` |
| `challenges`、`tags`、`challenge_tags` | 题目维度、推荐和训练事实拼装 | `challenge` |
| `submissions`、`user_scores` | 学生进度、解题、时间线 | `practice` |
| `contest_*`、`teams`、`team_members` | 竞赛参与、队伍和赛事上下文 | `contest` |
| `skill_profiles`、`reports` | 画像、推荐和归档状态 | `assessment` |
| `audit_logs`、`submission_writeups`、`awd_attack_logs`、`awd_traffic_events` | 证据链、攻击会话和复盘材料 | `ops` / `challenge` / `contest` |

所有 repository 位于 `teaching_analysis/infrastructure/`，端口位于 `teaching_analysis/ports/query.go`；ports 禁止带 GORM tag 或声明宽 `Repository`。

## 边界

- `teaching_analysis` 可以读取 identity、practice、contest、assessment 等事实拼装教师视角。
- `teaching_analysis` 不写业务状态，不拥有训练进度、评估画像、竞赛或用户事实。
- 如果查询只读取 practice 自有事实且属于用户态 progress / timeline，留在 `practice`。
- 如果查询结果变成可导出报告生命周期，owner 是 `assessment`。

## 主要用例

- 教师教学概览。
- 班级洞察和学生目录。
- 学生训练活动时间线和复盘详情。
- 教师复盘建议聚合，消费 assessment recommendation provider。

## 数据与副作用

- PostgreSQL：只读查询多个 owner 的表和读模型。
- 无业务写副作用：不写 outbox，不改训练、竞赛、评估或用户状态。
- 响应 DTO：由 `contracts` 和 `application/queries/*response*` 控制，不暴露 persistence row。

## 跨模块依赖

| 依赖 | 用途 | 接线 |
| --- | --- | --- |
| `identity` | 基础用户 lookup | `composition.BuildTeachingAnalysisModule` |
| `assessment` | 推荐、画像相关 provider | `composition.BuildTeachingAnalysisModule` |
| `practice` | 学生活动和训练事实只读来源 | repository 查询 |
| `contest` | 竞赛参与和 AWD 复盘只读来源 | repository 查询 |

## Guardrail

- `code/backend/internal/module/teaching_analysis/architecture_test.go`：禁止 API / queries 依赖 infrastructure，禁止宽 repository 和 GORM tag 泄漏到 ports。
- `code/backend/internal/module/teaching_analysis/api/http/handler_contract_test.go`：约束 handler 构造 contract。
- `code/backend/internal/module/teaching_analysis/application/queries/*_test.go`：覆盖概览、班级洞察和学生复盘查询。

## 已知限制

- 模块当前是只读聚合；如未来出现教师侧写操作，应先判断写状态的真实业务 owner，而不是直接扩展本模块。
