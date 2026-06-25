# practice 模块设计

> 状态：Current
> 事实源：`code/backend/internal/module/practice/`、`code/backend/internal/app/composition/practice_module.go`
> 替代：无

## 定位

`practice` 是训练场景 owner，负责普通训练开题、竞赛训练实例编排、AWD desired runtime reconciliation、Flag 提交、人工评阅、得分、个人训练进度和时间线。

## 本文档范围

| 本文档负责 | 本文档不负责（见其他文档） |
|-----------|-------------------------|
| practice 模块的职责边界、HTTP 入口和用例组织 | 实例生命周期和访问控制 → `instance` 模块文档 |
| 训练开题、提交、评阅、进度、时间线的 owner | 容器运行时能力 → `container_runtime` 模块文档 |
| AWD 训练实例 reconciliation 的实现细节 | 题目和 Flag 配置 → `challenge` 模块文档 |
| 模块内部组件协作和数据流 | 跨模块事件发布策略 → `docs/architecture/features/事件发布与降级策略.md` |

## 事实来源

- HTTP 入口：`code/backend/internal/module/practice/api/http/`
- 命令用例：`code/backend/internal/module/practice/application/commands/`
- 查询用例：`code/backend/internal/module/practice/application/queries/`
- 领域规则：`code/backend/internal/module/practice/domain/`
- 契约与端口：`code/backend/internal/module/practice/contracts/`、`code/backend/internal/module/practice/ports/`
- 持久化与 Redis 适配：`code/backend/internal/module/practice/infrastructure/`
- 装配入口：`code/backend/internal/module/practice/runtime/module.go`

## 当前设计

- `practice/api/http`
  - 负责：训练开题、提交、人工评阅、个人进度、时间线和 AWD 实例相关 HTTP 入口。
  - 不负责：直接依赖 command concrete、GORM、Redis 或 container runtime。
- `practice/application/commands`
  - 负责：InstanceLifecycle、RuntimeLifecycle、Submission、ManualReview、Score 等写路径用例族。
  - 不负责：题目事实、实例访问 owner、赛事状态 owner 或 Docker concrete。
- `practice/application/queries`
  - 负责：个人训练进度、时间线和得分查询。
  - 不负责：教师跨班级复盘聚合；这类查询进入 `teaching_analysis`。
- `practice/domain`
  - 负责：训练得分、拓扑运行计划等纯规则。
  - 不负责：Gin、GORM、Redis、Docker 或文件系统操作。
- `practice/infrastructure`
  - 负责：训练仓储、score state store、progress cache、scheduler state、desired AWD reconcile state、submission rate limit、manual review repository、runtime subject adapter。
  - 不负责：拥有 challenge / contest / instance 的源数据。

## API 入口设计

| 路由 | Handler | Service / 用例 |
| --- | --- | --- |
| `POST /api/v1/challenges/:id/instances` | `practice/api/http.Handler.StartChallenge` | `InstanceLifecycleService.StartChallenge`，普通题开题 |
| `POST /api/v1/challenges/:id/submit` | `Handler.SubmitFlag` / `SubmissionHandler` | `SubmissionService`，Flag 提交和判定 |
| `GET /api/v1/challenges/:id/submissions/mine` | `Handler.ListMyChallengeSubmissions` | `SubmissionHistoryService` |
| `GET /api/v1/scoreboard/ranking` | `Handler.GetRanking` | `queries.ScoreService` |
| `GET /api/v1/users/me/progress` | `Handler.GetProgress` | `queries.ProgressTimelineService` |
| `GET /api/v1/users/me/timeline` | `Handler.GetTimeline` | `queries.ProgressTimelineService` |
| `GET /api/v1/teacher/manual-review-submissions` | `Handler.ListTeacherManualReviewSubmissions` | `ManualReviewService` query side |
| `GET /api/v1/teacher/manual-review-submissions/:id` | `Handler.GetTeacherManualReviewSubmission` | `ManualReviewService.GetTeacherManualReviewSubmission` |
| `PUT /api/v1/teacher/manual-review-submissions/:id/review` | `Handler.ReviewManualReviewSubmission` | `ManualReviewService.ReviewManualReviewSubmission` |
| `POST /api/v1/contests/:id/challenges/:cid/instances` | `Handler.StartContestChallenge` | contest scope + instance lifecycle |
| `POST /api/v1/contests/:id/awd/services/:sid/instances` | `Handler.StartContestAWDService` | AWD team/service instance lifecycle |
| `POST /api/v1/contests/:id/awd/services/:sid/instances/restart` | `Handler.RestartContestAWDService` | AWD team/service instance recreate / restart |
| `GET /api/v1/admin/contests/:id/awd/instances` | `Handler.GetAdminContestAWDInstanceOrchestration` | AWD orchestration query |
| `POST /api/v1/admin/contests/:id/awd/instances` | `Handler.StartAdminContestAWDInstance` | 管理端单个 AWD 实例启动 |
| `POST /api/v1/admin/contests/:id/awd/instances/prewarm` | `Handler.PrewarmAdminContestAWDInstances` | 管理端批量预热 AWD 实例 |
| `PUT /api/v1/admin/contests/:id/awd/teams/:team_id/retirement` | `Handler.SetAdminContestAWDTeamRetired` | 队伍退休状态控制 |
| `PUT /api/v1/admin/contests/:id/awd/teams/:team_id/services/:sid/disabled` | `Handler.SetAdminContestAWDTeamServiceDisabled` | 队伍服务禁用状态控制 |
| `PUT /api/v1/admin/contests/:id/awd/teams/:team_id/services/:sid/suppression` | `Handler.SetAdminContestAWDDesiredReconcileSuppressed` | desired reconcile 抑制状态控制 |

