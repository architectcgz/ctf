<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# AWD defense SSH gateway split Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` (recommended) or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 AWD defense SSH listener 从 `ctf-api` 进程里拆成独立 gateway 进程，让 `2222` 端口按单独服务部署，同时保留现有 proxy ticket / scope 校验协议。

**Architecture:** `ctf-api` 继续承担控制面，只签发 SSH ticket 并返回接入信息；新的 `awd-defense-ssh-gateway` 进程负责监听 `2222`、校验 ticket、解析作用域并进入目标防守工作区容器。gateway 仍复用现有 `container.defense_ssh_*` 配置和 runtime 组合能力，但不再作为 `InstanceModule` 的 background job 随 API 一起启动；多 node 场景下，交互式 exec 必须改为按 `container_id -> node_id` 路由。

**Tech Stack:** Go 1.26, `golang.org/x/crypto/ssh`, Zap, Gorm, Redis, existing `internal/app/composition` + `internal/bootstrap` wiring, Docker Compose dev.

---

## Task Metadata

- Task Slug: `2026-06-07-awd-defense-ssh-gateway-split`
- Started At: `2026-06-07T14:33:34Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-07-awd-defense-ssh-gateway-split`
- Branch: `task/2026-06-07-awd-defense-ssh-gateway-split`

## Objective And Non-Goals

- Objective:
  - 新增独立命令入口 `cmd/awd-defense-ssh-gateway`
  - 新增独立 bootstrap 入口，独立启动/停止 SSH gateway，并自行管理 Postgres、Redis、runtime lifecycle
  - 把 SSH gateway 的依赖装配从 `BuildInstanceModule` 中抽到单独 composition builder
  - API 保留 `IssueAWDDefenseSSHTicket()` 和现有返回结构，不再内嵌启动 SSH listener
  - gateway 在多 node / remote-agent 场景下按容器所属 node 做交互 exec 路由
  - 本地 compose dev 与运维文档同步切到独立 gateway service
- Non-Goals:
  - 不改 SSH 登录协议、用户名格式、ticket payload 或 workbench scope 语义
  - 不把 gateway 合并进 `runtime-agent`
  - 不在本次引入新的配置命名空间；继续复用 `container.defense_ssh_*`
  - 不顺手重写整个 bootstrap 框架、runtime module 或 deployment system
  - 不承诺故障迁移、自动 HA 或多个 gateway 实例共享同一 host key 的编排能力

## Inputs

- Source docs:
  - `README.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/design/backend-module-boundary-target.md`
  - `docs/operations/runtime-agent-deployment.md`
- Related architecture/contracts:
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/internal/app/composition/runtime_module.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/bootstrap/run.go`
  - `code/backend/internal/bootstrap/runtime_agent.go`
  - `code/backend/internal/config/config.go`
  - `code/backend/internal/module/instance/application/queries/proxy_ticket_service.go`
  - `code/backend/internal/module/runtime/infrastructure/awd_target_proxy_repository.go`
  - `code/backend/Dockerfile`
  - `code/backend/scripts/docker-entrypoint.sh`
  - `docker/docker-compose.dev.yml`
- Related prior work:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-02-runtime-control-plane-agent-split-plan.md`
  - `docs/operations/runtime-agent-deployment.md`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 同时触达 composition owner、bootstrap 入口、容器镜像打包、compose dev 和架构/运维文档
  - 涉及进程边界和部署边界变化，不是局部可逆小修
  - 需要在多 node 场景下确认 SSH exec 的 authority owner，不能只做“把监听端口搬出去”的表面拆分

## Files

- Create:
  - `code/backend/cmd/awd-defense-ssh-gateway/main.go`
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go`
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway_test.go`
- Modify:
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/internal/app/composition/runtime_module.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
  - `code/backend/internal/app/router_composition_structure_test.go`
  - `code/backend/internal/bootstrap/run.go`
  - `code/backend/internal/bootstrap/run_test.go`
  - `code/backend/internal/config/config_test.go`
  - `code/backend/Dockerfile`
  - `code/backend/scripts/docker-entrypoint.sh`
  - `docker/docker-compose.dev.yml`
  - `scripts/lib/check-consistency/architecture.sh`
  - `README.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/operations/runtime-agent-deployment.md`
- Review:
  - `docs/architecture/backend/01-system-architecture.md` 是否需要补一行 service 拆分后的当前事实
  - `docs/design/backend-module-boundary-target.md` 是否仍与新 owner 边界一致
