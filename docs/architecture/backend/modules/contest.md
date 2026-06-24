# contest 模块设计

> 状态：Current
> 事实源：`code/backend/internal/module/contest/`、`code/backend/internal/app/composition/contest_module.go`
> 替代：无

## 定位

`contest` 是赛事 owner，负责竞赛配置、报名、队伍、题目编排、提交、排行榜、公告、状态推进、AWD 轮次、AWD 服务、checker、攻击、防守流量和实时事件。

## 事实来源

- HTTP 入口：`code/backend/internal/module/contest/api/http/`
- 命令用例：`code/backend/internal/module/contest/application/commands/`
- 查询用例：`code/backend/internal/module/contest/application/queries/`
- 后台任务：`code/backend/internal/module/contest/application/jobs/`
- 状态机：`code/backend/internal/module/contest/application/statusmachine/`、`code/backend/internal/module/contest/domain/status_transition.go`
- 持久化与运行时适配：`code/backend/internal/module/contest/infrastructure/`
- 专题事实：`docs/architecture/features/校园级CTF-AWD模式完整设计.md`

## 当前设计

- `contest/api/http`
  - 负责：赛事、报名、队伍、题目、提交、榜单、公告、AWD 管理、AWD workspace、攻击日志、流量和 realtime WebSocket HTTP 入口。
  - 不负责：直接访问 GORM、Redis、Docker、checker runner 或 challenge repository。
- `contest/application/commands`
  - 负责：赛事创建更新、报名审核、队伍生命周期、题目入赛、提交判分、榜单改分、AWD round/service/attack/readiness/preview 等写路径。
  - 不负责：题目源数据维护、训练实例生命周期或容器执行 concrete。
- `contest/application/queries`
  - 负责：赛事列表详情、可见题目、报名进度、队伍信息、排行榜、AWD summary / workspace / traffic / attack log 等查询。
  - 不负责：教师教学复盘聚合。
- `contest/application/jobs`
  - 负责：竞赛状态推进、AWD round updater、checker runner、服务检查、状态回写和锁续约。
  - 不负责：进程级 job 生命周期装配；生命周期由 runtime / composition 注册。
- `contest/infrastructure`
  - 负责：赛事、队伍、报名、提交、AWD、scoreboard、realtime outbox、checker、Redis state store 和 Docker checker adapter。
  - 不负责：拥有 challenge 源数据、practice 训练状态或 ops 通知事实。

## 状态机与并发控制

### 本章节范围

| 本文档负责 | 本文档不负责 |
|----------|------------|
| 说明 contest 状态转换规则、CAS 并发控制、Redis 分布式锁和代码位置 | 替代专题文档中的完整状态图；详细 side effect 编排见专题文档 |
| 记录状态推进后台 job、乐观锁 + 分布式锁组合机制 | 绕过状态机的手动管理命令设计 |
| 指向 `contest_status_update_repository.go` 的 CAS 实现 | 记录每个状态的业务语义；业务规则见 domain 层和专题文档 |

### 状态转换规则

- **状态流转**：`draft` → `preparing` → `running` → `ended` → `archived`
  - `draft`：初始创建态，管理员可以修改基本信息、时间窗口和模式
  - `preparing`：进入筹备态，允许添加题目、配置 AWD 服务、预分配 runtime placement
  - `running`：开赛，学生可见题目、提交 Flag、AWD 攻击和查看榜单
  - `ended`：比赛结束，冻结提交和攻击，可以解冻榜单和导出报告
  - `archived`：归档态，只读历史记录
- **代码位置**：
  - 状态定义：`code/backend/internal/module/contest/entity/contest.go`（`ContestStatus` 枚举）
  - 转换规则：`code/backend/internal/module/contest/domain/status_transition.go`（`ValidateStatusTransition`）
  - 定时推进：`code/backend/internal/module/contest/application/jobs/status_transition_service.go`（后台 job 按时间窗触发）
  - 手动命令：`code/backend/internal/module/contest/application/commands/contest_service.go`（管理员手动 freeze/unfreeze/archive）

