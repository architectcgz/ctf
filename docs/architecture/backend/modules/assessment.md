# assessment 模块设计

> 状态：Current
> 事实源：`code/backend/internal/module/assessment/`、`code/backend/internal/app/composition/assessment_module.go`
> 替代：无

## 定位

`assessment` 是技能画像、推荐、评估报告和复盘归档导出 owner，负责把练习与竞赛结果转换成可查询、可导出的评估事实。

## 本文档范围

| 本文档负责 | 本文档不负责（见其他文档） |
|-----------|-------------------------|
| assessment 模块的职责边界、HTTP 入口和用例组织 | 技能画像计算算法细节 → `docs/architecture/features/技能画像计算架构.md` |
| 画像增量更新、推荐生成、报告输出的 owner | AWD 复盘建议生成算法 → `docs/architecture/features/教学复盘建议生成架构.md` |
| 评估数据表结构和缓存策略 | AWD 归档与报告导出流程 → `docs/architecture/features/AWD教师复盘归档与报告导出设计.md` |
| 模块内部组件协作和数据流 | 跨模块事件发布策略 → `docs/architecture/features/事件发布与降级策略.md` |

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

## 技能画像计算

### 画像数据来源

技能画像基于用户提交记录聚合生成，数据来源：

- **训练提交**：`practice` 模块发布 `EventFlagAccepted` 事件，包含 `userID`、`challengeID`、`dimension`
- **竞赛提交**：`contest` 模块发布 `EventFlagAccepted` 和 `EventAWDAttackAccepted` 事件
- **题目维度**：从 `challenge` 模块读取题目维度（`web`、`crypto`、`pwn`、`reverse`、`misc` 等）

### 维度权重计算

画像计算采用 **增量更新 + 按需重建** 双轨策略：

1. **增量更新**（主流程）：
   - 订阅 `practice.FlagAccepted` 和 `contest.FlagAccepted` 事件
   - 事件触发后，调用 `ProfileService.UpdateSkillProfileForDimension(userID, dimension)`
   - 从数据库读取该用户在该维度下的所有正确提交，重新聚合分数
   - 更新 `skill_profiles.score` 字段

2. **维度权重计算公式**：
   - 基础分数 = Σ (题目难度分 × 题目维度权重)
   - 题目维度权重由 `challenge.tags` 和 `challenge_tags` 表决定
   - 维度总分缓存在 Redis，题目发布目录变更时失效

3. **缓存失效时机**：
   - `DimensionTotalCacheInvalidationService` 订阅 `challenge.PublishedCatalogChanged` 事件
   - 题目发布 / 下架 / 维度变更时，调用 `store.DeletePublishedDimensionTotals(ctx)` 清空维度总分缓存
   - 下次查询时重新计算并缓存

### 计算触发时机

| 触发方式 | 时机 | 代码位置 |
| --- | --- | --- |
| **实时增量更新** | 每次正确提交后，通过事件驱动 | `ProfileService.handleFlagAcceptedEvent()` |
| **后台重建 job**（按需） | 管理员手动触发或数据修复时 | `ProfileService.RebuildProfile(userID)` |
| **维度缓存失效** | 题目发布目录变更时 | `DimensionTotalCacheInvalidationService.handlePublishedCatalogChangedEvent()` |

**代码位置**：
- `code/backend/internal/module/assessment/application/commands/profile_service.go`：画像增量更新用例
- `code/backend/internal/module/assessment/entity/skill_profile.go`：画像持久化实体
- `code/backend/internal/module/assessment/application/commands/dimension_total_cache_invalidation_service.go`：维度总分缓存失效
- `code/backend/internal/module/assessment/infrastructure/state_store.go`：Redis 缓存适配器

**相关专题**：
- 推荐算法与画像应用 → `docs/architecture/features/教学复盘建议生成架构.md`

## 报告生成流程

### 报告类型与格式

`assessment` 模块支持多种报告类型和输出格式：

| 报告类型常量 | 代码值 | 输出格式 | 适用场景 |
| --- | --- | --- | --- |
| `ReportTypePersonal` | `"personal"` | PDF / JSON | 学生个人技能画像报告 |
| `ReportTypeClass` | `"class"` | Excel / PDF | 班级整体技能画像报告 |
| `ReportTypeContest` | `"contest_export"` | ZIP（包含 JSON + Excel） | 竞赛结果导出 |
| `ReportTypeReview` | `"review_archive"` | ZIP（包含学生复盘归档） | 教师学生复盘归档 |
| `ReportTypeAWDReviewArchive` | `"awd_review_archive"` | ZIP（包含 AWD 复盘数据） | AWD 复盘归档 |
| `ReportTypeAWDReviewReport` | `"awd_review_report"` | PDF | AWD 复盘分析报告 |

### 报告生成异步流程

报告生成采用 **请求 - 后台生成 - 轮询下载** 模式：

1. **创建报告请求**：
   - Handler 接收 HTTP 请求，调用 `ReportService.CreateXxxReport()`
   - 创建 `reports` 表记录，状态为 `ReportStatusProcessing`
   - 返回 `report_id` 给前端

2. **后台生成**：
   - 后台 goroutine 或 job 执行报告渲染
   - 调用 `reporting/` 下的渲染器生成 PDF / Excel / JSON 文件
   - 文件存储到 LocalFS（通过 `ReportOutputStore`）
   - 更新 `reports.status` 为 `ReportStatusReady` 或 `ReportStatusFailed`

3. **前端轮询**：
   - 前端通过 `GET /api/v1/reports/:id` 轮询报告状态
   - 状态为 `ready` 后，调用 `GET /api/v1/reports/:id/download` 下载文件

### 报告资产存储

- **存储路径**：LocalFS `data/reports/` 目录，由 `ReportOutputStore` 管理
- **路径结构**：`<storage-root>/<report-type>/<report-id>/<filename>`
- **路径安全**：通过 `ReportOutputStore` 检查相对路径，拒绝包含 `..` 的路径
- **生命周期**：报告记录包含 `expires_at` 字段，过期后由 `Cleaner` job 清理

### 缓存失效策略

报告生成不依赖缓存，但推荐和画像缓存会影响报告内容：

- **画像锁**：增量更新时通过 Redis 锁防止并发写入
- **推荐缓存**：个人报告中包含推荐题目，缓存 TTL 为配置的 `recommendation_cache_ttl`
- **维度总分缓存**：题目发布目录变更时失效，下次查询重新计算

**代码位置**：
- `code/backend/internal/module/assessment/application/commands/report_service.go`：报告生成用例
- `code/backend/internal/module/assessment/entity/report.go`：报告实体和状态常量
- `code/backend/internal/module/assessment/infrastructure/report_output_store.go`：报告文件存储适配器
- `code/backend/internal/module/assessment/application/reporting/`：PDF / Excel / JSON 渲染器
- `code/backend/internal/module/assessment/application/commands/cleaner.go`：报告清理 job

**相关专题**：
- AWD 复盘报告生成 → `docs/architecture/features/AWD教师复盘归档与报告导出设计.md`

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
- `code/backend/internal/module/assessment/application/commands/profile_service_test.go`：覆盖画像增量更新和事件订阅。
- `code/backend/internal/module/assessment/infrastructure/report_output_store_test.go`：覆盖报告文件存储路径安全检查。
- `docs/architecture/features/AWD教师复盘归档与报告导出设计.md`：AWD 复盘导出事实源。

## 已知限制

- 评估结果依赖 practice / contest 事件与查询事实；若事件补偿失败，需要通过重建画像或报告重新生成收敛。