- Test:
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - `code/backend/internal/app/router_composition_structure_test.go`
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway_test.go`
  - `code/backend/internal/bootstrap/run_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - `cmd/api` + `internal/bootstrap/run.go`
  - `cmd/runtime-agent` + `internal/bootstrap/runtime_agent.go`
  - `BuildInstanceModule` / `BuildContainerRuntimeModule`
  - `runtimeNodeExecutionRouter` 现有 `container_id -> node_id` 解析与路由能力
- Reuse / extend / split / create-new decision:
  - 复用现有 proxy ticket、scope repository、AWD defense SSH 协议和 `AWDDefenseSSHGateway` 本体
  - 新增独立 `cmd` 与 bootstrap，而不是把 listener 继续挂在 API background job
  - 扩展 `runtimeNodeExecutionRouter` 让交互式 exec 也能按 node 路由，而不是为 gateway 再造一套 runtime client 解析逻辑
  - 复用现有 `container.defense_ssh_*` 配置，不在本次再拆 `gateway.*`
- Owner boundary:
  - `ctf-api`: SSH ticket 签发、返回 host/port/username/password/command、控制面权限判断
  - `awd-defense-ssh-gateway`: TCP/SSH ingress、ticket 校验、session 建立、按 node 进入目标工作区容器
  - `runtime-agent`: 纯执行面，不感知 contest/team/service/ticket 语义
- Why this is the narrowest safe surface:
  - 端口监听与 API 横向扩缩容冲突，必须拆进独立进程
  - ticket/scope 语义已经稳定，保持在 API + gateway 间分工最小
  - 多 node 场景下不补交互 exec 路由，拆完仍会把 SSH 流量打到默认 node，属于同一 touched surface 上的 correctness 缺口，必须一起收口

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - 这是边界拆分任务，核心风险在 owner、部署与复用点，不是单点 bug 修复
- grill-with-docs findings:
  - 现有文档已经把 `runtime-agent` 定义成 execution plane，因此 gateway 不应并入 `runtime-agent`
  - `README.md` 和当前 compose dev 仍把 `2222` 视为 `ctf-api` 附属端口，实施时必须同步更新运行入口和文档
  - `runtimeNodeExecutionRouter` 已经具备 `container_id -> node_id` 解析能力，gateway 多 node 支持应复用这条链路，而不是新造 lookup
  - 当前不需要单独 ADR：控制面/执行面分离已在现有架构文档里说明，本次只是把 AWD SSH ingress 对齐到既有边界
- Plan adjustments after challenge:
  - 把 “独立 cmd” 扩成 “独立 cmd + 独立镜像入口/compose service”
  - 把 “拆监听端口” 扩成 “同时补齐 interactive exec node routing”
  - 保持配置复用，但明确记录这是本次的刻意非目标

## Validation

- Commands:
  - `cd code/backend && go test ./internal/app/composition -run 'TestAWDDefenseSSHGateway|TestRuntimeNodeExecutionRouter' -count=1`
  - `cd code/backend && go test ./internal/app -run 'TestBuildInstanceModule|TestBuildAWDDefenseSSHGateway' -count=1`
  - `cd code/backend && go test ./internal/bootstrap -run 'TestRunAWDDefenseSSHGateway|TestShutdownGracefully' -count=1`
  - `cd code/backend && go test ./internal/config -run 'TestValidate.*DefenseSSH' -count=1`
  - `bash scripts/check-workflow-governance.sh`
- Manual checks:
  - `docker compose -f docker/docker-compose.dev.yml up -d --build ctf-api ctf-awd-defense-ssh-gateway ctf-postgres ctf-redis`
  - 通过 API 申请 AWD defense SSH ticket，确认 `host/port` 仍指向 `container.defense_ssh_*`
  - 用返回的 `ssh user+contest+service@host -p 2222` 登录，确认能进入 `/workspace`
  - 在 `runtime_agent.enabled=true` 的单 remote-agent 场景重复一次，确认 SSH 会落到目标 node
- Review focus:
  - API 是否彻底停止注册/启动 SSH listener
  - gateway 是否只持有 ingress 所需依赖，没有回流 HTTP handler / contest owner 逻辑
  - interactive exec 是否按 `container_id -> node_id` 路由，而不是回退到默认 node
  - Dockerfile / entrypoint / compose 是否允许 API 与 gateway 共用镜像但独立运行
  - 多 API 副本下，负载均衡只依赖进程级 `/ready` 摘流，不把 startup recovery / practice scheduler 的分布式锁 leader 状态误当成 HTTP readiness

## Working Design