### 并发控制机制

- **乐观锁（数据库层）**：
  - `contests.status_version` 字段作为 CAS 版本号
  - `contest_status_update_repository.go` 使用 `WHERE id = ? AND status = ? AND status_version = ? AND deleted_at IS NULL` 条件更新
  - 更新成功时 `status_version` 加一，失败时返回 `RowsAffected == 0`，调用方判断是否重试
  - 同一事务内写入 `contest_status_transitions` 记录，确保状态行和迁移历史原子提交
- **分布式锁（Redis 层）**：
  - `code/backend/internal/module/contest/infrastructure/status_update_lock_store.go` 提供 Redis 分布式锁
  - 后台 job `status_transition_service.go` 在推进多个赛事状态前先获取全局锁 `contest:status:update`，避免多 API 副本重复推进
  - 锁持有时长默认 30 秒，推进完成后释放；如果进程崩溃，锁自动过期不会永久阻塞
- **组合策略**：
  - Redis 锁减少并发推进的无效 CAS 尝试，降低数据库压力
  - 数据库 CAS 保证最终一致性，即使 Redis 锁失效也不会产生脏数据
- **专题文档引用**：
  - 完整状态机图和 side effect 编排 → `docs/architecture/backend/design/contest-status-state-machine.md`
  - 事务管理模式和 CAS 模式 → `docs/architecture/features/数据库事务管理模式.md`

### 代码位置

- `code/backend/internal/module/contest/infrastructure/contest_status_update_repository.go`：CAS 更新实现
- `code/backend/internal/module/contest/application/jobs/status_transition_service.go`：定时推进 job
- `code/backend/internal/module/contest/infrastructure/status_update_lock_store.go`：Redis 分布式锁
- `code/backend/internal/module/contest/domain/status_transition.go`：状态转换规则验证
- `code/backend/internal/module/contest/infrastructure/contest_status_transition_repository.go`：迁移历史持久化

## 事务处理与事件发布

### 本章节范围

| 本文档负责 | 本文档不负责 |
|----------|------------|
| 列举 contest 模块的显式事务 runner 和代码位置 | 每个事务 runner 的完整业务不变量；不变量由测试覆盖 |
| 说明弱事件、realtime outbox、platform outbox 的使用场景 | 记录 outbox dispatcher 实现细节；详见专题文档 |
| 指向事务管理模式和事件发布策略的专题文档 | 替代专题文档；冲突时以专题文档为准 |

### Transaction Runner 模式

- **显式事务边界**：contest 写路径使用命名 runner，由 `ports/*.go` 暴露接口、`infrastructure/*.go` 打开 GORM transaction
- **当前 runner 清单**：
  - `WithinScoringTransaction`：提交判分、榜单更新
  - `WithinAnnouncementTransaction`：公告创建、更新、删除
  - `WithinServiceCheckTransaction`：AWD 服务检查结果写入
  - `WithinAttackLogTransaction`：AWD 攻击提交、Flag 验证、计分
  - `WithinRoundReconcileTransaction`：AWD 轮次推进、服务状态协调
  - `WithinRoundServiceWritebackTransaction`：AWD 服务状态回写
- **模式约束**：
  - Application 层只依赖事务 port，不直接依赖 `*gorm.DB`
  - Infrastructure 层打开事务并传递事务 scoped repository 给 application callback
  - 事务内 outbox enqueue 失败会返回错误，触发业务写入回滚
- **代码位置**：
  - Port 定义：`code/backend/internal/module/contest/ports/transaction.go`、`awd_transaction.go`
  - Infrastructure 实现：`code/backend/internal/module/contest/infrastructure/submission_repository.go`（`WithinScoringTransaction` 实现）、`awd_repository.go`（AWD 事务实现）
  - Application 使用：`code/backend/internal/module/contest/application/commands/submission_score_transaction.go`、`awd_attack_log_transaction.go`
