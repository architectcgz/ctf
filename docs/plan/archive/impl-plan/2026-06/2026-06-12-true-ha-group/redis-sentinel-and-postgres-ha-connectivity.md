<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# Redis Sentinel 与 PostgreSQL HA 接入 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` (recommended) or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking; flip each checkbox immediately after the expected result is reached.

**Goal:** 把 Redis / PostgreSQL 接入从单地址直连推进到可支撑 HA 拓扑的 infra owner，Redis 支持 `single` / `sentinel` 模式，PostgreSQL 连接显式保持 UTC 与 driver-level failover 语义，`/ready` 继续以 live Ping 反映依赖恢复状态。

**Architecture:** `internal/infrastructure/redis` 继续是 Redis client 构造唯一 owner；上层保持 `*redis.Client` 不变，因为 `go-redis/v9` 的 `NewFailoverClient` 也返回 `*redis.Client`。PostgreSQL 不在 service / health 层新增切换逻辑，只在 `PostgresConfig.DSN()` 和 infra open owner 内补齐连接契约；健康检查保持每次请求 live Ping，依赖恢复后自然重新 ready。

**Tech Stack:** Go, Viper config, GORM PostgreSQL driver, github.com/redis/go-redis/v9 v9.12.1, miniredis, code-workflow

---

## Task Metadata

- Task Slug: `2026-06-12-redis-sentinel-and-postgres-ha-connectivity`
- Started At: `2026-06-12T00:00:00Z`
- Worktree: `后续实现时运行 scripts/start-implementation.sh 2026-06-12-redis-sentinel-and-postgres-ha-connectivity 生成`
- Branch: `task/2026-06-12-redis-sentinel-and-postgres-ha-connectivity`

## Objective And Non-Goals

- Objective:
  - 为 `config.RedisConfig` 增加 `mode`、`master_name`、`sentinel_addrs`、`sentinel_username`、`sentinel_password` 等 Sentinel 接线字段。
  - 让 `infraredis.NewClient` 根据配置创建 single 或 sentinel client，并继续返回 `*redis.Client`，不扩散到业务模块。
  - 为 Redis mode 增加配置校验与单元测试，避免误配置成 Redis Cluster 或缺 Sentinel master。
  - 为 PostgreSQL DSN 显式追加 `TimeZone=UTC`，并保持 HA 切换由 driver / infra owner 承担。
  - 复用现有 `/ready` live Ping 语义，用测试确认 Redis/DB down 会返回 not ready。
- Non-Goals:
  - 不引入 Redis Cluster；现有 Lua / pipeline / 多 key 用法不具备 Cluster slot 约束设计。
  - 不把所有 Redis 依赖签名改成 interface 或 `UniversalClient`。
  - 不把 PostgreSQL 切换逻辑写入 handler、service 或 health 层。
  - 不搭建真实 Sentinel / PostgreSQL HA 集群作为本任务必需测试；真实演练进入运维验证。

## Inputs

- Source docs:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-control-plane-and-runtime-recovery-implementation-plan.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/runtime-agent-deployment.md`
- Related architecture/contracts:
  - `code/backend/internal/config/config.go`
  - `code/backend/internal/infrastructure/redis/redis.go`
  - `code/backend/internal/infrastructure/postgres/postgres.go`
  - `code/backend/internal/service/health/service.go`
  - `code/backend/internal/bootstrap/run.go`
  - `code/backend/configs/config.yaml`
  - `code/backend/configs/config.prod.yaml`
  - `code/backend/go.mod`
- Related prior work:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-08-multi-instance-distributed-lock-hardening-implementation-plan.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-08-multi-instance-startup-recovery-gate-fix-implementation-plan.md`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 触达全局配置、基础设施 client 构造和健康检查，是后续共享存储、事件总线、runtime failover 的前置 HA 基线。
  - 虽然业务行为不变，但配置契约一旦错误会导致生产启动或 failover 恢复失败。
  - Redis Cluster 必须明确排除，避免后续把 Sentinel 与 Cluster 语义混淆。

## Files

- Create:
  - `code/backend/internal/infrastructure/redis/redis_test.go`
- Modify:
  - `code/backend/internal/config/config.go`
  - `code/backend/internal/config/config_test.go`
  - `code/backend/internal/infrastructure/redis/redis.go`
  - `code/backend/internal/infrastructure/postgres/postgres.go`
  - `code/backend/configs/config.yaml`
  - `code/backend/configs/config.prod.yaml`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/runtime-agent-deployment.md`
- Review:
  - `code/backend/internal/service/health/service.go`
  - `code/backend/internal/bootstrap/run.go`
  - `code/backend/internal/app/http_server.go`
  - `code/backend/internal/app/router.go`
- Test:
  - `code/backend/internal/config/config_test.go`
  - `code/backend/internal/infrastructure/redis/redis_test.go`
  - `code/backend/internal/service/health/service_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - `infraredis.NewClient(ctx, cfg)` 当前只封装 `redis.NewClient` + startup Ping。
  - `postgres.Open(ctx, cfg)` 当前只封装 GORM open + pool config + startup Ping。
  - `/ready` 现有 `CheckReady` 每次 live Ping PostgreSQL / Redis，并已有 draining 状态。
  - `go.mod` 已使用 `github.com/redis/go-redis/v9 v9.12.1`，该版本提供 `NewFailoverClient(*FailoverOptions) *Client`。
