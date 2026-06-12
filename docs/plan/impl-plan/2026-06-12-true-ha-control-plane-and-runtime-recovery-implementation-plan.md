<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# 真正高可用控制面与运行时恢复 Implementation Plan

**Goal:** 把 `ctf` 从“支持多实例部署”推进到“控制面强 HA + 执行面故障后自动重建”的可执行路线，明确代码、配置、文档和验证切片，避免继续停留在单机 Docker + 多 API 副本的半完成状态。

**Architecture:** 沿用当前 `runtime-agent + node_id + PostgreSQL + Redis` 的总体方向，不引入 live migration 或多主写数据库，而是把 HA 拆成四层 owner：状态面 HA、共享文件/密钥 owner、跨副本事件总线、runtime node 故障检测与实例重建。每一层都按最小可审查切片推进，先收口 correctness 和 shared-state，再做执行面自动恢复。

**Tech Stack:** Go, GORM, PostgreSQL, Redis Sentinel, Redis Stream / Outbox, runtime-agent, shared storage or object storage, code-workflow

---

## Task Metadata

- Task Slug: `2026-06-12-true-ha-control-plane-and-runtime-recovery`
- Started At: `2026-06-12T00:00:00Z`
- Worktree: `umbrella spec only; T1-T5 implementation tasks each create their own worktree via scripts/start-implementation.sh`
- Branch: `N/A; umbrella spec only`

## Objective And Non-Goals

- Objective:
  - 定义并分阶段落地 `ctf` 的真正高可用目标：`控制面强 HA + 执行面故障后自动重建`。
  - 让 API / gateway / PostgreSQL / Redis 任一单点故障都不再直接导致整站不可用。
  - 让 runtime node 故障后，练习实例与 AWD 运行态能被识别、标记并在健康 node 上重建。
  - 把当前“本地文件 + 进程内事件总线 + 默认 node 选择”的单实例遗留面收口成显式 owner 和可验证切片。
- Non-Goals:
  - 不实现容器 live migration、进程级无感迁移或 TCP 会话透明漂移。
  - 不把 PostgreSQL 改成多主写，也不把 Redis 一步切到 Cluster。
  - 不在本轮把全部后台任务统一迁成独立微服务；继续保留 modular monolith + 外部 runtime-agent 的控制面形态。
  - 不把题目执行面迁到 Kubernetes；保留当前 `runtime-agent` 协议和 node routing 作为执行 authority。

## Inputs

- Source docs:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/04-api-design.md`
  - `docs/operations/runtime-agent-deployment.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
