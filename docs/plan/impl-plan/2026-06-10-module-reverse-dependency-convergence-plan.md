# 模块反向依赖收口方案（container_runtime→contest、instance→contest）

> 状态：Draft
> 事实源：`code/backend/internal/module/architecture_baseline_test.go` 当前 baseline、`docs/design/backend-module-boundary-target.md` 目标版图、各模块当前 import 事实
> 替代：无；落地后稳定结论回收到 `docs/architecture/backend/07-modular-monolith-refactor.md`

## 定位

本方案收口 `moduleDependencyBaseline` 中两条反向依赖边：`container_runtime -> contest` 与 `instance -> contest`。

- 负责：给出这两条边的根因、收口方向、阶段切片、兼容回退、验证和 review 重点。
- 不负责：改 AWD checker 判定逻辑、改 runtime-agent 传输行为、改实例 / AWD 业务行为；不重做 runtime 残余状态拆分（见 `2026-06-10-runtime-residual-state-owner-split-plan.md`）。

## 背景与时序

目标版图把 `container_runtime` 定为最底层平台适配（不依赖业务模块），`contest` / `instance` 为业务 owner。runtime 残余拆分正在消解 `runtime -> contest`、`runtime -> container_runtime` 等边；本方案处理的是模块边界纯净化的**下一层**：那些不属于 runtime、但方向仍然反了的边。

时序约束：

- 边1（`container_runtime -> contest`）独立，与 runtime 拆分无强耦合，可立即推进。
- 边2（`instance -> contest`）的主体是 instance 越界实现了 contest 的 AWD 准入查询；它与 runtime 拆分块3（`awd_scope_controls` owner 收口 contest）方向一致。**边2 应在块3 之后落地**，让 instance 里重复的 scope control helper 副本与准入查询一起迁走。

## 输入与现状

### 边1：container_runtime → contest（checker 执行契约）

- 根因文件：
  - `container_runtime/infrastructure/agentclient/bridge.go`：`Bridge` 实现 `contestports.CheckerRunner`，`RunChecker` 把 checker 转发到远程节点。
  - `container_runtime/infrastructure/agentserver/service.go`：`Service` 持有 `contestports.CheckerRunner`，agent 端执行 checker。
  - `container_runtime/agentcontracts/messages.go`：agent 消息体直接内嵌 `contestports.CheckerRunJob` / `CheckerRunResult`。
- 契约结构（`contest/ports/checker_runner.go`）：
  - `CheckerRunJob`：`Runtime / Image / Entry / Args / Env / Files / OutputMode / NetworkMode / TargetAllowlist / Timeout / Limits / Metadata`。
  - `CheckerRunResult`：`Status / Reason / ExitCode / Stdout / Stderr / Duration / OutputLimitHit / ResourceLimitHit / StartedAt / FinishedAt`。
- 关键洞察：`CheckerRunJob` / `CheckerRunResult` **几乎全是通用的"沙箱化容器进程执行"参数**（镜像、入口、参数、环境、文件、网络模式、资源限制、超时、退出码、输出）。唯一带业务语义的是 `CheckerRunMetadata`（`ContestID / ServiceID / TeamID / RoundNumber / NodeID`）。
- 结论：这个契约放错了模块。"在容器节点上跑一个受限进程并取回结果"是 container_runtime 的平台能力；contest 只应定义 AWD checker 的业务语义（哪种 checker、判定规则）。当前方向反了——contest 定义执行契约，container_runtime 去实现它。
- 消费/实现拓扑：contest 侧有 `docker_checker_runner`、`awd_http_checker_sandbox`、`awd_script_checker_runner`、`awd_tcp_checker_runner` 等多种 checker 实现与 `checker_audit`；container_runtime agent 是远程执行通道；composition（`contest_module.go` / `runtime_module.go` / `runtime_node_execution_router.go`）与 `bootstrap/runtime_agent.go` 负责 wiring。

### 边2：instance → contest（实为三类，主体是 owner 错位）

读真实使用点后，`instance -> contest` 不是笼统的"AWD 实例需要 contest 业务事实"，而是三类性质不同的依赖：