- Reuse / extend / split / create-new decision:
  - 扩展 `RedisConfig` 和 `infraredis.NewClient`，不新增平行 Redis owner。
  - 保留 `*redis.Client` 签名，不改 app/composition 和各模块注入签名。
  - 为 Sentinel options 构造抽小函数做纯单测，避免为了单元测试启动真实 Sentinel。
  - PostgreSQL 只补 DSN UTC 和文档说明，不引入自定义 failover state machine。
- Owner boundary:
  - `internal/config`：配置字段、默认值、合法性校验 owner。
  - `internal/infrastructure/redis`：single / sentinel client 构造 owner。
  - `internal/infrastructure/postgres`：PostgreSQL open / pool / DSN owner。
  - `internal/service/health`：依赖 live Ping 与 readiness HTTP 状态 owner。
- Why this is the narrowest safe surface:
  - Sentinel 单 master 不要求改现有 Redis key、Lua、pipeline 设计；保留 `*redis.Client` 能把影响面限制在 config + infra。
  - `/ready` 已经具备依赖恢复后重新 ready 的语义，本任务不把它重写成缓存状态机。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
  - `dispatching-parallel-agents`
- Why this pass fits:
  - 该任务是 5 个 HA 切片中风险最低但最基础的一环，需要先确认 Redis client 类型、health 语义和 Postgres DSN owner 是否可最小扩展。
- grill-with-docs findings:
  - 项目规则要求 PostgreSQL 连接显式 UTC；当前 `PostgresConfig.DSN()` 尚未包含 `TimeZone=UTC`。
  - 现有 Redis 使用包含 Lua、pipeline、SetNX lock；这些在 Sentinel 单 master 下兼容，但不适合直接切 Redis Cluster。
  - `/ready` 每次 live Ping Redis / PostgreSQL，依赖恢复后无需额外状态即可恢复 ready。
- Plan adjustments after challenge:
  - 不引入 Redis Cluster 与 `UniversalClient`。
  - 不扩散 Redis interface；Sentinel client 仍返回 `*redis.Client`。
  - 真实 Sentinel failover 演练放入运维 manual checks，不作为单元测试前置。

## Ordered Task Slices

### Slice 1: Redis / Postgres config contract

- [ ] **Step 1: 写 Redis mode 配置校验失败测试**
  - 修改：`code/backend/internal/config/config_test.go`
  - 覆盖：未知 mode 拒绝、`single` 缺 `addr` 拒绝、`sentinel` 缺 `master_name` 拒绝、`sentinel` 缺 `sentinel_addrs` 拒绝。
  - 期望：新增测试先失败，报字段不存在或校验缺失。

- [ ] **Step 2: 写 PostgreSQL UTC 与 keyword/value escaping 测试**
  - 修改：`code/backend/internal/config/config_test.go`
  - 断言：`PostgresConfig{...}.DSN()` 包含 `TimeZone=UTC`。
  - 覆盖 password / database / username 含空格、单引号或反斜杠时，生成的 PostgreSQL keyword/value DSN 不会被截断或错误解析。
  - 期望：当前失败。

- [ ] **Step 3: 增加 RedisConfig 字段与默认值**
  - 修改：`code/backend/internal/config/config.go`
  - 字段：`Mode string`、`MasterName string`、`SentinelAddrs []string`、`SentinelUsername string`、`SentinelPassword string`。
  - 默认：`redis.mode = single`，保留 timeout 默认。

- [ ] **Step 4: 增加 Redis config validation**
  - 修改：`code/backend/internal/config/config.go`
  - 规则：`mode` 只能是 `single` 或 `sentinel`；`single` 要求 `addr`；`sentinel` 要求 `master_name` 和非空 `sentinel_addrs`。
  - 明确拒绝 `cluster`，错误信息包含 `redis.mode`。

- [ ] **Step 5: 更新 Postgres DSN UTC 与安全拼接**
  - 修改：`code/backend/internal/config/config.go`
  - 在 DSN 中追加 `TimeZone=UTC`。
  - 不继续裸 `fmt.Sprintf("host=%s ...")` 拼接特殊字符字段；新增 keyword/value DSN builder，对 host、user、password、dbname、sslmode、TimeZone 等 value 做 PostgreSQL 连接串需要的 quoting / escaping。

- [ ] **Step 6: 运行 config focused tests**
  - Run: `cd code/backend && go test ./internal/config -run 'Test(Config|Postgres|Redis)' -count=1`
  - Expected: 新增 config / DSN 测试通过。

### Slice 2: Redis client factory

