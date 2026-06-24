# assessment 模块设计

> 状态：Current
> 事实源：`code/backend/internal/module/assessment/`、`code/backend/internal/app/composition/assessment_module.go`
> 替代：无

## 定位

`assessment` 是技能画像、推荐、评估报告和复盘归档导出 owner，负责把练习与竞赛结果转换成可查询、可导出的评估事实。

## 事实来源

- HTTP 入口：`code/backend/internal/module/assessment/api/http/`
- 命令用例：`code/backend/internal/module/assessment/application/commands/`
- 查询用例：`code/backend/internal/module/assessment/application/queries/`
- 报告渲染：`code/backend/internal/module/assessment/application/reporting/`
- 领域模型：`code/backend/internal/module/assessment/domain/`
- 持久化与输出存储：`code/backend/internal/module/assessment/infrastructure/`
- 专题事实：`docs/architecture/features/教学复盘建议生成架构.md`、`docs/architecture/features/AWD教师复盘归档与报告导出设计.md`

## 当前设计

- `assessment/api/http`
  - 负责：画像、推荐、报告、教师 AWD review 导出等 HTTP 入口。
  - 不负责：直接写报告文件、访问 GORM 或拼接跨 owner SQL。
- `assessment/application/commands`
  - 负责：画像增量更新、维度总分缓存失效、报告生成、报告输出、AWD review export builder / renderer、报告清理。
  - 不负责：训练提交判分、竞赛榜单计算或教师页面聚合。
- `assessment/application/queries`
  - 负责：画像查询、推荐查询、教师 AWD review 查询。
  - 不负责：班级洞察和复盘页面的跨模块聚合；这类查询由 `teaching_analysis` 承接。
- `assessment/application/reporting`
  - 负责：PDF / spreadsheet / JSON 等报告渲染细节。
  - 不负责：决定业务权限和报告生命周期。
- `assessment/infrastructure`
  - 负责：画像、报告、AWD review repository、Redis state store、LocalFS report output store。
  - 不负责：拥有 practice / contest / challenge 源数据。

## API 入口设计

| 路由 | Handler | Service / 用例 |
| --- | --- | --- |
| `GET /api/v1/users/me/skill-profile` | `assessment/api/http.Handler.GetMySkillProfile` | `queries.ProfileService` |
| `GET /api/v1/users/me/recommendations` | `Handler.GetRecommendations` | `queries.RecommendationService` |
| `GET /api/v1/users/:id/skill-profile` | `Handler.GetStudentSkillProfile` | 教师视角画像查询 |
| `POST /api/v1/reports/personal` | `ReportHandler.CreatePersonalReport` | `commands.ReportService` |
| `POST /api/v1/reports/class` | `ReportHandler.CreateClassReport` | `commands.ReportService` |
| `POST /api/v1/admin/contests/:id/export` | `ReportHandler.CreateContestExport` | `commands.ReportService.CreateContestExport`，竞赛导出报告 |
| `GET /api/v1/reports/:id` | `ReportHandler.GetReportStatus` | report query/status |
| `GET /api/v1/reports/:id/download` | `ReportHandler.DownloadReport` | `ReportOutputStore` 下载 |
| `GET /api/v1/teacher/awd/reviews`、`/awd/reviews/:id` | `TeacherAWDReviewHandler` | `queries.TeacherAWDReviewService` |
| `POST /api/v1/teacher/awd/reviews/:id/export/archive`、`/export/report` | `TeacherAWDReviewHandler` | AWD review export builder / renderer |
| `GET /api/v1/teacher/students/:id/review-archive`、`POST /api/v1/teacher/students/:id/review-archive/export` | `ReportHandler.GetStudentReviewArchive`、`CreateStudentReviewArchive` | student review archive builder |

## Application / Service 设计

