# 后端模块边界目标设计稿

> 状态：Draft
> 事实源：`code/backend/internal/module/` 当前实现、`docs/architecture/backend/07-modular-monolith-refactor.md` 当前事实、模块边界复核结论
> 替代：无；迁移完成后应回收到 `docs/architecture/backend/07-modular-monolith-refactor.md`

## 定位

本文档说明后端模块化单体的目标边界，不把当前代码形态直接视为合理终态。

- 负责：定义目标模块划分、组合方式、依赖方向、对外暴露口径、已知技术债和迁移切片。
- 不负责：记录当前已落地事实、替代当前架构事实源、设计微服务拆分方案、改写外部 HTTP 路由或论文正文。

## 当前设计判断

- 当前后端仍应保持单进程模块化单体。
  - 负责：继续用一个 Go API 进程部署，降低校园内网场景的运维成本。
  - 不负责：提前拆成微服务、引入服务注册发现、跨服务链路治理或分布式事务。

- 目标架构应按业务 owner 划分模块，而不是按页面、角色或历史目录命名。
  - 负责：让每条写路径、状态机、权限判断、重试策略和副作用都有唯一 owner。
  - 不负责：为了目录整齐给每个模块机械创建空 `domain`、空 `ports` 或空 `contracts`。

- `readmodel` 只用于跨 owner 只读聚合。
  - 负责：承接教师视角、复盘视角、跨表统计和页面聚合查询。
  - 不负责：拥有写侧状态、执行业务状态流转，或成为绕开 owner contract 的万能查询仓库。

- 容器运行能力应从业务实例生命周期中拆清。
  - 负责：区分“实例状态与访问权”这种业务事实，以及“Docker 网络、容器、ACL、文件读写”这种运行时适配。
  - 不负责：继续让一个宽泛 `runtime` 同时承担实例写模型、Docker 适配、教师查询、代理 ticket 和 AWD 文件操作。

## 目标模块版图

| 目标模块 | 类型 | 负责 | 不负责 | 对外暴露 |
| --- | --- | --- | --- | --- |
| `identity` | 写模型 | 用户、角色、账号状态、资料、密码哈希与密码变更、管理端用户能力 | session、token、CAS、WebSocket ticket | 用户查询、用户写入、凭据校验、资料命令/查询 contract，管理端 handler |
| `auth` | 认证能力 / 写模型 | 登录、登出、会话、CAS、WebSocket ticket、认证中间件所需 token 能力 | 用户资料 owner、用户仓储实现、管理端用户 CRUD | 认证 handler、session/token contract、认证上下文构造能力 |
| `challenge` | 写模型 | 题目元数据、题包、附件、镜像引用、Flag 策略、Writeup、拓扑模板 | 实例状态、容器生命周期、竞赛计分 | 题目 catalog、Flag validator、image store、题目管理 handler |
| `instance` | 写模型 | 实例记录、排队、状态机、启动/续期/销毁、访问 ticket、实例清理和实例可见性查询 | Docker SDK、题目元数据 owner、提交计分 | 实例 command/query contract、实例 handler、调度 background job |
| `container_runtime` | 平台适配 | Docker Engine、网络、ACL、容器文件读写、探活、镜像探测、资源限制 | 用户权限、竞赛规则、实例业务状态 | container engine ports 的 adapter；默认不暴露 HTTP handler |
| `practice` | 写模型 | 日常训练、练习提交、个人解题状态、练习排行榜、人工评审入口 | 容器底层执行、题目定义、能力画像写入 | 练习 handler、提交/解题事件、练习 query contract |
| `contest` | 写模型 | 竞赛、报名、队伍、题目入赛、排行榜、冻结榜、公告、AWD 轮次、AWD 攻击计分 | 题目定义、容器底层执行、通知发送实现 | 竞赛 handler、AWD handler、排行榜 query、竞赛事件 |
| `assessment` | 分析产物写模型 | 能力画像、推荐、报告、复盘归档、评估重建任务 | 提交判定、竞赛计分主链路、教师页面聚合查询 | 画像/推荐/报告 contract、报告 handler、事件消费者 |
| `ops` | 运营支撑 | 审计、通知、WebSocket 广播适配、运行概览、风险视图 | 业务状态 owner、业务规则决策 | audit recorder、notification handler、realtime broadcaster adapter |
| `teaching_readmodel` | 读模型 | 教师班级、学生证据链、复盘、跨训练/竞赛/评估的只读聚合 | 写入训练/竞赛/评估事实、替代 owner 规则 | 教师端 query handler、只读 query service |

### 命名说明

- `instance` 是建议目标名，用来承接“靶机实例业务生命周期”。如果迁移成本较高，可先保留 `runtime` 包名作为兼容 facade，再逐步抽出 `instance` 与 `container_runtime`。
- `container_runtime` 可以落在 `internal/module/container_runtime`，也可以落在 `internal/platform/container`。选择标准不是目录偏好，而是它是否承载业务状态：如果只实现 Docker/ACL/文件等适配，优先放平台适配层；如果还拥有实例表和调度状态，就不应叫 `container_runtime`。