- 类A：仅共用枚举常量（最轻）。`instance/application/queries/instance_service.go:155` 用 `contestcontracts.ContestModeAWD` 和实例自己的 `inst.ContestMode` 比较；`instance/infrastructure/repository.go` 用 `ContestModeAWD`、`ContestRegistrationStatusApproved`、`ContestStatusRunning` 等字符串常量。属共享枚举值，不读 contest 数据。
- 类B：instance 越界实现了 contest 的 AWD 准入规则查询（主体，真问题）。`instance/infrastructure/awd_target_proxy_repository.go` 的 `FindAWDTargetProxyScope` / `FindAWDDefenseSSHScope` 直接 join 了 contest 拥有的表（`contests` / `team_members` / `teams` / `contest_awd_services` / `awd_rounds` / `awd_scope_controls`），把"轮次进行、AWD 模式、竞赛在进行/冻结、攻防队伍未 retired、服务未 disabled"整套准入规则实现在 instance 里，还自带一份 `joinAWDActiveScopeControls` / `applyAWDActiveScopeFilter` 副本（与 contest、runtime 三处重复）。`repository.go` 列实例时 join `contest_registrations` + AWD service 可见性过滤同属此类。这不是"需要 contest 事实"，是准入规则 owner 错位。
- 类C：消费 contest 数据 / 输入契约（相对合理）。`startup_runtime_recovery_service.go` 经 port 拿 `contestcontracts.Contest`（`AddPausedDurationToActiveAWDContests`、竞赛有效结束时间）；`handler.go` 的 `RecordAWDProxyTrafficEvent(AWDProxyTrafficEventInput)`（runtime 拆分块6 已覆盖）；`repository.go` 的 `DecodeContestAWDServiceSnapshot`。
- 与 runtime 拆分的关系：块3 把 `awd_scope_controls` owner 收口到 contest 后，instance 里类B 那份 scope control helper 副本应与本方案一起消除。

## 目标

- 消解 `container_runtime -> contest`：container_runtime 不再依赖任何业务模块。
- 收口 `instance -> contest`：类B 的 AWD 准入查询经依赖倒置（DIP）整条迁回 contest 实现，类A 共享枚举与类C 数据消费按最小改动处理；instance 不再 import `contest/contracts`。
- baseline 中这两条边按真实 import 消失而移除，不靠放宽守卫。

## 非目标

- 不改 AWD checker 的判定逻辑、四种 checker 实现的业务行为。
- 不改 runtime-agent 的 wire 协议语义与 TLS / gRPC 传输行为。
- 不改实例生命周期、AWD 实例业务行为、数据库 schema。
- 不在本方案重复 runtime 残余状态拆分。

## 方案设计

### 边1：沙箱执行能力反转（推荐，无争议）

把 checker 的"执行通道"从 contest 契约里抽出来，下沉为 container_runtime 的中性能力，反转依赖方向。

1. 在 `container_runtime/ports`（或 `agentcontracts`）定义中性的沙箱执行契约，例如 `SandboxExecutor`：
   - `SandboxExecJob`：`Runtime / Image / Entry / Args / Env / Files / OutputMode / NetworkMode / TargetAllowlist / Timeout / Limits`——即去掉业务 `Metadata` 的通用执行参数。
   - `SandboxExecResult`：`Status / ExitCode / Stdout / Stderr / Duration / OutputLimitHit / ResourceLimitHit / StartedAt / FinishedAt`。
   - 业务关联（`ContestID / ServiceID / TeamID / RoundNumber`）作为不透明 label / 关联 ID 透传，container_runtime 不解释其语义。
2. agent 改造：`agentclient/bridge.go`、`agentserver/service.go`、`agentcontracts/messages.go` 改为实现 / 传输中性 `SandboxExecutor` 与 `SandboxExecJob/Result`，不再 import `contest/ports`。
3. contest 适配：`contest` 保留自己的 `CheckerRunner` 业务接口与 `CheckerRunMetadata`，其 checker runner（docker/http/script/tcp）在远程场景下经 container_runtime 的 `SandboxExecutor` 执行——把 `CheckerRunJob` 映射成中性 `SandboxExecJob`，结果反向映射回 `CheckerRunResult`。
4. composition wiring：把 `SandboxExecutor` 能力从 container_runtime 注入 contest 的 checker runner，依赖方向变为 `contest -> container_runtime`。
5. 移除 `container_runtime -> contest` 的 baseline 边。

备选（不推荐）：把 `CheckerRunner/Job/Result` 整体移到中性 shared 包。缺点是 job 字段仍带容器执行语义，放 shared 反而稀释 container_runtime 的能力归属；而本方案的洞察是这些字段本就属于 container_runtime。

### 边2：instance → contest 收口（依赖倒置，按三类分别处理）

核心思路：不拆查询、不"喂上下文让 instance 继续 join contest 表"，而是把 owner 摆正——AWD 准入查询本属 contest，用依赖倒置（DIP）保持合规的 `contest -> instance` 方向。

类B（主体）——依赖倒置 + 查询整体迁 contest：

1. `instance/ports` 定义窄接口（消费方定义），复用已有 `instanceports.AWDTargetProxyScope` 类型：
   `AWDProxyTargetResolver.ResolveAWDTargetProxyScope(ctx, userID, contestID, serviceID, victimTeamID) (*AWDTargetProxyScope, error)`；`FindAWDDefenseSSHScope` 同理定义对应接口。