## Application / Service 设计

| Service | 代码路径 | 负责 | 不负责 |
| --- | --- | --- | --- |
| InstanceLifecycleService | `practice/application/commands/instance_start_service.go`、`contest_awd_operations.go` | 普通开题、竞赛开题、AWD team/service 启停、scope control 和 prewarm | 实例访问 proxy、runtime node 心跳 |
| RuntimeLifecycleService | `instance_provisioning.go`、`instance_provisioning_scheduler.go`、`awd_desired_runtime_reconciler.go` | 后台 provisioning loop、desired AWD reconcile、异步任务生命周期 | 竞赛状态 owner |
| SubmissionService | `submission_service.go`、`submission_history_service.go` | Flag 提交、动态 Flag 判定、重复解题、提交历史 | 题目 Flag 配置 owner |
| ManualReviewService | `manual_review_service.go` | 教师人工评阅、班级可见性、评阅错误映射 | 题解推荐/隐藏 |
| Score command/query | `score_service.go`、`queries/score_service.go` | 题目得分、用户总分、排行榜缓存 | 竞赛排行榜 |
| ProgressTimelineService | `queries/progress_timeline_service.go` | 用户态 progress / timeline 查询 | 教师跨班聚合 |

## 数据设计

| 表 / 存储 | Entity / Adapter | Owner 语义 | 主要写入方 |
| --- | --- | --- | --- |
| `submissions` | practice repository / `contest/entity.Submission` 共享表 | 普通训练和竞赛提交记录；practice 写普通训练、manual review 相关状态 | submission service / contest submission service |
| `user_scores` | `practice/entity.UserScore` | 用户训练总分、解题数和排名 | score command |
| `challenges` 只读投影 | `practice/entity.Challenge` | practice-facing challenge projection，只读题目事实 | challenge owner 写入 |
| Redis score state | `score_state_store.go` | 用户得分缓存和训练排行榜 sorted-set | score command/query |
| Redis progress cache | `progress_cache.go` | 用户 progress 查询缓存 | progress query / flag accepted handler |
| Redis scheduler / desired reconcile state | `scheduler_state_store.go`、`desired_awd_reconcile_state_store.go` | provisioning 和 AWD desired reconcile 幂等状态 | runtime lifecycle |
| Redis rate limit | `submission_rate_limit_store.go` | Flag 提交限流窗口 | submission service |
| runtime readiness probe | `instance_readiness_probe.go` | access URL HTTP/TCP readiness 探测 | instance lifecycle |

## 边界

- `practice` 拥有训练提交、训练分数、训练进度、训练实例编排状态和人工评阅事实。
- `challenge` 拥有题目、Flag 规则和运行定义；practice 只消费。
- `instance` 拥有实例访问和 runtime identity；practice 通过 instance contract 消费。
- `contest` 拥有 contest scope、队伍、AWD 轮次和 runtime placement；practice 按 contest contract 校验和编排。
- `assessment` 消费 `practice.flag_accepted` 事件更新画像，不写 practice 状态。

## 主要用例

- 学生启动普通题或 AWD 题实例。
- 后台 provisioning loop 创建和重建容器。
- AWD desired runtime reconciler 根据竞赛状态维持队伍服务实例。
- 学生提交 Flag，记录提交历史，更新分数与排行榜缓存。
- 教师人工评阅提交。
- 用户态 `GET /api/v1/users/me/progress`、`GET /api/v1/users/me/timeline`。

## 数据与副作用

- PostgreSQL：提交、用户得分、人工评阅、训练进度相关记录。
- Redis：提交限流、用户分数缓存、排行榜 sorted-set、progress cache、scheduler / reconcile 状态。
- Outbox：Flag accepted 写入事件，触发画像更新、进度缓存删除和通知。
- Runtime：通过 `instance` 与 `container_runtime` capability 创建和维护容器。

## 跨模块依赖

| 依赖 | 用途 | 接线 |
| --- | --- | --- |
| `challenge` | 题目目录、运行定义、Flag 校验 | `composition.BuildPracticeModule` |
| `instance` | 实例仓储、runtime service、访问入口 | `composition.BuildPracticeModule` |
| `contest` | contest scope、AWD runtime placement、队伍约束 | runtime deps / adapters |
| `assessment` | Flag accepted 后刷新画像 | platform outbox handler |
| `ops` | Flag accepted 通知 | platform outbox handler |

## Guardrail

- `code/backend/internal/module/practice/architecture_test.go`：约束 handler 使用聚焦 service、禁止宽 command service facade、保护分层。
- `code/backend/internal/module/practice/ports/*_context_contract_test.go`：约束端口 context 传播。
- `code/backend/internal/app/practice_flow_*_test.go`：覆盖训练主链路。
- `code/backend/internal/module/practice/application/commands/*_test.go`：覆盖开题、provisioning、提交、AWD runtime 和人工评阅。

## 已知限制

- `practice` 当前仍承担普通训练和 AWD 训练编排两类复杂场景；通过聚焦 service 和 runtime lifecycle 拆分控制边界。
