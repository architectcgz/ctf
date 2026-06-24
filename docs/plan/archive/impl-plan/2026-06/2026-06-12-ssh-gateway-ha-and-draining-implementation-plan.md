<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# AWD Defense SSH Gateway HA 与 Draining Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` (recommended) or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking; flip each checkbox immediately after the expected result is reached.

**Goal:** 让 AWD defense SSH ingress 从单 gateway 进程可用推进到多 gateway 副本 behind TCP LB：共享 ticket 校验、稳定 host key、可观测 readiness，并支持停止接新连接的 draining 语义。

**Architecture:** Gateway 自身不持有唯一业务状态；ticket claims 继续由 Redis / PostgreSQL 提供共享事实，host key 共源由 T2 提供前置 contract。T3 在 gateway 生命周期内新增 readiness/draining 状态：对外 readiness 直接体现在 TCP listener 是否还接收新连接，LB 通过 `container.defense_ssh_port` 的 TCP health check 摘流；`Drain` 用于停止接新 SSH 连接但不立即杀掉已有 session；`Stop` 保留 hard-stop 语义用于最终关闭。

**Tech Stack:** Go, x/crypto/ssh, Redis-backed proxy tickets, PostgreSQL AWD scope repository, TCP listener lifecycle, code-workflow

---

## Task Metadata

- Task Slug: `2026-06-12-ssh-gateway-ha-and-draining`
- Started At: `2026-06-12T00:00:00Z`
- Worktree: `后续实现时运行 scripts/start-implementation.sh 2026-06-12-ssh-gateway-ha-and-draining 生成`
- Branch: `task/2026-06-12-ssh-gateway-ha-and-draining`

## Objective And Non-Goals

- Objective:
  - 明确 `container.defense_ssh_host` / `defense_ssh_port` 是 LB 对外地址，不是 gateway 进程 bind host。
  - 增加 gateway readiness / draining 状态，支持 TCP LB 在 shutdown 前摘除副本。
  - 证明一个副本签发的 defense SSH ticket 可由另一个 gateway 副本校验。
  - 保持 host key 稳定共享，避免 TCP LB 切换后客户端看到 host key mismatch。
  - 文档化多 gateway 副本的接流、摘流和 shutdown 操作步骤。
- Non-Goals:
  - 不实现已有 SSH session 的透明迁移；gateway 或 runtime node 故障时 session 可以中断，用户后续重连。
  - 不把 SSH gateway 改成独立微服务；继续沿用当前 bootstrap / composition 形态。
  - 不在 T3 重新设计 host key 存储；依赖 T2 的 shared source contract。
  - 不改 proxy ticket 的核心 claims 结构，除非测试发现多副本校验缺字段。

## Inputs

- Source docs:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-control-plane-and-runtime-recovery-implementation-plan.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/shared-storage-owner-convergence.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/runtime-agent-deployment.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
- Related architecture/contracts:
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway_builder.go`
  - `code/backend/internal/app/composition/runtime_http_service_adapter.go`
  - `code/backend/internal/module/instance/infrastructure/proxy_ticket_store.go`
  - `code/backend/internal/module/instance/application/queries/proxy_ticket_service.go`
  - `code/backend/internal/module/contest/infrastructure/awd_proxy_scope_repository.go`
  - `code/backend/internal/config/config.go`
- Related prior work:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-07-awd-defense-ssh-gateway-split-implementation-plan.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/shared-storage-owner-convergence.md`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 触达 SSH gateway 生命周期、ticket 校验链路、bootstrap shutdown 顺序、配置语义和运维接流规则。
  - 需要明确 LB 摘流与 session 中断边界，不能把“多副本”写成无感迁移。
  - T3 依赖 T1/T2 的 Redis HA 与 host key 共源；缺少前置条件时无法成立。

## Files

- Create:
  - 如有需要，新增 `code/backend/internal/app/composition/awd_defense_ssh_gateway_readiness.go`
- Modify:
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go`
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway_test.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway_builder.go`
  - `code/backend/internal/app/composition/runtime_http_service_adapter.go`
  - `code/backend/internal/module/instance/infrastructure/proxy_ticket_store.go`
  - `code/backend/internal/module/instance/infrastructure/proxy_ticket_store_test.go`
  - `code/backend/internal/module/instance/application/proxy_ticket_service_test.go`
  - `code/backend/internal/config/config.go`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/runtime-agent-deployment.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
- Review:
  - `code/backend/internal/module/contest/infrastructure/awd_proxy_scope_repository.go`
  - `code/backend/internal/app/composition/instance_module.go`