## 目标依赖方向

```mermaid
flowchart LR
    API["api / handler"] --> APP["application"]
    APP --> DOMAIN["domain"]
    APP --> PORTS["consumer-side ports"]
    INFRA["infrastructure adapters"] -.implements.-> PORTS
    COMPOSITION["composition root"] --> API
    COMPOSITION --> INFRA

    AUTH["auth"] --> IDENTITY["identity contracts"]
    PRACTICE["practice"] --> CHALLENGE["challenge contracts"]
    PRACTICE --> INSTANCE["instance contracts"]
    CONTEST["contest"] --> CHALLENGE
    CONTEST --> INSTANCE
    INSTANCE --> CONTAINER["container runtime ports"]
    ASSESSMENT["assessment"] -.consumes events / queries.-> PRACTICE
    ASSESSMENT -.consumes events / queries.-> CONTEST
    TEACHING_RM["teaching_readmodel"] --> PRACTICE
    TEACHING_RM --> CONTEST
    TEACHING_RM --> ASSESSMENT
    OPS["ops"] -.subscribes events / implements adapters.-> AUTH
    OPS -.subscribes events / implements adapters.-> CONTEST
```

目标规则：

- `auth -> identity contracts` 可以存在；`identity -> auth` 不应存在。
- `practice` 和 `contest` 可以依赖 `challenge`、`instance` 的 contract；不能依赖它们的 `infrastructure`。
- `instance` 通过 container runtime ports 调用 Docker/ACL/文件能力；业务模块不能直接依赖 Docker SDK。
- `assessment` 不应挂在提交计分的同步写路径上更新画像；优先消费训练/竞赛事件或通过显式重建任务收敛。
- `ops` 不应成为业务模块的硬依赖；业务模块发布事件或使用窄 port，通知和 WebSocket 广播由 `ops` 适配。
- `teaching_readmodel` 可以跨 owner 读取只读事实，但不能反向回写 owner 表，也不能承载状态转换规则。

## 对外暴露规则

### 模块暴露什么

每个模块对外只暴露四类能力：

1. `api/http` handler：只给路由层挂载。
2. `contracts`：其他模块可依赖的稳定业务能力或事件类型。
3. `ports`：消费方定义的最小能力接口，由 composition 绑定具体 adapter。
4. `runtime.Module` 输出：模块装配后的 handler、contract implementation、background job 和 closer。

### 模块不暴露什么

- 不向其他模块暴露 `infrastructure.Repository` 具体类型。
- 不把 GORM、Redis、Docker client、Gin context 放进跨模块 contract。
- 不把 API DTO 当成跨模块内部 contract；跨模块 contract 应使用模块自己的输入输出结构或领域值对象。
- 不把路由命名空间当模块边界，例如 `/teacher/*` 是外部接口分组，不是 `teacher` 写模型。

### 事件暴露

跨模块异步协作使用事件时，事件类型应放在 owner 模块的 `contracts` 或稳定事件包中：

- `practice` 发布 `SubmissionRecorded`、`ChallengeSolved`、`InstanceStarted` 等事实事件。
- `contest` 发布 `ContestSubmissionScored`、`ScoreboardUpdated`、`AWDAttackRecorded`、`ContestStatusChanged` 等事实事件。
- `challenge` 发布 `PublishCheckFinished`、`ChallengePublished` 等事实事件。
- `assessment` 和 `ops` 作为消费者处理画像更新、缓存失效、通知、广播和审计补充。

事件不是事务替代品。强一致写路径仍由 owner application service 和 repository transaction 负责。

## 当前技术债判断

| 债务 | 当前风险 | 目标状态 | 迁移优先级 |
| --- | --- | --- | --- |
| `runtime` 职责过宽 | 实例状态、Docker 适配、proxy ticket、教师查询、AWD 文件能力混在一个 owner | 拆成 `instance` 业务 owner 与 `container_runtime` 适配能力 | 高 |
| `assessment / ops` 事件化边界仍未完全收口 | `practice` 画像链已切到事件消费，但通知、广播、缓存失效和其他副作用仍有继续统一 owner 表达的空间 | 副作用默认经事件或窄 port 触发，避免业务写路径继续同步背负实现细节 | 中 |
| `contest -> ops` 实时广播耦合 | 竞赛业务服务知道 WebSocket 适配细节 | `contest` 使用 broadcaster port 或事件，`ops` 实现适配 | 中 |
| application 层 GORM/Redis allowlist 多 | 用例层仍暴露框架和存储实现，影响测试和迁移 | 用 ports 包装事务、缓存、锁和查询能力 | 中 |
| readmodel repository 过宽 | 容易成为跨表万能仓库 | 按教师目录、学生证据、班级洞察等 query capability 拆小 | 低到中 |

## 迁移切片建议