- Related architecture/contracts:
  - `code/backend/internal/infrastructure/postgres/postgres.go`
  - `code/backend/internal/infrastructure/redis/redis.go`
  - `code/backend/internal/platform/events/bus.go`
  - `code/backend/internal/module/ops/infrastructure/contest_realtime_stream.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - `code/backend/internal/module/assessment/infrastructure/report_output_store.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_attachment_store.go`
- Related prior work:
  - `docs/plan/impl-plan/2026-06-02-runtime-control-plane-agent-split-plan.md`
  - `docs/plan/impl-plan/2026-06-07-contest-realtime-relay-externalization-implementation-plan.md`
  - `docs/plan/impl-plan/2026-06-08-multi-instance-distributed-lock-hardening-implementation-plan.md`
  - `docs/plan/impl-plan/2026-06-08-multi-instance-startup-recovery-gate-fix-implementation-plan.md`

## Task Classification

- Classification: `结构性改动 / 非琐碎任务`
- Why:
  - 同时触达配置接入、共享状态、后台任务 owner、事件分发、文件存储和 runtime failover。
  - 需要跨 `internal/app`、`auth`、`ops`、`practice`、`instance`、`container_runtime`、`assessment`、`challenge` 多个 owner 统一边界。
  - 如果不先收口方案，后续实现极易出现“多副本看似可跑，但下载、通知、重建不一致”的二次返工。

## Problem Statement

当前仓库已经具备“多 API 实例下部分后台任务不重复执行”的基础，但仍停留在半完成状态：

- 会话、WS ticket、proxy ticket、startup recovery lock、contest scheduler lock 已经进入 Redis，说明控制面不是纯单实例假设。
- `runtime-agent + node_id` 已经把执行 authority 从“当前 API 进程连哪台 Docker”收口到显式 node binding。
- 但 `platform/events.Bus` 仍是进程内内存实现，通知、practice progress cache invalidation 等异步 side effect 不是跨副本广播。
- 题目附件与报告导出仍是 LocalFS owner；API 多副本 behind LB 时，下载请求可能落到不持有文件的副本。
- Redis 接入仍是单 `redis.Client`，配置层没有 Sentinel / failover 抽象；当前 key 设计和 Lua/pipeline 模式也不适合直接切 Redis Cluster。
- `runtime_nodes` 已有数据模型，但默认 selector 只保证“找一个 schedulable node”；尚未把 node heartbeat、capacity、失效摘除、实例重建 owner 收口成当前事实。

如果继续只做“多起几个 API 副本”，系统会表现为：登录可用、部分后台任务不重复、但通知不稳、下载串实例、node 故障后实例长期失活。这不满足真正高可用。

## Target Architecture

### High-Level Goal

把系统收口到以下运行模型：

1. 控制面：`LB -> 多个 ctf-api`，全部副本共享 PostgreSQL、Redis、shared secret、shared file/object store，并通过 `/ready` 做摘流。
2. 状态面：PostgreSQL 继续单主写 + HA，Redis 采用 Sentinel failover，承担会话、锁、实时流、临时票据和 read-model cache。
3. 事件面：所有跨副本必须一致的异步 side effect 统一经过 DB outbox / Redis Stream，不再依赖进程内事件总线。
4. 执行面：多个 `runtime-agent` node 作为可调度执行宿主；node 故障后由控制面识别并重建期望运行态。

### Availability Definition

本计划中的“真正高可用”定义为：

- 任一 `ctf-api` 副本、单个 gateway 进程、单个 runtime-agent node 宕机，不应导致整站不可用。
- PostgreSQL primary、Redis primary 故障后，在 failover 窗口内允许短暂降级，但系统无需人工改配置即可恢复接流。
- 运行中的 WebSocket / SSH / HTTP proxy 会话在副本或 node 故障时允许中断，但对应用户后续请求必须能在健康副本上重新建立。
- Jeopardy / AWD 实例在 node 故障后允许重建，不承诺原容器会话与内存态持续存在。

## Slice Strategy

### Slice 1: 状态面 HA 接入改造

**目标**

- 把 PostgreSQL / Redis 接入层从“单地址直连”改成“支持 HA provider”的显式 owner。
- 保持现有上层业务模块不感知 Sentinel / failover 细节。

**Files:**
- Modify:
  - `code/backend/internal/config/config.go`
  - `code/backend/configs/config.yaml`
  - `code/backend/configs/config.prod.yaml`
  - `code/backend/internal/infrastructure/postgres/postgres.go`
  - `code/backend/internal/infrastructure/redis/redis.go`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/runtime-agent-deployment.md`
- Review:
  - `code/backend/internal/bootstrap/run.go`
  - `code/backend/internal/service/health/service.go`

**Plan**

- 为 PostgresConfig 增加更明确的 DSN / HA 接线说明，保证单主写 + failover 后重连仍由 driver / infra owner 处理。
- 为 RedisConfig 增加 `mode`、`master_name`、`sentinel_addrs` 等字段，允许 `single` 与 `sentinel` 两种模式。
- `infraredis.NewClient` 改为根据配置返回合适 client；当前不要引入 `ClusterClient`。
- 健康检查补充对“当前连接模式”和关键依赖 failover 结果的可观测字段，确保 `/ready` 能反映依赖恢复状态。