2. `FindAWDTargetProxyScope` / `FindAWDDefenseSSHScope` 的整条 SQL **原封不动**从 `instance/infrastructure` 迁到 `contest/infrastructure`——contest 本就 own `awd_rounds` / `awd_scope_controls` / `contest_awd_services` / `teams`，join `instances` 表取访问字段即可。性能不变（仍一条 SQL、一次往返）。
3. composition 把 contest 的实现注入 instance proxy handler。依赖方向变为 `contest -> instance/ports`（合规），instance 不再 import `contest/contracts`。
4. 列实例的 AWD 可见性过滤同法：可见性判定经 contest 提供的过滤能力，instance 列表查询不再自带 contest 可见性规则。

为什么不拆两段：拆"contest 准入 + instance 访问"会引入两次往返、跨模块调用链、准入与访问的事务一致性负担；DIP 把同一条查询换个模块落地 + 一个接口，全部避免。

类A（枚举常量）：把共享枚举在中性位置定义或 instance 侧镜像，消除为常量依赖；改动小，并入类B 切片顺手处理。

类C（数据消费）：`RecordAWDProxyTrafficEvent` 由 runtime 拆分块6 覆盖；`startup_runtime_recovery` 的竞赛暂停补偿、`DecodeContestAWDServiceSnapshot` 经 contest 窄 port / 中性结果类型消费，保留为合规的 `contest -> instance` 或显式注入，不强行消灭。

诚实记录的张力：

- contest 的实现直接 join `instances` 物理表——方向 `contest -> instance`（合规、baseline 已有），用 contest 自己的 row struct scan、不 import instance GORM model。
- `AWDTargetProxyScope` 带 AWD 字段，留在 `instance/ports`（现状即如此）算消费方接口的输入输出；要更纯可移中性位置，非必须。
- proxy 若为极热路径，可叠加准入结果缓存作为独立性能优化，与本 owner 方案正交。

落地前提：在 runtime 拆分块3（scope control owner 收口 contest）之后做，两者方向一致、helper 副本一起消除。

## 阶段切片

- 切片 1（边1，立即）：定义 container_runtime 中性 `SandboxExecutor` 契约 + 改 agent bridge/server/messages 用中性类型；contest checker runner 暂时保留经旧路径，仅打通中性契约。
- 切片 2（边1）：contest checker runner 经中性 capability 执行 + composition wiring 反转 + 移除 `container_runtime -> contest` 边 + 更新 baseline 与 container_runtime 架构守卫。
- 切片 3（边2，依赖 runtime 拆分块3）：`instance/ports` 定义 `AWDProxyTargetResolver` 等窄接口，AWD 准入查询整条迁 contest 实现并经 composition 注入；顺带处理类A 枚举常量；移除 instance→contest import 并更新 baseline。类C 中合理的数据消费按合规方向保留。

每个切片各自走 `bash scripts/start-implementation.sh` 绑定 task slug。

## 兼容性与回退

- 边1 agent wire 兼容：保持现有 JSON codec 与字段语义，只换 Go 类型归属与 import 方向；checker 判定结果不变。
- 不改 checker 四种实现的业务行为、不改 AWD 运行行为、不改 schema。
- 每个切片独立可 revert；边2 切片在 runtime 拆分完成前不启动。

## 验证计划

```bash
cd code/backend && go test ./internal/module/container_runtime/... -count=1
cd code/backend && go test ./internal/module/contest/... -count=1
cd code/backend && go test ./internal/module -run TestModuleDependencyBaselineIsCurrent -count=1
cd code/backend && bash ../../scripts/check-backend-architecture.sh --full
```

- 边1 追加：runtime-agent 远程 checker 的集成 / e2e 验证（`runtime_node_execution_router_e2e_test.go` 一类），确认远程 checker 执行结果一致。
- 边2 追加：instance 模块测试 + AWD 实例启动 / 恢复 / proxy 回归。

## Review 重点

- 边1 中性契约不得回带业务语义（container_runtime 不解释 contest/team/service 的含义，只透传关联 ID）。
- agent 协议 wire 行为与 checker 判定结果保持不变。
- baseline 边按真实 import 消失而移除，不靠放宽 archtest。
- 边2 收口后，AWD 准入查询确实迁到 contest 实现、instance 经自己的 `instance/ports` 接口消费；contest 实现不得 import instance 的 GORM model（只读 join + 自有 row struct）；DIP 注入在 composition 完成，instance 不再 import `contest/contracts`。

## 完成判定

- `moduleDependencyBaseline` 不再包含 `container_runtime -> contest`。
- `instance -> contest` 收口：类B 的 AWD 准入查询经 DIP 迁回 contest 实现，instance 不再 import `contest/contracts`；类C 合理数据消费若保留，降为合规的 `contest -> instance` 或显式注入。
- container_runtime 不依赖任何业务模块。
- 全部验证通过，checker 判定、AWD 行为、agent 协议、schema 无变化。