### 阶段 1：先收口认证与身份边界

当前状态（2026-05-11）：

- 已完成首个迁移切片。
- `identitycontracts.Authenticator` 与 `identity -> auth` allowlist 已删除。
- token service 现在由 `code/backend/internal/app/router.go` 统一创建，并传给 `auth` runtime、认证中间件、通知 WS 和竞赛实时 WS。

目标：

- `identity` 不再导入 `auth/contracts`。
- `auth` 保留 session、token、CAS、WS ticket owner。
- `identity` 只暴露用户、凭据、资料和管理能力。

建议动作：

1. 在 `auth/contracts` 或更中性的 `internal/authctx`/`internal/shared/authn` 明确 token/session contract。
2. 删除 `identitycontracts.Authenticator` 对 `authcontracts.TokenService` 的包装关系。
3. composition 直接把 auth token service 传给认证中间件和需要认证能力的适配器。
4. 增加架构测试禁止 `identity -> auth`。

### 阶段 2：拆清实例业务与容器适配

目标：

- 实例状态、调度、访问 ticket、续期和销毁归 `instance`。
- Docker Engine、网络、ACL、文件、探活归 `container_runtime` 或平台适配层。
- `practice`、`contest` 只依赖 `instance` contract，不知道 Docker 细节。

当前状态（2026-05-12，phase 2 / slice 14）：

- `internal/module/instance/` 已经落地 `application/commands`、`application/queries`、`ports`、`domain`，实例命令、实例查询、proxy ticket 和 maintenance use case 已有独立物理 owner。
- `code/backend/internal/app/composition/instance_module.go` 现在直接装配 `instancecmd.NewInstanceService`、`instanceqry.NewInstanceService`、`instanceqry.NewProxyTicketService`、`instancecmd.NewInstanceMaintenanceService`，并把它们接到 runtime repo 与显式 capability adapter。
- `runtime_cleaner` 与 AWD defense SSH gateway 已经从 `runtime/runtime.Module` 上移到 `InstanceModule` 注册；用户实例路由、教师实例路由、AWD target proxy 和 defense SSH 入口继续统一通过 `InstanceModule.Handler` 挂载。
- app 层已经把 challenge / contest / ops 依赖的容器能力显式命名为 `ContainerRuntimeModule`；`BuildChallengeModule`、`BuildContestModule`、`BuildOpsModule`、`BuildInstanceModule` 都已经改依赖这个视图。
- `runtime/runtime/module.go` 不再生产装配 instance command/query、proxy ticket 或 maintenance service，只保留 container-facing builder 与 practice/challenge/ops/contest 仍需复用的显式 runtime capability fields，不再向上暴露宽 `Engine`。
- `code/backend/internal/app/practice_flow_integration_test.go` 与 `code/backend/internal/module/runtime/service_test.go` 已继续切到 `instance/*` owner，减少了外部直接 new compat service 的调用点。
- `internal/module/instance/contracts` 已经落地；生产使用的 runtime HTTP adapter 已收口到 `internal/app/composition/runtime_http_service_adapter.go`，`runtime/runtime/adapters.go` 只保留 practice / challenge / ops 仍在复用的底层 adapter，不再平行保留一份 runtime HTTP adapter。
- `runtime/application/*` 中原本保留的 instance / proxy ticket / maintenance compat wrapper 已删除；实例命令、查询、proxy ticket、maintenance 的唯一 owner 固定在 `instance`。
- 原本放在 `runtime/application` 目录里的实例行为测试已经切到 `instancecmd` / `instanceqry`，compat 层只保留最小 wrapper 测试。
- `runtime/application` 中仍保留的 provisioning / cleanup / container file / image / stats service，已经统一依赖 `runtime/ports/container_runtime.go` 里的 container runtime ports；`runtime/runtime.Module` 现在只暴露 `ProvisioningRuntime`、`CleanupRuntime`、`FileRuntime`、`ManagedContainerInventory`、`InteractiveExecutor` 等显式能力字段，maintenance 需要的 inspect/start 组合留在 composition 边缘完成。
- `code/backend/internal/app/composition/instance_module.go` 现在直接基于 `runtimeinfra.NewRepository(...)` 与本地 practice runtime adapter 暴露 `PracticeInstanceRepository`、`PracticeRuntimeService`；`runtime/runtime.Module` 与 `runtime/runtime/adapters.go` 已不再 import `practice/ports`，`runtime -> practice` allowlist 也已删除。

建议动作：

1. 继续判断 `runtime/ports/container_runtime.go` 这组 capability port 的最终物理落点，是继续留在 `runtime` 过渡，还是后续随 `container_runtime` 物理模块一起迁出。
2. 如果未来确实再次出现兼容 import path 需求，需要重新评估边界，而不是默认恢复旧 wrapper。

### 阶段 3：事件化评估与运营副作用

目标：

- 训练和竞赛写路径不被画像、通知、广播实现拖住。
- `assessment`、`ops` 作为消费者处理可重试副作用。