- **专题文档引用**：完整事务管理模式 → `docs/architecture/features/数据库事务管理模式.md`

### 事件发布策略

- **弱事件（Weak Event）**：
  - 用于"主业务已完成，事件失败不应影响用户动作结果"的场景
  - `SubmissionService.publishWeakEvent()` 与 `AWDService.publishWeakEvent()`：`eventBus` 缺失或 `Publish()` 失败只记录 `publish_contest_event_failed` warn 日志，不回滚业务写入
  - 适用场景：实时 UI 刷新提示、非关键日志、可丢失的状态通知
- **Contest Realtime Outbox**：
  - `contest_realtime_outbox` 表用于可靠投递竞赛实时 relay 事件
  - 使用场景：公告推送、提交榜单更新、checker preview 进度、状态推进 realtime side effect
  - 在需要强 side effect 时，`EnqueueRealtimeRelay()` 必须在业务事务内调用，失败会触发回滚
- **Platform Event Outbox**：
  - `platform_event_outbox` 表用于跨模块可靠投递的平台事件
  - 使用场景：通知、发布检查、需要审计 / 去重 / 恢复的 side effect
  - Outbox dispatcher 按 `route`（`handler` / `stream`）分发，失败后按 retry backoff 延后重试
- **取舍原则**：
  - 弱事件：主业务已完成，事件失败不应阻塞用户动作
  - Outbox：需要恢复投递、去重或跨模块消费的 side effect
  - 禁止把弱事件失败写成业务失败；也禁止把需要审计的 side effect 只放进弱事件
- **专题文档引用**：完整事件发布与降级策略 → `docs/architecture/features/事件发布与降级策略.md`

## AWD Flag 轮换与注入

- `contest/application/commands/awd_flag_support.go` 与 `contest/application/jobs/awd_round_flag_lookup_support.go`
  - 负责：优先从 Redis round state 读取当前 materialized Flag；不存在时用 `contestdomain.BuildAWDRoundFlag(contestID, roundNumber, teamID, awdChallengeID, flagSecret, flagPrefix)` 确定性复算。
  - 不负责：把每轮 Flag 明文写入 PostgreSQL，或依赖容器内 Flag 作为判题真相。
- `contest/application/commands/awd_attack_flag_rotation_support.go`
  - 负责：攻击成功时用 `ReplaceAWDRoundFlagIfMatch()` 对当前 victim team + service 的 Flag 做 CAS 轮换；注入失败会回滚 Redis cache，并尝试把旧 Flag 重新注入容器。
  - 不负责：在攻击失败、重复提交或上一轮宽限命中时轮换当前轮 Flag。
- `contest/application/jobs/awd_round_flag_sync.go`、`contest/infrastructure/awd_flag_injector.go`、`contest/runtime/wiring.go`
  - 负责：轮次推进和攻击轮换都通过 `AWDFlagInjector` 写入对应队伍服务容器；装配时使用 `NewDockerAWDFlagInjector(...)`，缺少 writer 时降级为 noop injector 并记录日志。
  - 不负责：跨 runtime node 查找容器；实例定位仍依赖 service_id、team_id 和 instance/runtime identity。

## API 入口设计