| Service | 代码路径 | 负责 | 不负责 |
| --- | --- | --- | --- |
| Profile command service | `assessment/application/commands/profile_service.go` | 正确提交后的画像增量更新和重建 | practice 提交判分 |
| Dimension cache invalidation | `dimension_total_cache_invalidation_service.go` | 维度总分缓存失效 | 推荐生成规则 |
| Recommendation query service | `assessment/application/queries/recommendation_service.go` | 推荐题目生成、默认 limit 缓存和失效 | 题目目录 owner |
| Report service | `assessment/application/commands/report_service.go`、`report_generation.go` | 个人、班级、竞赛、学生复盘归档报告的生命周期、权限校验、生成和状态 | 教师聚合页面查询、竞赛状态推进 |
| Reporting renderers | `assessment/application/reporting/` | PDF / spreadsheet / JSON 渲染 | 报告业务状态 |
| AWD review query/export | `teacher_awd_review_service.go`、`awd_review_export_*` | AWD 复盘数据查询、归档和报告导出 | AWD 轮次状态推进 |
| Cleaner | `assessment/application/commands/cleaner.go` | 过期报告输出清理 | storage 全局 GC |

## 数据设计

| 表 / 存储 | Entity / Adapter | Owner 语义 | 主要写入方 |
| --- | --- | --- | --- |
| `skill_profiles` | `assessment/entity.SkillProfile` | 用户技能画像、维度分和推荐基础事实 | profile command |
| `reports` | `assessment/entity.Report` | 报告任务、状态、类型、输出路径和错误 | report service |
| contest export 只读数据集 | `assessment/ports.AssessmentContestExportRepository`、`report_repository.go` | 竞赛导出读取 scoreboard、challenge、team 等快照，不写 contest 状态 | report service |
| `challenges` / `tags` / `challenge_tags` 只读 | assessment repository rows | 推荐和画像维度来源 | challenge owner 写入 |
| LocalFS report output | `report_output_store.go` | 报告文件、ZIP/PDF/XLSX/JSON 输出和安全路径检查 | report service |
| Redis state store | `assessment/infrastructure/state_store.go` | 画像锁、推荐缓存、维度总分缓存 | profile / recommendation services |
| AWD review repository | `awd_review_repository.go` | AWD 复盘查询数据集 | assessment query/export |

## 边界

- `assessment` 拥有技能画像、推荐缓存、报告记录和报告输出文件事实。
- `practice` 的正确提交事件触发画像更新，但 practice 不写 assessment 表。
- `contest` 可作为报告数据来源，但 contest 不拥有报告生命周期。
- `teaching_analysis` 可消费 assessment 推荐和画像能力拼装教师视角。
- `challenge` 提供题目维度、目录等评估所需事实。

## 主要用例

- 正确提交后增量更新技能画像。
- 查询个人技能画像和推荐题目。
- 生成学生、班级、竞赛或 AWD 复盘报告。
- 渲染 PDF / spreadsheet / JSON 输出并提供下载。
- 清理过期报告输出。

## 数据与副作用

- PostgreSQL：技能画像、报告记录和评估相关实体。
- Redis：画像锁、推荐缓存、维度总分缓存。
- LocalFS：报告输出文件，路径安全由 `report_output_store.go` 检查。
- Background tasks：报告导出和清理由 assessment module 注册到 app lifecycle。

## 跨模块依赖

| 依赖 | 用途 | 接线 |
| --- | --- | --- |
| `challenge` | 题目维度和目录 | `composition.BuildAssessmentModule` |
| `practice` | Flag accepted 事件驱动画像更新 | platform outbox handler |
| `contest` | 竞赛报告和 AWD review 数据来源 | repository / contract |
| `teaching_analysis` | 教师聚合消费推荐和画像 | `composition.BuildTeachingAnalysisModule` |

## Guardrail

- `code/backend/internal/module/assessment/architecture_test.go`：约束 API / commands / queries / ports 分层。
- `code/backend/internal/module/assessment/ports/state_store_context_contract_test.go`：约束 Redis state store context。
- `code/backend/internal/module/assessment/application/commands/report_*_test.go`：覆盖报告生成、渲染、命名和生命周期。
- `docs/architecture/features/AWD教师复盘归档与报告导出设计.md`：AWD 复盘导出事实源。

## 已知限制

- 评估结果依赖 practice / contest 事件与查询事实；若事件补偿失败，需要通过重建画像或报告重新生成收敛。