当前状态（2026-05-12，phase 3 / slices 1-3）：

- `practice` runtime / composition 对 `assessment.ProfileService` 的直接注入已删除。
- 正确提交与人工评审通过后的能力画像增量更新，统一通过 `practice.flag_accepted` 事件交给 `assessment` 消费。
- 题目发布自检完成后的教师通知，统一通过 `challenge.publish_check_finished` 事件交给 `ops` 消费。
- 竞赛公告创建/删除、榜单刷新、AWD 预览进度，统一通过 `contest` 事件交给 `ops` relay 做 WebSocket 广播。
- phase 3 里与缓存失效和其他副作用相关的统一收口还没有全部结束。

建议动作：

1. 明确 practice/contest/challenge 的业务事件 contract。
2. 将画像失效、推荐缓存失效、通知发送、WebSocket 广播改为事件消费者或窄 port adapter。
3. 对关键用户可见副作用保留同步 fallback 或失败记录，避免静默丢失。

### 阶段 4：复核 readmodel

当前状态（2026-05-13，phase 4 / slice 4）：

- `/api/v1/users/me/progress` 与 `/api/v1/users/me/timeline` 已并回 `practice/application/queries` 与 `practice/api/http`。
- `internal/module/practice_readmodel/` 已删除，因为这两条查询只读取 `practice` 自有事实，不构成跨 owner readmodel。
- 当前 readmodel 只保留 `teaching_readmodel` 这类真实跨 owner 聚合入口。
- 教师总览 `GetOverview` 已从宽 `teaching_readmodel/application/queries.Service` 中拆到独立 `OverviewService`；`teaching_readmodel/api/http.Handler` 不再通过单个大一统 query 接口承接 overview。
- 班级详情 `GetClassSummary`、`GetClassTrend`、`GetClassReview` 已继续从剩余宽 `Service` 中拆到独立 `ClassInsightService`；`teaching_readmodel/api/http.Handler` 现在分别依赖 `Service`、`OverviewService`、`ClassInsightService`。
- 学生复盘 `GetStudentProgress`、`GetStudentRecommendations`、`GetStudentTimeline`、`GetStudentEvidence`、`GetStudentAttackSessions` 已从剩余宽 `Service` 中拆到独立 `StudentReviewService`；`teaching_readmodel/api/http.Handler` 现在分别依赖目录 `Service`、`OverviewService`、`ClassInsightService`、`StudentReviewService`。

目标：

- `teaching_readmodel` 保留为教师视角跨 owner 聚合。
- 如果未来个人进度/时间线再次跨 owner，再重新评估是否需要专门 readmodel。

建议动作：

1. 逐个查询标注数据来源和 UI consumer。
2. 纯 practice 查询继续留在 `practice/application/queries`。
3. 跨 owner 查询保留 `teaching_readmodel`，并继续评估目录查询是否还要进一步显式命名为更窄 owner，但不再把学生复盘回挂到目录 query surface。

### 阶段 5：收窄 application concrete allowlist

当前状态（2026-05-13，phase 5 / slices 1-35）：