- 依据：
  - `instance_module.go` 当前通过 `root.RegisterBackgroundJob("awd_defense_ssh_gateway", ...)` 把 listener 绑在 API 生命周期里
  - `awd_defense_ssh_gateway.go` 的 `Start()` 直接 `net.Listen(":%d")`，天然是单端口单 owner
  - `runtime_agent.go` 当前只承担 host executor + checker runner 的 gRPC 执行面
  - `runtime_node_execution_router.go` 已具备 `FindRuntimeNodeIDByContainerID()` 路由基础，但还没暴露 interactive exec
- 目标：
  - API 副本数不再决定 `2222` 的部署形态
  - SSH 接入保留现有用户体验与权限语义
  - 多 node 场景下 SSH 进入正确的工作区容器所在 node
- 边界：
  - 只调整进程/部署 owner 和 runtime exec 路由
  - 不改 ticket contract、不改前端调用、不改 defense workbench HTTP API
- 风险：
  - 如果 Docker 镜像入口处理不干净，gateway 容器可能误跑 migration 或误启动 API
  - 如果只拆 listener 不补 node routing，多 node 环境会把 SSH 打到错误 node
  - 继续复用 `container.defense_ssh_*` 意味着 API-only 部署仍需携带这组配置；这是已知取舍
- 验证：
  - `TDD`
  - 先写结构/路由/bootstrap 失败测试，再做最小实现，最后用 compose dev 做一次真实 SSH 联调

### Addendum: 多实例 API health / readiness 缺口

- 背景：
  - `ctf-api` 拆出 `2222` 后可以横向部署多个 HTTP API 副本。
  - 后台全局编排已经通过 Redis owner lock 收敛，但 HTTP 健康检查仍只有 `/health`，且 `/health` 同时承担诊断和流量接入判断，容易在多实例部署里把依赖诊断、存活探测和摘流语义混在一起。
- 决策：
  - `/live`：只表示进程仍能响应，不检查 Postgres / Redis。
  - `/ready`：表示当前 API 副本可接收流量，检查 Postgres、Redis 和本进程是否进入 shutdown drain。
  - `/health`：继续作为依赖诊断聚合，保留原有 `ok / degraded` 语义。
  - readiness 不读取 startup recovery / practice scheduler lock owner；standby API 副本只要依赖可用且未 drain，就应当可接 HTTP 流量。
- 代码 owner：
  - `internal/service/health`：health / live / ready 的状态构造与依赖检查。
  - `internal/handler/health`：HTTP handler。
  - `internal/app/router.go`：根路径与 `/api/v1` 路由接线。
  - `internal/app/http_server.go`：shutdown 开始时先把 readiness 标为 draining，再停止后台任务。
  - `internal/module/instance/application/commands/startup_runtime_recovery_service.go`：standby 副本未拿到 startup recovery lock 时，`Start()` 只启动后台选举循环并立即返回；初始拿到锁的 leader 仍同步完成恢复初始化后再返回。
- 运行侧：
  - `docker/docker-compose.dev.yml` 的 `ctf-api` healthcheck 改用 `/ready`。
  - 运维手册中把 `/ready` 作为接流量检查，`/health` 作为依赖诊断保留。

### Task 1: 锁定结构边界与多 node 路由测试