| 路由组 | 代表路由 | Handler | Service / 用例 |
| --- | --- | --- | --- |
| 学生赛事 | `GET /api/v1/contests`、`GET /contests/:id`、`POST /contests/:id/register` | `Handler`、`ParticipationHandler` | contest query、participation command/query |
| 学生队伍 | `/contests/:id/teams`、`/my-team`、`/teams/:tid/join` | `TeamHandler` | team command/query |
| 学生题目与提交 | `/contests/:id/challenges`、`/challenges/:cid/submissions` | `ChallengeHandler`、`SubmissionHandler` | challenge query、submission service |
| 学生榜单和公告 | `/contests/:id/scoreboard`、`/announcements`、WS `/ws/contests/:id/*` | `Handler`、`ParticipationHandler`、`RealtimeHandler` | scoreboard query、announcement query、realtime relay |
| 管理端赛事 | `/api/v1/admin/contests`、`/:id/freeze`、`/:id/unfreeze` | `Handler` | contest command/query |
| 管理端竞赛导出 | `/api/v1/admin/contests/:id/export` | `assessment.ReportHandler.CreateContestExport` | URL 挂在 contest 管理路由下，报告生命周期和 `reports` 表 owner 是 `assessment` |
| 管理端编排 | `/admin/contests/:id/challenges`、`/registrations`、`/announcements`、`/scoreboard/live` | challenge / participation / scoreboard handlers | contest challenge、participation、scoreboard admin |
| AWD 管理 | `/admin/contests/:id/awd/rounds`、`/services`、`/checker-preview`、`/traffic/*`、`/attacks` | `AWDHandler` | AWD command/query/jobs |
| AWD 学生侧 | `/contests/:id/awd/workspace`、`/awd/services/:sid/submissions` | `AWDHandler`、`SubmissionHandler` | AWD workspace query、attack submit |

## Application / Service 设计

| Service / 子域 | 代码路径 | 负责 | 不负责 |
| --- | --- | --- | --- |
| Contest service | `contest/application/commands/contest_service.go`、`queries/contest_service.go` | 赛事创建、更新、冻结、列表详情 | 报名/队伍子生命周期 |
| Participation service | `participation_service.go`、`participation_*` | 报名、审核、公告、我的进度 | 队伍成员命令 |
| Team service | `team_service.go`、`team_*` | 建队、入队、退队、踢人、队伍查询 | 用户身份事实 |
| Challenge service | `challenge_service.go`、`challenge_*` | 题目入赛、可见题目、管理员题目列表 | challenge 源数据维护 |
| Submission service | `submission_service.go`、`submission_*` | 竞赛提交、判分、错误限流、榜单同步 | practice 训练提交 |
| Scoreboard service | `scoreboard_*` | 榜单查询、冻结、改分、重建 | 训练排行榜 |
| AWD service | `awd_*.go`、`contest_awd_service_*` | AWD 轮次、服务、readiness、攻击、preview、状态缓存 | 容器 concrete |
| AWD jobs | `contest/application/jobs/awd_*` | checker runner、round updater、状态回写、锁续约 | HTTP 路由权限 |
| Status jobs | `status_*` | 竞赛状态推进和 side effect | 手动业务命令 |

## 数据设计

| 表 / 存储 | Entity / Adapter | Owner 语义 | 主要写入方 |
| --- | --- | --- | --- |
| `contests` | `contest/entity.Contest` | 赛事基本信息、状态、时间窗口和配置 | contest service |
| `contest_registrations` | `ContestRegistration` | 报名申请和审核状态 | participation service |
| `teams`、`team_members` | `Team`、`TeamMember` | 队伍和成员生命周期 | team service |
| `contest_challenges` | `ContestChallenge` | 题目入赛关系、分值和排序 | contest challenge service |
| `submissions` | `Submission` | 竞赛提交、人工评阅状态和得分 | contest submission service |
| `contest_announcements` | `ContestAnnouncement` | 公告事实 | participation announcement service |
| `contest_status_transitions` | `ContestStatusTransition` | 状态迁移历史和 side effect 状态 | status machine / jobs |
| `contest_realtime_outbox` | `ContestRealtimeOutbox` | 竞赛实时事件可恢复投递 | realtime broadcaster |
| `contest_runtime_placements` | `ContestRuntimePlacement` | contest 绑定 runtime node 的 placement 事实 | AWD / runtime placement adapter |
| `contest_awd_services` | `ContestAWDService` | AWD 服务快照、score/runtime/checker 配置和 preview 状态 | AWD service |
| `awd_rounds`、`awd_team_services` | `AWDRound`、`AWDTeamService` | AWD 轮次和队伍服务检查/计分事实 | AWD commands / jobs |
| `awd_attack_logs`、`awd_traffic_events` | `AWDAttackLog`、`AWDTrafficEvent` | 攻击提交与防守流量证据 | AWD attack / proxy traffic |
| `awd_defense_workspaces`、`awd_service_operations`、`awd_scope_controls` | AWD workspace / operation / scope control entities | 防守工作区、服务操作审计、禁用/退休/抑制范围 | practice + contest AWD |
| Redis scoreboard / AWD state | `scoreboard_state_store.go`、`awd_round_state_store.go` | 榜单 sorted-set、冻结快照、current round、service status runtime state | scoreboard / AWD service |
| Redis preview / rate limit / locks | `awd_checker_preview_token_store.go`、`submission_rate_limit_store.go`、`status_update_lock_store.go` | checker preview token、错误提交限流、状态推进锁 | contest application |