- `auth/application/commands/cas_service.go` 现在通过 `auth/ports.CASTicketValidator` 校验 CAS ticket；`auth/infrastructure/cas_ticket_validator.go` 统一承接 CAS validate request、XML principal 解析、用户名校验和 invalid ticket sentinel，`auth/runtime/module.go` 也不再把 `net/http` concrete 留在 auth command service。
- `code/backend/internal/module/architecture_allowlist_test.go` 已删除 `auth/application/commands/cas_service.go -> net/http` 这条例外。
- `challenge/application/queries/challenge_service.go` 里的 solved-count 缓存已通过 `challenge/ports.ChallengeSolvedCountCache` 下沉到模块内 infrastructure Redis adapter。
- `code/backend/internal/module/architecture_allowlist_test.go` 已删除 `challenge/application/queries/challenge_service.go -> github.com/redis/go-redis/v9` 这条例外。
- `challenge/application/queries/challenge_service.go` 现在通过 `challenge/ports.ErrChallengeQueryChallengeNotFound` 识别 challenge lookup not-found 语义；`challenge/infrastructure/challenge_query_repository.go` 统一承接 raw challenge repository 的 `gorm.ErrRecordNotFound` 到模块内 sentinel 的映射，challenge query surface 不再直接 import GORM sentinel。
- `code/backend/internal/module/architecture_allowlist_test.go` 已删除 `challenge/application/queries/challenge_service.go -> gorm.io/gorm` 这条例外。
- `challenge/application/commands/challenge_service.go` 现在通过 `challenge/ports.ErrChallengeCommandChallengeNotFound`、`ErrChallengePublishCheckJobNotFound`、`ErrChallengeImageNotFound` 与 `ErrChallengeTopologyNotFound` 识别题目、发布自检任务、镜像与拓扑 lookup 的 not-found 语义；`challenge/infrastructure/challenge_command_repository.go` 统一承接 raw challenge repository 的 `FindByID` / publish-check lookup `gorm.ErrRecordNotFound` 到模块内 sentinel 的映射，`challenge/runtime/module.go` 只把 challenge command adapter、image query adapter 与 topology service adapter 注入 core command service，`challenge_service.go` 不再直接 import GORM sentinel。
- `code/backend/internal/module/architecture_allowlist_test.go` 已删除 `challenge/application/commands/challenge_service.go -> gorm.io/gorm` 这条例外。
- `challenge/application/commands/image_build_service.go` 现在通过 `challengeports.RegistryVerifier` 校验 external image manifest；`challenge/infrastructure/registry_client.go` 统一承接 registry URL、认证头、accept header 和 digest 提取的 HTTP 细节，`challenge/runtime/module.go` 也不再从 application 包构造 registry verifier。
- `code/backend/internal/module/architecture_allowlist_test.go` 已删除 `challenge/application/commands/registry_client.go -> net/http` 这条例外。
- `challenge/application/queries/image_service.go` 现在通过 `challenge/ports.ErrChallengeImageNotFound` 识别镜像 not-found 语义；`challenge/infrastructure/image_query_repository.go` 统一承接 raw image repository 的 `gorm.ErrRecordNotFound` 到模块内 sentinel 的映射，image query surface 不再直接 import GORM sentinel。
- `challenge/application/commands/image_service.go` 现在通过 `challenge/ports.ErrChallengeImageNotFound` 识别 image lookup not-found 语义；`challenge/infrastructure/image_command_repository.go` 统一承接 raw image repository 的 `FindByID` / `FindByNameTag` `gorm.ErrRecordNotFound` 到模块内 sentinel 的映射，image command surface 不再直接 import GORM sentinel。
- `code/backend/internal/module/architecture_allowlist_test.go` 已删除 `challenge/application/commands/image_service.go -> gorm.io/gorm` 这条例外。
- `challenge/application/commands/flag_service.go` 与 `challenge/application/queries/flag_service.go` 现在通过 `challenge/ports.ErrChallengeFlagChallengeNotFound` 识别 challenge lookup not-found 语义；`challenge/infrastructure/flag_repository.go` 统一承接 raw challenge repository 的 `gorm.ErrRecordNotFound` 到模块内 sentinel 的映射，flag command/query surface 不再直接 import GORM sentinel。
- `challenge/application/commands/awd_challenge_service.go` 与 `challenge/application/queries/awd_challenge_service.go` 现在通过 `challenge/ports.ErrAWDChallengeNotFound` 识别 AWD challenge lookup not-found 语义；`challenge/infrastructure/awd_challenge_repository.go` 统一承接 raw AWD challenge repository 的 `gorm.ErrRecordNotFound` 到模块内 sentinel 的映射，AWD challenge command/query surface 不再直接 import GORM sentinel。
- `challenge/application/commands/writeup_service.go` 与 `challenge/application/queries/writeup_service.go` 现在通过 `challenge/ports.ErrChallengeWriteupChallengeNotFound`、`ErrChallengeWriteupRequesterNotFound`、`ErrChallengeOfficialWriteupNotFound`、`ErrChallengeReleasedWriteupNotFound`、`ErrChallengeSubmissionWriteupNotFound`、`ErrChallengeSubmissionWriteupDetailNotFound` 与 `ErrChallengeTeacherSubmissionWriteupNotFound` 识别 writeup 相关 lookup 的 not-found 语义；`challenge/infrastructure/writeup_service_repository.go` 统一承接 raw challenge / user / writeup repository 的 `gorm.ErrRecordNotFound` 到模块内 sentinel 的映射，writeup command/query surface 不再直接 import GORM sentinel。
- `challenge/application/commands/topology_service.go` 与 `challenge/application/queries/topology_service.go` 现在通过 `challenge/ports.ErrChallengeTopologyChallengeNotFound`、`ErrChallengeTopologyNotFound`、`ErrChallengeTopologyTemplateNotFound` 与 `ErrChallengeTopologyPackageRevisionNotFound` 识别 challenge / topology / template / package revision lookup 的 not-found 语义；`challenge/infrastructure/topology_service_repository.go` 统一承接 raw challenge / topology / template / package revision repository 的 `gorm.ErrRecordNotFound` 到模块内 sentinel 的映射，topology command/query surface 不再直接 import GORM sentinel。
- `code/backend/internal/module/architecture_allowlist_test.go` 已删除 `challenge/application/commands/topology_service.go -> gorm.io/gorm` 与 `challenge/application/queries/topology_service.go -> gorm.io/gorm` 这两条例外。
- `contest/application/commands/challenge_service.go` 里未使用的 Redis 注入链已删除，不再把无效 cache client 传入 contest challenge command service。
- `contest/application/statusmachine/side_effects.go` 现在只负责编排冻结榜快照创建、解冻快照清理和比赛结束缓存清理，具体 Redis key / client 细节已通过 `contest/ports.ContestStatusSideEffectStore` 下沉到模块内 infrastructure adapter。
- `contest/application/commands/contest_service.go` 不再为状态迁移副作用链持有 Redis client；`contest/application/jobs/status_updater.go` 的状态调度锁也已通过 `contest/ports.ContestStatusUpdateLockStore` 下沉到 infrastructure adapter，application/jobs 只保留持锁编排与 keepalive 语义。
- `contest/application/jobs/AWDRoundUpdater` 当前通过 `contest/ports.AWDRoundStateStore` 与 `contest/infrastructure/awd_round_state_store.go` 承接 scheduler lock、round lock、current round pointer、round flags 和 live service status cache；application/jobs 不再直接知道 Redis key、`redis.Nil`、pipeline 或 `redislock.Acquire(...)`。
- `contest/application/queries/scoreboard_service.go` 与 `contest/application/commands/scoreboard_admin_service.go` 现在通过 `contest/ports.ContestScoreboardStateStore` 读取和更新排行榜状态；`contest/infrastructure/scoreboard_state_store.go` 统一承接 live/frozen scoreboard 列表、frozen snapshot create/clear、team rank、分数增量和全量 rebuild 的 Redis sorted-set 细节，`contest/runtime/module.go` 也不再把 Redis client 直接注回 scoreboard application surface。
- `contest/application/commands/submission_service.go` 现在通过 `contest/ports.ContestSubmissionRateLimitStore` 读取和写入错误提交限流状态；`contest/infrastructure/submission_rate_limit_store.go` 统一承接 configured prefix、默认 prefix 回退和 Redis `Exists/Set` 细节，`contest/runtime/module.go` 也不再把 Redis client 直接注回 submission application surface。
- `contest/application/commands/submission_submit_validation.go` 现在通过 `contest/ports.ErrContestSubmissionChallengeNotFound` 与 `ErrContestSubmissionChallengeEntityNotFound` 识别 contest challenge lookup 与 challenge entity lookup 的 not-found 语义；`contest/infrastructure/submission_registration_repository.go` 统一承接 raw contest challenge / challenge repository 的 `gorm.ErrRecordNotFound` 到模块内 sentinel 的映射，submission validation surface 不再直接 import GORM sentinel。
- `code/backend/internal/module/architecture_allowlist_test.go` 已删除 `contest/application/commands/submission_submit_validation.go -> gorm.io/gorm` 这条例外。
- `contest/application/queries/awd_support.go` 与 `contest/application/queries/awd_workspace_query.go` 现在通过 `contest/ports.ErrContestAWDRoundNotFound` 与 `ErrContestUserTeamNotFound` 识别 round/team lookup not-found 语义；`contest/infrastructure/awd_query_repository.go` 统一承接 raw AWD repository 的 `FindRoundByContestAndID` / `FindRunningRound` / `FindContestTeamByMember` `gorm.ErrRecordNotFound` 到模块内 sentinel 的映射，query `AWDService` 不再直接 import GORM sentinel。
- `code/backend/internal/module/architecture_allowlist_test.go` 已删除 `contest/application/queries/awd_support.go -> gorm.io/gorm` 与 `contest/application/queries/awd_workspace_query.go -> gorm.io/gorm` 这两条例外。
- `contest/application/commands/participation_register_commands.go`、`contest/application/commands/participation_review_commands.go`、`contest/application/commands/submission_validation.go` 与 `contest/application/queries/participation_progress_query.go` 现在通过 `contest/ports.ErrContestParticipationRegistrationNotFound` 和 `contest/ports.ErrContestUserTeamNotFound` 识别报名 lookup、用户队伍 lookup 的 not-found 语义；`contest/infrastructure/participation_registration_repository.go`、`contest/infrastructure/submission_registration_repository.go` 与 `contest/infrastructure/team_finder_repository.go` 统一承接 raw participation/submission/team repository 的 `gorm.ErrRecordNotFound` 到模块内 sentinel 的映射，`contest/runtime/module.go` 负责把这些 adapter 注入 participation / submission wiring。
- `contest/application/queries/team_info_query.go` 与 `contest/application/queries/team_list_query.go` 现在通过 `contest/ports.ErrContestTeamNotFound` 和 `contest/ports.ErrContestUserTeamNotFound` 识别队伍详情、用户当前队伍的 not-found 语义；`contest/infrastructure/team_query_adapter.go` 统一承接 raw team repository 的 `FindByID` / `FindUserTeamInContest` not-found 映射，`contest/runtime/module.go` 只把它注入 team query wiring，team command surface 继续保留 raw team repository。
- `contest/application/commands/team_captain_manage_commands.go`、`team_create_retry_support.go`、`team_join_commands.go`、`team_leave_commands.go` 与 `team_support.go` 现在通过 `contest/ports.ErrContestTeamNotFound`、`ErrContestUserTeamNotFound` 与 `ErrContestParticipationRegistrationNotFound` 识别队伍 lookup、当前队伍 lookup 与报名 lookup 的 not-found 语义；`contest/infrastructure/team_command_adapter.go` 统一承接 raw team repository 的 `FindByID` / `FindUserTeamInContest` / `FindContestRegistration` 和 `CreateWithMember` / `AddMemberWithLock` 里的 registration binding `gorm.ErrRecordNotFound` 映射，`contest/runtime/module.go` 负责把该 adapter 注入 command `TeamService`。
- `contest/application/commands/awd_preview_runtime_support.go` 现在通过 `contest/ports.ErrContestAWDPreviewChallengeNotFound` 与 `ErrContestAWDPreviewImageNotFound` 识别 preview challenge / image lookup 的 not-found 语义；`contest/infrastructure/awd_preview_runtime_lookup_repository.go` 统一承接 raw AWD challenge / image repository 的 `gorm.ErrRecordNotFound` 到模块内 sentinel 的映射，AWD preview runtime surface 不再直接 import GORM sentinel。
- `code/backend/internal/module/architecture_allowlist_test.go` 已删除 `contest/application/commands/awd_preview_runtime_support.go -> gorm.io/gorm` 这条例外。
- `AWDService` 与 `contest_awd_service_service` 现在通过 `contest/ports.AWDRoundStateStore` 和 `contest/ports.AWDCheckerPreviewTokenStore` 读取 current round / round flag / service status runtime state，并承接 checker preview token 的存取；`contest/infrastructure/awd_round_state_store.go` 与 `contest/infrastructure/awd_checker_preview_token_store.go` 统一落地 Redis 细节，`contest/runtime/module.go` 也不再把 Redis client 直接注回 AWD application surface。
- 当前 `contest` application surface 的 Redis concrete allowlist 已收口完成。
- 当前 `contest` application surface 已完成 submission challenge lookup、AWD preview runtime lookup、AWD query round/team lookup、报名/进度、team query 与 team command 这六组 GORM concrete 收口。
- `practice/application/commands/score_service.go` 与 `practice/application/queries/score_service.go` 现在通过 `practice/ports.PracticeScoreStateStore` 读取和写入用户得分相关 Redis 状态；`practice/infrastructure/score_state_store.go` 统一承接用户计分锁、用户得分缓存和排行榜 sorted-set 细节，`practice/runtime/module.go` 也不再把 Redis client 直接注回 practice score command/query surface。
- `practice/application/queries/score_service.go` 现在还通过 `practice/ports.ErrPracticeUserScoreNotFound` 识别用户尚无得分记录的 not-found 语义；`practice/infrastructure/score_query_repository.go` 统一承接 raw score repository 的 `gorm.ErrRecordNotFound` 到模块内 sentinel 的映射，practice score query surface 不再直接 import GORM sentinel。
- `practice/application/commands/service.go` 与 `practice/application/commands/submission_service.go` 现在通过 `practice/ports.PracticeFlagSubmitRateLimitStore` 读取和更新 flag submit 限流状态；`practice/infrastructure/submission_rate_limit_store.go` 统一承接 prefix fallback、计数 key、`Incr` 和首次窗口 `Expire` 细节，`practice/runtime/module.go` 也不再把 Redis client 直接注回 practice command service。
- `practice/application/commands/instance_provisioning.go` 现在通过 `practice/ports.PracticeInstanceReadinessProbe` 执行单次 access URL 探测；`practice/infrastructure/instance_readiness_probe.go` 统一承接 HTTP GET、TCP dial、`url.Parse` 和响应体回收细节，`practice/runtime/module.go` 也不再把 `net/http` concrete 留在 practice provisioning surface。
- `practice/application/commands/contest_instance_scope.go` 与 `contest_awd_operations.go` 现在通过 `practice/ports.PracticeContestScopeRepository` 和 `practice/ports.PracticeRuntimeSubjectRepository` 识别 contest / challenge / topology not-found 语义；`practice/infrastructure/contest_scope_repository.go` 与 `practice/infrastructure/runtime_subject_repository.go` 统一承接 raw practice repository、challenge contract 到模块内 sentinel 的映射，practice application surface 不再直接 import GORM sentinel。
- `practice/application/commands/manual_review_service.go` 现在通过 `practice/ports.PracticeManualReviewRepository` 和 `practice/ports.PracticeRuntimeSubjectRepository` 识别人工评阅提交、已通过提交、教师用户与 challenge runtime subject 的 not-found 语义；`practice/infrastructure/manual_review_repository.go` 与 `practice/infrastructure/runtime_subject_repository.go` 统一承接 raw practice repository、challenge contract 的 `gorm.ErrRecordNotFound` 到模块内 sentinel 的映射，manual review application surface 不再直接 import GORM sentinel。
- `practice/application/commands/submission_service.go` 现在通过 `practice/ports.PracticeSolvedSubmissionRepository` 和 `practice/ports.PracticeRuntimeSubjectRepository` 识别正确提交与 challenge runtime subject 的 not-found 语义；`practice/infrastructure/solved_submission_repository.go` 与 `practice/infrastructure/runtime_subject_repository.go` 统一承接 raw practice repository、challenge contract 的 `gorm.ErrRecordNotFound` 到模块内 sentinel 的映射，submission application surface 不再直接 import GORM sentinel。
- 当前 `practice` application surface 的 Redis / HTTP / GORM concrete allowlist 已收口完成；后续如果继续推进，重点转到其他模块里仍残留的 GORM concrete surface。
- `ops/application/queries/dashboard_service.go` 现在通过 `ops/ports.DashboardStateStore` 读取和写入 dashboard cache，并统计在线会话；`ops/infrastructure/dashboard_state_store.go` 统一承接 dashboard cache key、在线 session 扫描、session payload JSON 和 Redis `Get/Set/Scan/MGet` 细节，`ops/runtime/module.go` 也不再把 Redis client 直接注回 dashboard query surface。
- `ops/application/commands/notification_service.go` 现在通过 `ops/ports.ErrNotificationNotFound` 识别 repo not-found 语义；`ops/infrastructure/notification_repository.go` 统一承接 `gorm.ErrRecordNotFound` 到模块内错误契约的映射，`notification_service.go` 不再直接 import GORM sentinel。
- 当前 `ops` application surface 的 concrete allowlist 已收口完成。
- `assessment/application/commands/profile_service.go` 现在通过 `assessment/ports.AssessmentProfileLockStore` 获取画像锁；`assessment/infrastructure/state_store.go` 统一承接画像锁 key、TTL、`SetNX` 与 safe release 细节，`assessment/runtime/module.go` 也不再把 Redis client 直接注回 profile command surface。
- `assessment/application/queries/recommendation_service.go` 现在通过 `assessment/ports.AssessmentRecommendationCacheStore` 读取、写入和失效 recommendation cache；`assessment/infrastructure/state_store.go` 统一承接 recommendation cache key、JSON 编解码和 Redis `Get/Set/Del` 细节，`assessment/runtime/module.go` 也不再把 Redis client 直接注回 recommendation query surface。
- `assessment/application/commands/report_service.go` 现在通过 `assessment/ports.ErrAssessmentReportNotFound` 和 `assessment/ports.ErrAssessmentContestNotFound` 识别 repo not-found 语义；`assessment/infrastructure/report_repository.go` 统一承接 `gorm.ErrRecordNotFound` 到模块内错误契约的映射，report service 不再直接 import GORM sentinel。
- 当前 `assessment` application surface 的 concrete allowlist 已收口完成；phase 5 这一轮 `application` 层的 Redis / GORM concrete 例外已经全部清空，后续如果继续推进，重点转到其他模块里尚未收口的 GORM / HTTP concrete surface。
- 当前 `auth` application surface 已清空 HTTP concrete allowlist。
- 当前 `challenge` application surface 已清空 Redis / HTTP concrete allowlist，并继续删掉了 challenge query、challenge core command、image query、image command、flag command/query、AWD challenge command/query、writeup command/query 与 topology command/query 这十二条 GORM concrete 例外。