**Files:**
- Modify: `code/backend/internal/app/router_composition_structure_test.go`
- Modify: `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
- Modify: `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`

- [ ] **Step 1: 写失败的结构测试，禁止 `InstanceModule` 继续注册 SSH gateway**

  在 `TestBuildInstanceModuleDelegatesToSubBuilders` 附近新增 guard，明确 `instance_module.go` 不再包含：

  ```go
  blocked := []string{
      `root.RegisterBackgroundJob(`,
      `"awd_defense_ssh_gateway"`,
      `NewAWDDefenseSSHGateway(`,
  }
  ```

- [ ] **Step 2: 写失败的路由测试，要求 interactive exec 走 node router**

  在 `runtime_node_execution_router_test.go` 新增用例，构造 `containerID -> nodeID` 映射并断言：

  ```go
  err := router.ExecContainerInteractive(ctx, "workspace-ctr", []string{"sh"}, stdin, stdout)
  ```

  只能调用目标 node 对应的 executor，不能落到默认 node。

- [ ] **Step 3: 写失败的 gateway 组合测试，要求 builder 在有 node router 时使用它**

  新增测试覆盖：
  - `runtime.nodeRouter != nil` 时，gateway 的 executor 应来自 router
  - `runtime.nodeRouter == nil` 时，仍回退到 `module.InteractiveExecutor`

- [ ] **Step 4: 运行失败测试确认 guard 生效**

  Run:

  ```bash
  cd code/backend
  go test ./internal/app -run 'TestBuildInstanceModule' -count=1
  go test ./internal/app/composition -run 'TestRuntimeNodeExecutionRouter|TestAWDDefenseSSHGateway' -count=1
  ```

  Expected: 至少一个 FAIL，且失败点指向旧的 API 注册或缺失的 interactive exec 路由。

- [ ] **Step 5: Commit**

  ```bash
  git add code/backend/internal/app/router_composition_structure_test.go code/backend/internal/app/composition/runtime_node_execution_router_test.go code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go
  git commit -m "test(backend): 锁定 ssh gateway 拆分边界"
  ```

### Task 2: 抽出独立 gateway builder，并从 API 生命周期移除 listener

**Files:**
- Modify: `code/backend/internal/app/composition/instance_module.go`
- Modify: `code/backend/internal/app/composition/runtime_module.go`
- Modify: `code/backend/internal/app/composition/runtime_node_execution_router.go`
- Modify: `code/backend/internal/app/composition/runtime_node_execution_router_test.go`

- [ ] **Step 1: 新增 composition-level builder，专门装配 SSH gateway**

  在 `composition` 包增加 builder，目标签名：

  ```go
  func BuildAWDDefenseSSHGateway(root *Root, runtime *ContainerRuntimeModule) *AWDDefenseSSHGateway
  ```

  该 builder 负责复用 `proxyTicketService`、scope reader、logger 和 `container.defense_ssh_*` 配置，但不注册 background job。

- [ ] **Step 2: 扩展 `runtimeNodeExecutionRouter` 支持 interactive exec**

  增加 `nodeRuntimeClient` / `runtimeNodeExecutionRouter` 的 `ExecContainerInteractive(...)` 实现，内部复用：
  - `clientForContainerID()`
  - `resolveNodeIDForContainer()`

  不新增第二套 node lookup。

- [ ] **Step 3: 让 gateway builder 在存在 router 时优先使用 router**

  逻辑目标：
  - 默认 executor = `runtime.runtime.InteractiveExecutor`
  - 若 `runtime.nodeRouter != nil`，且其已实现 `ExecContainerInteractive`，则改用 router

- [ ] **Step 4: 从 `BuildInstanceModule` 删除 SSH gateway 注册**

  `BuildInstanceModule` 只保留：
  - `startup_runtime_recovery`
  - `runtime_cleaner`
  - `instance_stopping_cleanup`

  不再出现 `awd_defense_ssh_gateway` background job 注册。

- [ ] **Step 5: 跑 Task 1 的测试，直到全部转绿**

  Run:

  ```bash
  cd code/backend
  go test ./internal/app -run 'TestBuildInstanceModule' -count=1
  go test ./internal/app/composition -run 'TestRuntimeNodeExecutionRouter|TestAWDDefenseSSHGateway' -count=1
  ```

  Expected: PASS。

- [ ] **Step 6: Commit**

  ```bash
  git add code/backend/internal/app/composition/instance_module.go code/backend/internal/app/composition/runtime_module.go code/backend/internal/app/composition/runtime_node_execution_router.go code/backend/internal/app/composition/runtime_node_execution_router_test.go code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go code/backend/internal/app/router_composition_structure_test.go
  git commit -m "refactor(backend): 抽离 awd defense ssh gateway 装配"
  ```

### Task 3: 新增独立 cmd / bootstrap / 镜像入口

**Files:**
- Create: `code/backend/cmd/awd-defense-ssh-gateway/main.go`
- Create: `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go`
- Create: `code/backend/internal/bootstrap/awd_defense_ssh_gateway_test.go`
- Modify: `code/backend/internal/bootstrap/run.go`
- Modify: `code/backend/internal/bootstrap/run_test.go`
- Modify: `code/backend/Dockerfile`
- Modify: `code/backend/scripts/docker-entrypoint.sh`

- [ ] **Step 1: 为 gateway 进程写失败的 bootstrap 测试**

  覆盖最小行为：
  - 能独立加载 config / logger / db / redis / runtime module
  - 只启动 gateway 自己，不启动 HTTP server
  - shutdown 时会关闭 listener、runtime lifecycle、db、redis

- [ ] **Step 2: 实现新的命令入口**

  新增：

  ```go
  // cmd/awd-defense-ssh-gateway/main.go
  func main() {
      bootstrap.RunAWDDefenseSSHGateway()
  }
  ```

  `RunAWDDefenseSSHGateway()` 复用 `mustOpenPostgres`、`mustOpenRedis`、`closeResources` 这类共享 helper，不重复抄初始化逻辑。

- [ ] **Step 3: 让镜像同时产出 API 与 gateway 二进制**

  `Dockerfile` 目标：
  - 保留 `/app/ctf-api`
  - 新增 `/app/ctf-awd-defense-ssh-gateway`

  `docker-entrypoint.sh` 目标：
  - 默认无参数时仍启动 API
  - 仅当最终执行目标是 `/app/ctf-api` 时跑 migration
  - gateway service 通过显式 command 启动时跳过 migration

- [ ] **Step 4: 运行 bootstrap 测试与现有关机测试**

  Run:

  ```bash
  cd code/backend
  go test ./internal/bootstrap -run 'TestRunAWDDefenseSSHGateway|TestShutdownGracefully' -count=1
  ```

  Expected: PASS。

- [ ] **Step 5: Commit**

  ```bash
  git add code/backend/cmd/awd-defense-ssh-gateway/main.go code/backend/internal/bootstrap/awd_defense_ssh_gateway.go code/backend/internal/bootstrap/awd_defense_ssh_gateway_test.go code/backend/internal/bootstrap/run.go code/backend/internal/bootstrap/run_test.go code/backend/Dockerfile code/backend/scripts/docker-entrypoint.sh
  git commit -m "feat(backend): 新增独立 awd defense ssh gateway 进程"
  ```

### Task 4: 更新 dev compose、文档和守卫

**Files:**
- Modify: `docker/docker-compose.dev.yml`
- Modify: `scripts/lib/check-consistency/architecture.sh`
- Modify: `README.md`
- Modify: `docs/architecture/backend/03-container-architecture.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Modify: `docs/operations/runtime-agent-deployment.md`
- Review: `docs/architecture/backend/01-system-architecture.md`