- [ ] **Step 7: 写 redis options builder 单测**
  - Create: `code/backend/internal/infrastructure/redis/redis_test.go`
  - 覆盖：single mode 映射 `redis.Options`；sentinel mode 映射 `redis.FailoverOptions`；nil ctx 返回 `redis client requires context`。
  - 说明：Sentinel builder 纯单测，不启动真实 Sentinel。

- [ ] **Step 8: 重构 redis.NewClient 为 mode 分支**
  - Modify: `code/backend/internal/infrastructure/redis/redis.go`
  - `single` 使用 `redis.NewClient(&redis.Options{...})`。
  - `sentinel` 使用 `redis.NewFailoverClient(&redis.FailoverOptions{...})`。
  - 保持 startup Ping timeout 与错误包装。

- [ ] **Step 9: 用 miniredis 验证 single mode 构造**
  - Modify: `code/backend/internal/infrastructure/redis/redis_test.go`
  - 测试：`miniredis.RunT(t)` + `NewClient(ctx, RedisConfig{Mode:"single", Addr: mini.Addr()})` 能 Ping。

- [ ] **Step 10: 运行 redis focused tests**
  - Run: `cd code/backend && go test ./internal/infrastructure/redis -count=1`
  - Expected: redis 构造测试通过。

### Slice 3: readiness contract and docs

- [ ] **Step 11: 补 readiness 语义测试或确认现有覆盖**
  - Review: `code/backend/internal/service/health/service_test.go`
  - 如果现有 Redis down / DB down 覆盖不足，补一个 `CheckReady` dependency down 返回 503 的 focused test；不要把 health service 改成缓存状态机。

- [ ] **Step 12: 更新配置样例**
  - Modify: `code/backend/configs/config.yaml`
  - Modify: `code/backend/configs/config.prod.yaml`
  - 增加 `redis.mode: single`，并在 prod 中注释 Sentinel 示例字段：`master_name`、`sentinel_addrs`、`sentinel_username`、`sentinel_password`。

- [ ] **Step 13: 更新架构与运维说明**
  - Modify: `docs/architecture/backend/01-system-architecture.md`
  - Modify: `docs/architecture/backend/03-container-architecture.md`
  - Modify: `docs/operations/runtime-agent-deployment.md`
  - 说明：控制面支持 Redis Sentinel client；PostgreSQL 仍是单主写 HA；Redis Cluster 明确不在本阶段。

- [ ] **Step 14: 运行最小验证**
  - Run: `cd code/backend && go test ./internal/config ./internal/infrastructure/redis ./internal/infrastructure/postgres ./internal/service/health -count=1`
  - Expected: PASS。

- [ ] **Step 15: Commit**
  - Run: `git add code/backend/internal/config/config.go code/backend/internal/config/config_test.go code/backend/internal/infrastructure/redis/redis.go code/backend/internal/infrastructure/redis/redis_test.go code/backend/configs/config.yaml code/backend/configs/config.prod.yaml docs/architecture/backend/01-system-architecture.md docs/architecture/backend/03-container-architecture.md docs/operations/runtime-agent-deployment.md && git commit -m "feat(backend): 支持 Redis Sentinel 接入" -m "补齐 Redis single/sentinel 配置契约与 infra client 构造，保持上层 *redis.Client 注入不变。" -m "同步 PostgreSQL UTC DSN 与状态面 HA 运维说明，为后续多副本切片提供依赖基线。" -m "Task: 2026-06-12-redis-sentinel-and-postgres-ha-connectivity"`

## Validation

- Commands:
  - `cd code/backend && go test ./internal/config -run 'Test(Config|Postgres|Redis)' -count=1`
  - `cd code/backend && go test ./internal/infrastructure/redis -count=1`
  - `cd code/backend && go test ./internal/infrastructure/postgres -count=1`
  - `cd code/backend && go test ./internal/service/health -count=1`
  - `git diff --check -- code/backend/internal/config/config.go code/backend/internal/infrastructure/redis/redis.go code/backend/internal/infrastructure/postgres/postgres.go code/backend/configs/config.yaml code/backend/configs/config.prod.yaml docs/architecture/backend/01-system-architecture.md docs/architecture/backend/03-container-architecture.md docs/operations/runtime-agent-deployment.md`
- Manual checks:
  - 在 single Redis dev 环境启动 API，确认 `/ready` 为 200。
  - 在 staging Sentinel 环境配置 `redis.mode=sentinel`、`master_name`、`sentinel_addrs`，切换 Redis primary 后确认 API 无需改配置即可恢复 `/ready`。
  - 确认 PostgreSQL HA 代理 / VIP / DNS failover 后 driver 重新连接，不需要业务层改配置。
- Review focus:
  - Sentinel 接入是否仍返回 `*redis.Client`，没有把 Redis owner 扩散到业务模块。
  - 配置校验是否明确拒绝 Redis Cluster，避免与 Sentinel 混淆。
  - PostgreSQL UTC DSN 是否满足项目后端时间约束。
  - `/ready` 是否仍只表达依赖当前可用性，不承担额外 failover 状态机。