- Test:
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway_test.go`
  - `code/backend/internal/module/instance/infrastructure/proxy_ticket_store_test.go`
  - `code/backend/internal/module/instance/application/proxy_ticket_service_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - Gateway `Start` 当前 `net.Listen("tcp", fmt.Sprintf(":%d", port))`，host 只用于对外返回连接信息。
  - Gateway `Stop` 当前 cancel context、close listener、close active conns，是 hard stop。
  - Proxy ticket claims 已写 Redis，校验时再从 PostgreSQL 查 AWD defense scope，具备跨副本校验基础。
  - Host key 当前来自本地 path，不存在时生成；T3 依赖 T2 把该 path/source 变成多副本共源。
- Reuse / extend / split / create-new decision:
  - 复用现有 Redis proxy ticket store 与 Postgres scope reader，不新增 gateway 专属 ticket store。
  - 新增 gateway readiness/draining 状态而不是把 HTTP health service 直接塞进 SSH gateway。
  - 保留 `Stop` hard-stop；新增 `Drain(ctx)` 或等价状态用于 LB 摘流。
  - 文档中明确 `DefenseSSHHost` 是 LB/external host，不改变监听语义。
- Owner boundary:
  - `app/composition/AWDDefenseSSHGateway`：SSH listener、authentication、readiness/draining lifecycle owner。
  - `instance/proxy_ticket_service` + `proxy_ticket_store`：ticket 签发与 Redis claims owner。
  - `contest/awd_proxy_scope_repository`：AWD defense target scope 当前事实 owner。
  - T2 shared storage：host key 共源 owner。
- Why this is the narrowest safe surface:
  - 多副本成立的关键是 shared ticket + stable host key + LB 摘流；无需重写 SSH protocol 或 runtime interactive executor。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
  - `dispatching-parallel-agents`
- Why this pass fits:
  - SSH gateway HA 涉及连接生命周期和用户可见 trust fingerprint，必须先分清可承诺的恢复范围。
- grill-with-docs findings:
  - Availability definition 明确允许 SSH 会话在副本或 node 故障时中断；只要求后续请求可在健康副本重建。
  - T2 是 T3 前置依赖：没有稳定 host key，共享 ticket 也无法避免客户端 host key mismatch。
  - `DefenseSSHHost` 不参与 `net.Listen`，当前实际是对外地址配置。
- Plan adjustments after challenge:
  - 不把 draining 伪装成会话迁移；只做到停止接新连接与有界等待。
  - 不额外引入 HTTP `/ready`；gateway 的可观测摘流 owner 直接是 TCP listener / LB health check。
  - 不在 gateway 层复制 ticket 数据；继续以 Redis/PostgreSQL 为共享事实。

## Ordered Task Slices

### Slice 1: cross-replica ticket and host-key contract tests

- [x] **Step 1: 写 proxy ticket store round-trip 测试**
  - Create: `code/backend/internal/module/instance/infrastructure/proxy_ticket_store_test.go`
  - 覆盖 Save/Find、TTL 过期、missing 返回 nil、claims JSON 字段完整。

- [x] **Step 2: 写跨副本 ticket 校验测试**
  - Modify: `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
  - 构造两个 gateway / auth service，共用同一 Redis store 和 scope reader；副本 A 签发 ticket，副本 B authenticate 成功。

- [x] **Step 3: 写 shared host key fingerprint 测试**
  - Modify: `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
  - 两个 gateway 使用同一 T2 shared host key source/path，fingerprint 一致；非法 key fail fast。

- [x] **Step 4: 运行 ticket/host-key focused tests**
  - Run: `cd code/backend && go test ./internal/module/instance/infrastructure ./internal/app/composition -run 'ProxyTicket|AWDDefenseSSH.*(Ticket|HostKey|Authenticate)' -count=1`

### Slice 2: readiness and draining lifecycle

- [x] **Step 5: 定义 gateway readiness/draining 状态接口**
  - Modify/Create: `code/backend/internal/app/composition/awd_defense_ssh_gateway.go` 或 `awd_defense_ssh_gateway_readiness.go`
  - 状态建议：not_started、ready、draining、stopped。
  - `Ready()` 在 listener active 且非 draining 时 true。

- [x] **Step 6: 写 readiness 状态测试**
  - Modify: `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
  - 覆盖 Start 前 not ready、Start 后 ready、Drain 后 not ready、Stop 后 stopped。

- [x] **Step 7: 实现 Drain(ctx) 语义**
  - Modify: `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
  - Drain 后停止接新连接（关闭 listener 或让 accept loop 拒绝），但不主动关闭已有 active conns；deadline 到期后由 Stop hard close。