目标：

- application 不长期直接依赖 GORM、Redis、Docker、HTTP client。
- 事务、缓存、锁、外部调用都通过用例级 ports 表达。

建议动作：

1. 从高频变更模块开始处理：`contest`、`practice`、`challenge`。
2. 每次只收口一个 use case 或一个事务边界。
3. 每收口一个 allowlist 项，同步架构测试和最小行为测试。

## 迁移完成标准

- `code/backend/internal/module/architecture_allowlist_test.go` 中不再允许 `identity -> auth`。
- `runtime` 不再同时承担实例业务 owner 和 Docker adapter；旧 facade 已删除或只剩兼容薄层。
- `practice`、`contest` 不直接依赖 `ops` 具体实现，通知和广播通过事件或窄 port。
- `practice` 写路径不同步依赖 `assessment` 写服务；画像更新改为事件消费或显式重建。
- 每个模块的 `runtime.Module` 只暴露 handler、contract implementation、background job 和 closer，不暴露具体 repository。
- readmodel 模块只有只读聚合查询，不包含业务状态流转和写路径。
- application concrete dependency allowlist 明显收缩，并且新增项必须有 review 理由。

## 论文写作口径

在迁移完成前，论文只应写当前已落地的模块化单体事实，不能把本文档中的目标模块版图写成已实现架构。

迁移完成后，可以把论文第 3 章中的后端模块关系更新为：

- 认证与身份分离：`auth` 负责认证会话，`identity` 负责用户与凭据。
- 实例业务与容器适配分离：`instance` 负责实例生命周期，`container_runtime` 封装 Docker。
- 训练、竞赛、评估通过事件和 contract 协作，读模型负责教师复盘和跨模块聚合。