**Why this slice first**

- 没有 HA 状态面，其余 owner 收口都无法在生产拓扑成立。
- 这一步基本不改业务语义，风险可控，适合作为后续切片前置基线。

### Slice 2: 共享文件与共享密钥 owner 收口

**目标**

- 消除 API 多副本对本地磁盘的隐式依赖。
- 把“哪些文件必须共享、哪些密钥必须一致”写成显式 owner。

**Files:**
- Modify:
  - `code/backend/internal/config/config.go`
  - `code/backend/internal/module/assessment/infrastructure/report_output_store.go`
  - `code/backend/internal/module/assessment/application/commands/report_service.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_attachment_store.go`
  - `code/backend/internal/module/challenge/api/http/handler.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
- Create:
  - 如需要，新增 `shared storage` port / adapter，例如 `assessment/ports.ReportBinaryStore`、`challenge/ports.AttachmentBinaryStore`

**Plan**

- 先把 `LocalFS` 语义抽成显式 port，而不是让 application service 继续只知道本地路径。
- 报表输出与题目附件至少支持 `shared_fs` owner；如团队接受更稳路线，可直接设计为 `object storage` owner。
- AWD dynamic flag secret、SSH host key 在文档和配置上都要明确“多实例必须共用同源 secret / key material”。

**Acceptance**

- 任一 API 副本都能完成报告下载与附件下载。
- 重启任一 gateway / API 副本后，SSH host key fingerprint 不漂移。

### Slice 3: 跨副本事件总线与实时 side effect 收口

**目标**

- 让当前依赖 `inMemoryBus` 的关键 side effect 变成跨副本一致。
- 统一“领域事件”和“实时广播事件”的 owner。

**Files:**
- Modify:
  - `code/backend/internal/platform/events/bus.go`
  - `code/backend/internal/app/composition/root.go`
  - `code/backend/internal/module/ops/application/commands/notification_service.go`
  - `code/backend/internal/module/practice/application/queries/progress_timeline_service.go`
  - `code/backend/internal/module/challenge/application/challengecatalog/published_catalog_event.go`
  - `code/backend/internal/module/practice/application/commands/service_lifecycle.go`
  - `code/backend/internal/module/ops/runtime/module.go`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
- Create:
  - DB outbox / relay owner 所需的新 port、repository、dispatcher 文件
  - 如需要，`docs/reviews/backend/` 对事件 owner 的 review 记录

**Plan**

- 不建议把所有现有事件都一刀切搬到 Redis Pub/Sub；推荐做法是：
  - 关键领域 side effect：DB outbox + dispatcher
  - contest realtime / notification / cache invalidation：Redis Stream fanout
- 首批必须迁移的事件：
  - `practice.flag_accepted`
  - `challenge.publish_check_finished`
  - notification created/read
  - user progress cache invalidation
- `platform/events.Bus` 可以保留本地接口，但默认实现不再只是本地 map；要让跨副本 side effect 有唯一 owner。

**Acceptance**

- 用户连到任意 API 副本，都能收到自己触发的通知。
- practice progress cache 在任一副本提交 flag 后都会及时失效。

### Slice 4: runtime node health / capacity / failover

**目标**

- 把 `runtime_nodes` 从“可选 metadata + 默认 selector”提升为“可用性 owner”。
- 当 node 故障时，控制面能摘除、标记并重建期望运行态。

**Files:**
- Modify:
  - `code/backend/internal/module/container_runtime/entity/runtime_node.go`
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/instance_practice_runtime_node_selector_adapter.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
  - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
  - `code/backend/internal/module/instance/infrastructure/cleaner.go`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/runtime-agent-deployment.md`
- Create:
  - runtime node heartbeat updater / health evaluator job
  - runtime node selector tests / node failover runtime tests

**Plan**