- [x] **Step 8: 调整 bootstrap shutdown 顺序**
  - Modify: `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go`
  - SIGTERM 后先标记 draining / 摘流等待，再 hard Stop；避免一进入 shutdown 就取消 active executor context。

- [x] **Step 9: 运行 lifecycle tests**
  - Run: `cd code/backend && go test ./internal/app/composition ./internal/bootstrap -run 'AWDDefenseSSH.*(Ready|Drain|Stop|Shutdown)' -count=1`

### Slice 3: LB-facing config and docs

- [x] **Step 10: 明确 config 命名/注释**
  - Modify: `code/backend/internal/config/config.go`
  - 注释或校验中说明 `container.defense_ssh_host` 是客户端访问/LB host；监听继续是 `:port` 或显式新增 bind host 字段（如需）。

- [x] **Step 11: 检查 HTTP access response**
  - Review/Modify: `code/backend/internal/app/composition/runtime_http_service_adapter.go`
  - 确认返回给客户端的是 LB host/port，不是 pod bind address。

- [x] **Step 12: 更新运维文档**
  - Modify: `docs/architecture/backend/03-container-architecture.md`
  - Modify: `docs/operations/runtime-agent-deployment.md`
  - Modify: `docs/operations/awd-host-reboot-recovery-drill.md`
  - 写清：TCP LB health/readiness、drain before terminate、host key mount、ticket Redis/Postgres 依赖、session 可中断边界。

- [x] **Step 13: 运行最小验证**
  - Run: `cd code/backend && go test ./internal/app/composition ./internal/bootstrap ./internal/module/instance/... -run 'AWDDefenseSSH|ProxyTicket|DefenseSSH' -count=1`

- [x] **Step 14: Commit**
  - Run: `git add code/backend/internal/bootstrap/awd_defense_ssh_gateway.go code/backend/internal/bootstrap/awd_defense_ssh_gateway_test.go code/backend/internal/app/composition/awd_defense_ssh_gateway.go code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go code/backend/internal/app/composition/runtime_http_service_adapter.go code/backend/internal/module/instance/infrastructure/proxy_ticket_store_test.go code/backend/internal/config/config.go docs/architecture/backend/03-container-architecture.md docs/operations/runtime-agent-deployment.md docs/operations/awd-host-reboot-recovery-drill.md && git commit -m "feat(backend): 支持 SSH gateway 多副本摘流" -m "为 AWD defense SSH gateway 增加 readiness/draining 生命周期，并验证共享 ticket 与 host key 的跨副本契约。" -m "同步 TCP LB 接流、摘流和会话中断边界说明，避免把多副本误写成透明迁移。" -m "Task: 2026-06-12-ssh-gateway-ha-and-draining"`

## Validation

- Commands:
  - `cd code/backend && go test ./internal/module/instance/infrastructure -run ProxyTicket -count=1`
  - `cd code/backend && go test ./internal/module/instance/application -run ProxyTicket -count=1`
  - `cd code/backend && go test ./internal/app/composition -run 'AWDDefenseSSH.*(Ticket|HostKey|Ready|Drain|Stop|Authenticate)' -count=1`
  - `cd code/backend && go test ./internal/bootstrap -run AWDDefenseSSHGateway -count=1`
  - `git diff --check -- code/backend/internal/bootstrap/awd_defense_ssh_gateway.go code/backend/internal/app/composition/awd_defense_ssh_gateway.go docs/architecture/backend/03-container-architecture.md docs/operations/runtime-agent-deployment.md docs/operations/awd-host-reboot-recovery-drill.md`
- Manual checks:
  - 两个 gateway 副本 behind TCP LB，客户端看到同一 host key fingerprint。
  - 副本 A 签发 ticket，副本 B 接收 SSH 登录并成功查到 Redis claims / PostgreSQL scope。
  - 对副本 B 发 drain 后，新连接不再进入 B；已有连接在 drain deadline 内不被主动关闭。
  - hard Stop 后已有连接关闭，资源释放，进程退出。
- Review focus:
  - draining 是否只是摘流，不承诺 session 透明迁移。
  - host key 是否依赖 T2 的共源 contract，而不是每副本 fallback 生成。
  - ticket 校验是否仍以 Redis/PostgreSQL 为共享事实，gateway 没有引入本地状态。
  - shutdown 顺序是否避免过早 cancel active session context。