## 边界

- `contest` 拥有赛事、报名、队伍、排行榜、公告、contest challenge binding、AWD round/service/runtime placement 和比赛提交事实。
- `challenge` 拥有题目与镜像源数据；contest 只保存入赛关系、快照或比赛侧配置。
- `practice` 负责实际训练/运行实例编排；contest 持有 AWD scope 和 placement 事实。
- `ops` 负责 WebSocket relay 和通知 fanout；contest 产出 realtime 事件。
- `container_runtime` 提供 checker / preview / runtime 执行能力，不拥有 contest 状态。

## 主要用例

- 创建和维护赛事、公告、题目编排。
- 学生报名、组队、入队、退队、审核。
- 竞赛提交 Flag、错误提交限流、分数计算和排行榜读写。
- 管理员冻结榜单、改分、重建榜单。
- AWD 开赛就绪门禁、轮次推进、服务状态检查、checker preview、攻击提交、攻击日志和流量查询。
- Realtime WebSocket 推送公告、榜单和 AWD preview 进度。

## 数据与副作用

- PostgreSQL：contest、registration、team、contest_challenge、submission、scoreboard、announcement、AWD round/service/traffic/attack/realtime outbox 等事实。
- Redis：scoreboard state、AWD round state、checker preview token、submission rate limit、status update lock、realtime stream。
- Docker / sandbox：checker runner 和 preview 通过 runtime / checker adapter 执行。
- Outbox / WebSocket：contest realtime event 由 `ops` relay 推送。

## 跨模块依赖

| 依赖 | 用途 | 接线 |
| --- | --- | --- |
| `challenge` | 题目入赛、镜像/运行定义、AWD challenge lookup | `composition.BuildContestModule` |
| `container_runtime` | checker、preview、容器文件/探针能力 | `composition.BuildContestModule` |
| `auth` | WebSocket token service | `code/backend/internal/app/router.go` |
| `identity` | 用户和角色上下文 | middleware / route deps |
| `instance` | AWD 实例访问和 runtime recovery adapter | app composition adapter |
| `ops` | realtime relay 与通知 | `contest` events -> `ops` |

## Guardrail

- `code/backend/internal/module/contest/architecture_test.go`：约束 API / commands / queries / ports 分层、禁止 Docker concrete 落入普通 infrastructure 路径。
- `code/backend/internal/module/contest/ports/*_context_contract_test.go`：约束 state store 和端口 context 传播。
- `code/backend/internal/app/full_router_contest_state_matrix_test.go`、`full_router_awd_state_matrix_test.go`：覆盖赛事和 AWD 路由状态矩阵。
- `code/backend/internal/module/contest/infrastructure/contest_status_update_repository_test.go`：覆盖状态 CAS、transition journal 和 realtime outbox rollback。
- `code/backend/internal/module/contest/application/commands/awd_service_attack_test.go`、`code/backend/internal/module/contest/application/jobs/awd_round_updater_test.go`：覆盖 AWD Flag 读取、轮换、注入和回滚。
- `docs/architecture/features/校园级CTF-AWD模式完整设计.md`：AWD 总览事实源。

## 已知限制

- `contest` 是当前最大的业务 owner；AWD 子域仍在同一模块内，通过 commands / queries / jobs / ports 拆分控制复杂度。