- 为 `runtime_nodes` 明确维护这些事实：`health_status`、`last_seen_at`、`capacity_snapshot`、`schedulable`。
- selector 不再只按 `first schedulable node` 选；至少要支持：
  - 排除 unhealthy node
  - 优先使用显式绑定 node
  - 对新实例选健康 node
- 当 node 失联时：
  - 把该 node 上 active runtime 标记为 lost candidate
  - 让 cleanup / reconcile / desired-runtime owner 决定是否可重建
  - Jeopardy 与 AWD 的重建策略分别写清楚
- 当前不承诺透明会话迁移；SSH / WebSocket 断线后由用户重连。

**Acceptance**

- 停掉一个 runtime-agent node 后，新实例不会再调度到它。
- 该 node 上的期望 AWD service 最终会在健康 node 上补齐。

### Slice 5: awd-defense-ssh-gateway 多副本与接流规则

**目标**

- 让 SSH defense ingress 从“单进程可用”升级到“多 gateway 副本 + 稳定 host key + ticket 校验一致”。

**Files:**
- Modify:
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/runtime-agent-deployment.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`

**Plan**

- gateway 自身不持有唯一业务状态；真正关键的是：
  - ticket claims 来自共享 PostgreSQL / Redis
  - host key 稳定共享
  - 负载均衡器支持 TCP 层分发
- 若多 gateway 副本 behind LB，必须补“health / readiness / draining”文档与操作手册。

**Acceptance**

- 切掉单个 gateway 副本后，新的 SSH ticket 仍可用。
- 客户端只要信任同一 host key，不会因副本切换触发 host key mismatch。

## Files

- Create:
  - `docs/plan/impl-plan/2026-06-12-true-ha-control-plane-and-runtime-recovery-implementation-plan.md`
  - 后续实现阶段按 slice 需要新增 Redis Sentinel config、event relay、shared storage ports、node health job 等文件
- Modify:
  - `code/backend/internal/config/config.go`
  - `code/backend/internal/infrastructure/postgres/postgres.go`
  - `code/backend/internal/infrastructure/redis/redis.go`
  - `code/backend/internal/platform/events/bus.go`
  - `code/backend/internal/module/ops/runtime/module.go`
  - `code/backend/internal/module/ops/application/commands/notification_service.go`
  - `code/backend/internal/module/practice/application/queries/progress_timeline_service.go`
  - `code/backend/internal/module/assessment/infrastructure/report_output_store.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_attachment_store.go`
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/module/instance/infrastructure/cleaner.go`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/operations/runtime-agent-deployment.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`

## 复用与 Owner 决策

- Existing patterns searched:
  - Redis lease / keepalive：`instance cleaner`、`contest status updater`、`startup runtime recovery`
  - Redis Stream fanout：`contest realtime stream`
  - `runtime-agent + node_id` execution routing
  - LocalFS stores：report output / challenge attachment
- Reuse / extend / split / create-new decision:
  - 复用现有 Redis lock、Redis Stream、`runtime_nodes` 数据模型和 `runtimeNodeExecutionRouter`。
  - 扩展 `infraredis.NewClient` 支持 Sentinel，而不是新建平行 Redis owner。
  - 为共享文件与跨副本事件新增显式 port / adapter，不继续让业务层隐式依赖本地磁盘和本地内存 bus。
  - runtime failover 继续由 modular monolith 控制面 owner 承担，不拆出额外 control plane service。
- Owner boundary:
  - `internal/infrastructure/{postgres,redis}`：依赖接线与 HA client owner
  - `platform/events` + outbox/stream relay：跨副本异步 side effect owner
  - `assessment/challenge` storage adapters：共享文件 owner
  - `container_runtime` + `instance` + `practice`：runtime node health、调度和恢复 owner
  - `ops`：实时推送桥接 owner
- Why this is the narrowest safe surface:
  - 保留当前 modular monolith 与 runtime-agent 结构，不强行演进到微服务或 k8s operator。
  - 优先修 shared-state correctness，再做 runtime failover，能避免“功能看似分布式，结果边界不一致”的返工。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
  - `architect-agent`
- Why this pass fits:
  - 这不是单一模块改造，而是要先判断哪些单实例假设仍然存在，以及哪些已具备分布式 owner 基础。
- grill-with-docs findings:
  - 当前事实已经明确 `runtime-agent` 与 `node_id` 是执行 authority，但文档也明确没有把自动 failover 写成已落地能力。
  - 当前 Redis 用法更适合先走 Sentinel，不适合直接上 Cluster。
  - 共享文件与进程内事件总线是 API 多副本后最直接的 correctness blocker。
- Plan adjustments after challenge:
  - 不把所有后台任务迁移都塞进本计划；优先级收口为状态面、共享存储、事件总线、runtime failover。
  - 把 SSH gateway HA 单独作为后置 slice，而不是前置 blocker。

## Validation

- Commands:
  - 文档阶段：
    - `python3 scripts/check-docs-consistency.py`
  - 后续实现阶段最小命令基线：
    - `go test ./internal/infrastructure/redis -count=1`
    - `go test ./internal/platform/events -count=1`
    - `go test ./internal/module/ops/... -count=1`
    - `go test ./internal/module/practice/... -count=1`
    - `go test ./internal/module/instance/... -count=1`
    - `go test ./internal/app/composition -count=1`
- Manual checks:
  - 切 Redis primary / PostgreSQL primary 后，API 无需改配置即可恢复 `/ready`。
  - API 多副本 behind LB 时，通知、报告下载、附件下载不因落到不同副本而失效。
  - 停掉单个 runtime-agent node 后，新实例不会继续被调度到该 node，且期望 AWD service 最终补齐。
- Review focus:
  - Sentinel / failover 接入是否隐藏了与当前 key 设计冲突的 Redis 语义。
  - outbox/stream owner 是否覆盖了当前所有跨副本必须一致的 side effect。
  - shared storage port 是否真正消除了 LocalFS 假设，而不是只包一层路径字符串。
  - runtime node failover 是否把“可重建”和“不可透明迁移”的边界写清楚。

## Execution Recommendation

建议后续实现按从易到难、依赖优先的顺序拆成 5 个独立 task；每个 task 已单独展开为可执行 implementation plan：

1. T1 `redis-sentinel-and-postgres-ha-connectivity`（易，状态面 HA 前置基线）
   - Plan: `docs/plan/impl-plan/2026-06-12-redis-sentinel-and-postgres-ha-connectivity-implementation-plan.md`
2. T2 `shared-storage-owner-convergence`（较易，文件与密钥共源）
   - Plan: `docs/plan/impl-plan/2026-06-12-shared-storage-owner-convergence-implementation-plan.md`
3. T3 `ssh-gateway-ha-and-draining`（中等，依赖 T1 Redis HA 与 T2 host key 共源）
   - Plan: `docs/plan/impl-plan/2026-06-12-ssh-gateway-ha-and-draining-implementation-plan.md`
4. T4 `distributed-event-bus-and-outbox-relay`（较难，跨副本 side effect 与 outbox/stream owner）
   - Plan: `docs/plan/impl-plan/2026-06-12-distributed-event-bus-and-outbox-relay-implementation-plan.md`
5. T5 `runtime-node-health-and-failover-rebuild`（最难，runtime node health、调度排除与故障后重建）
   - Plan: `docs/plan/impl-plan/2026-06-12-runtime-node-health-and-failover-rebuild-implementation-plan.md`

每个 slice 单独走 startup gate、独立 review 和最小验证，不建议把整个 HA 路线压成一次长分支实现。T3 纯按难度可排在 T4 之前，但必须等 T2 的 host key / secret 共源 contract 成立后再实现。