- [ ] **Step 1: 把 dev compose 改成独立 gateway service**

  目标结构：
  - `ctf-api` 只暴露 `8080`
  - 新增 `ctf-awd-defense-ssh-gateway` service 暴露 `2222`
  - gateway 复用 backend image、storage volume、docker sock、postgres/redis 依赖
  - gateway 使用显式 command 启动 `/app/ctf-awd-defense-ssh-gateway`

- [ ] **Step 2: 更新架构一致性守卫**

  在 `scripts/lib/check-consistency/architecture.sh` 里保留 “启用 defense SSH 时必须暴露对应端口” 的检查，并追加一条：
  - `docker-compose.dev.yml` 必须存在 `ctf-awd-defense-ssh-gateway` service

- [ ] **Step 3: 更新 README 与运行文档**

  至少改清：
  - `2222` 现在属于独立 gateway service，不再属于 `ctf-api`
  - 本地 compose dev 的启动命令要包含 `ctf-awd-defense-ssh-gateway`
  - `runtime-agent` 仍不是 SSH gateway owner

- [ ] **Step 4: 更新当前架构事实文档**

  至少改清：
  - `03-container-architecture.md`：runtime HTTP facade / SSH ingress 的 owner
  - `07-modular-monolith-refactor.md`：`InstanceModule` 不再注册 gateway
  - 如 `01-system-architecture.md` 已明确写错 owner，再补最小事实修正

- [ ] **Step 5: 跑最小充分验证**

  Run:

  ```bash
  cd code/backend
  go test ./internal/app/composition -run 'TestAWDDefenseSSHGateway|TestRuntimeNodeExecutionRouter' -count=1
  go test ./internal/app -run 'TestBuildInstanceModule' -count=1
  go test ./internal/bootstrap -run 'TestRunAWDDefenseSSHGateway|TestShutdownGracefully' -count=1
  go test ./internal/config -run 'TestValidate.*DefenseSSH' -count=1
  cd /home/azhi/workspace/projects/.worktrees/ctf/2026-06-07-awd-defense-ssh-gateway-split
  bash scripts/check-workflow-governance.sh
  ```

  Expected: 全部 PASS。

- [ ] **Step 6: 做一次 compose dev 手工联调**

  Run:

  ```bash
  docker compose -f docker/docker-compose.dev.yml up -d --build ctf-api ctf-awd-defense-ssh-gateway ctf-postgres ctf-redis
  ```

  然后验证：
  - API `GET /health` 正常
  - `2222` 由 `ctf-awd-defense-ssh-gateway` 监听
  - 申请 AWD defense SSH ticket 后能通过 SSH 登录到 `/workspace`

- [ ] **Step 7: Commit**

  ```bash
  git add docker/docker-compose.dev.yml scripts/lib/check-consistency/architecture.sh README.md docs/architecture/backend/07-modular-monolith-refactor.md docs/operations/runtime-agent-deployment.md
  git commit -m "docs(ops): 对齐 awd defense ssh gateway 独立部署"
  ```
