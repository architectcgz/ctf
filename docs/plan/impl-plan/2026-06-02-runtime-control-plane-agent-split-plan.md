# 2026-06-02 runtime control plane / agent split plan

## Objective

- 把当前混在 `ctf-api` 里的“控制面编排”和“宿主机执行面”拆开，形成可扩展到多 Docker 宿主机的清晰边界。
- 让 API 节点只负责认证、权限、业务状态、实例生命周期编排、节点选择和审计，不再直接执行宿主机 `iptables` 或依赖本地 bind mount 驱动远端容器。
- 为“API 主机与 Docker 宿主机分离”的正式部署形态提供一条可渐进落地的实现路径，而不是继续在 API 进程里堆本地 / 远端分支判断。

## Non-goals

- 不在本轮把整套 runtime 重写成 Kubernetes、Nomad 或其他编排平台。
- 不在本轮把 `iptables` 迁移为 `nftables`、eBPF 或 Docker AuthZ plugin；本轮先收口 owner 与调用边界。
- 不在第一阶段就完成多节点智能调度、故障自动迁移、容量均衡或跨机房部署。
- 不顺手重构题目导入、CAS、通知、报表等与 runtime host 执行面无关的模块。

## Inputs

- `code/backend/internal/module/runtime/infrastructure/engine.go`
- `code/backend/internal/module/runtime/infrastructure/engine_files.go`
- `code/backend/internal/module/runtime/infrastructure/engine_provisioning.go`
- `code/backend/internal/module/runtime/infrastructure/acl.go`
- `code/backend/internal/module/runtime/ports/container_runtime.go`
- `code/backend/internal/module/contest/infrastructure/docker_checker_runner.go`
- `code/backend/internal/app/composition/runtime_module.go`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/config/config.go`
- `docker/ctf/docker-compose.dev.yml`
- `docs/architecture/backend/01-system-architecture.md`

## Current problem statement

- 当前 runtime engine 与 checker runner 都通过 `client.FromEnv` 构造 Docker client，说明“远程 Docker Engine”这一层本身已经支持从环境变量切换连接目标。
- 但 API 进程仍直接承担宿主机执行职责：
  - `code/backend/internal/module/runtime/infrastructure/acl.go` 在 API 本机直接执行 `iptables`。
  - `code/backend/internal/module/contest/infrastructure/docker_checker_runner.go` 先在 API 本地落盘 checker 文件，再通过 bind mount 挂到容器。
  - `engine_files.go`、container exec / copy 等调用路径默认假设“发起调用的进程就处在正确的宿主机执行边界”。
- 这导致当前系统真实架构不是“API 调远端 runtime”，而是“API 同时是控制面和执行面”，从而带来三个结构问题：
  - 只切 `DOCKER_HOST` 不能让 ACL 生效到正确宿主机，规则会打到 API 本机。
  - checker sandbox 在远端 Docker 场景下会因为本地 bind mount 路径不在远端宿主机存在而失败。
  - 后续扩到多宿主机时，API 必须持续理解低层 Docker / iptables / 宿主目录细节，边界会越来越散。

## Working design

### End-state architecture

- `ctf-api` 是控制面：
  - 负责认证、权限、业务状态、实例生命周期状态机、节点选择、审计与对外 API。
  - 不负责宿主机级 `iptables`、checker 临时文件落地、本地 bind mount、远端 Docker 文件注入这些执行面细节。
- 每个 Docker 宿主机运行一个 `runtime-agent`：
  - 负责容器 / 网络 / 镜像操作。
  - 负责 ACL 下发与回收。
  - 负责 checker sandbox 执行。
  - 负责容器文件复制、容器内 exec、宿主机临时目录管理等执行面副作用。
- API 与 agent 之间使用显式协议通信，推荐 `gRPC + mTLS`。
- API 侧引入 runtime node 概念，让实例、AWD service、checker 任务最终都绑定明确的 `node_id`，agent 只执行本节点任务。

### Incremental landing strategy

- 第一阶段不直接大改业务流程，先把“谁负责宿主机执行”从 `ctf-api` 代码里收口为清晰 port。
- 第二阶段在不改变 application 语义的前提下，增加 `remote-agent` adapter，与当前本地 adapter 并存。
- 第三阶段再引入节点注册、节点选择和多宿主调度；不要在边界没收清之前先堆调度逻辑。

### Boundary recommendation

- API 侧推荐收口为三类执行能力：
  - `RuntimeHostExecutor`：容器、网络、镜像、容器文件与 exec 等执行能力。
  - `RuntimeACLManager`：实例级 ACL apply / remove。
  - `CheckerExecutionService`：AWD checker sandbox 执行。
- 当前 `runtime/ports/container_runtime.go` 已经承载了部分低层能力。第一阶段不要求一次性重新设计所有 runtime port，但必须明确哪些能力属于“host execution boundary”，后续统一从 agent 走。
- `runtime-agent` 内部可继续复用现有 `Engine`、`acl.go`、`docker_checker_runner.go` 的实现，不把这些逻辑复制两份。

## Task slices

### Slice 1：明确控制面 / 执行面 owner，并为远端执行抽 port

- Goal：让 API 代码先不再直接依赖“当前进程就在正确宿主机上”这一隐含假设。
- Touched files：
  - `code/backend/internal/module/runtime/ports/container_runtime.go`
  - `code/backend/internal/module/runtime/infrastructure/engine.go`
  - `code/backend/internal/module/runtime/infrastructure/engine_files.go`
  - `code/backend/internal/module/runtime/infrastructure/engine_provisioning.go`
  - `code/backend/internal/module/contest/infrastructure/docker_checker_runner.go`
  - `code/backend/internal/app/composition/runtime_module.go`
  - `code/backend/internal/app/composition/instance_module.go`
- Implementation notes：
  - 梳理当前 runtime / checker / ACL 的执行入口，给出“控制面 port”与“宿主机执行 port”的明确 owner。
  - 保留本地实现作为 `local` adapter，但 application / composition 不再直接 new 低层执行细节。
  - 不要求这一刀就统一所有接口命名；重点是把后续要进 agent 的能力单独收口。
- Validation：
  - `cd code/backend && go test ./internal/app/composition ./internal/module/runtime/... ./internal/module/contest/... -count=1`
- Review focus：
  - API 是否已经只依赖抽象执行能力，而不是隐含依赖本地宿主机环境
  - 没有把 port 设计成“远程 Docker 客户端的机械透传”

### Slice 2：引入 runtime agent 协议与 remote adapter

- Goal：把“执行能力可以在远端宿主机落地”这件事做成正式协议，而不是靠 SSH 或临时命令拼接。
- Touched files：
  - 新增 `code/backend/internal/module/runtime/agentcontracts/*`
  - 新增 `code/backend/internal/module/runtime/infrastructure/agentclient/*`
  - 新增 `code/runtime-agent/*` 或项目内等价 agent 目录
  - `code/backend/internal/app/composition/runtime_module.go`
  - `code/backend/internal/config/config.go`
- Implementation notes：
  - 协议层推荐 `gRPC + mTLS`，并显式区分：
    - container / network operations
    - file copy / exec
    - ACL apply / remove
    - checker run
    - node health / capability report
  - 先提供单节点配置，允许 API 明确指向一个 remote agent。
  - 当前本地实现保留为 fallback / dev mode，用于平滑迁移和本地验证。
- Validation：
  - `cd code/backend && go test ./internal/module/runtime/... ./internal/app/composition -count=1`
  - `cd code/runtime-agent && go test ./... -count=1`（若 agent 单独目录存在）
- Review focus：
  - 协议是否表达“宿主机执行语义”，而不是暴露过多 Docker 私有细节
  - TLS 身份与节点配置是否足够明确，避免后续继续补洞

### Slice 3：把 checker 文件传输从本地 bind mount 改成宿主机侧 owner

- Goal：消除 `docker_checker_runner.go` 对“API 本机目录必须与执行容器宿主同机”的依赖。
- Touched files：
  - `code/backend/internal/module/contest/infrastructure/docker_checker_runner.go`
  - `code/backend/internal/module/contest/infrastructure/docker_checker_runner_test.go`
  - agent 侧 checker 执行实现
  - 相关 composition / wiring 文件
- Implementation notes：
  - 废弃“API 本地 `materializeCheckerFiles()` + bind mount 到 checker 容器”的主路径。
  - 推荐新路径：
    - API 把 checker 文件内容作为执行请求发送给 agent。
    - agent 在宿主本地临时落盘，或直接打 tar 流。
    - agent 用 Docker API `CopyToContainer` 或等价机制把文件送入容器。
  - `HostWorkRoot` 语义改成 agent 本机的临时工作目录，而不是 API 本机目录。
- Validation：
  - `cd code/backend && go test ./internal/module/contest/infrastructure -count=1`
  - agent 侧 checker 集成测试
- Review focus：
  - 远端执行路径里是否彻底移除了 API 本地 bind mount 依赖
  - checker 文件生命周期与清理是否仍然可控

### Slice 4：把 ACL owner 从 API 本机迁到 agent

- Goal：让实例级 ACL 明确落在容器真实宿主机执行，而不是 API 所在机器。
- Touched files：
  - `code/backend/internal/module/runtime/infrastructure/acl.go`
  - `code/backend/internal/module/runtime/infrastructure/engine_provisioning.go`
  - `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
  - `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
  - agent 侧 ACL 执行实现
  - 相关测试
- Implementation notes：
  - API 侧保留 `ApplyACL` / `RemoveACL` 业务语义，但实际执行改为通过 agent 下发。
  - 当前实例级 chain / handle 模型继续保留；本轮迁的是 owner，不是重新改 ACL 语义。
  - 如果某些环境暂不启用 agent，本地模式仍可复用现有 ACL 实现，但这只是 dev / fallback，不再作为正式多机方案。
- Validation：
  - `cd code/backend && go test ./internal/module/runtime/infrastructure ./internal/module/runtime/application/commands ./internal/module/runtime -count=1`
  - agent 侧 ACL 执行与回滚测试
- Review focus：
  - ACL authority 是否已经明确在宿主机执行侧
  - API 侧是否不再需要知道 `iptables` 命令细节

### Slice 5：引入 runtime node 模型与节点选择

- Goal：让“多 Docker 宿主机”成为明确的数据模型，而不是配置里隐藏的一根远端地址。
- Touched files：
  - 新增 runtime node 实体 / repository / migration
  - `code/backend/internal/module/runtime/*`
  - `code/backend/internal/module/practice/application/commands/*`
  - `code/backend/internal/module/contest/application/*`
  - 相关 handler / DTO / query / tests
- Implementation notes：
  - 推荐最小 node 模型字段：
    - `id`
    - `name`
    - `endpoint`
    - `tls_identity`
    - `schedulable`
    - `labels`
    - `health_status`
    - `capacity_snapshot`
  - 实例、AWD service、checker 执行都绑定 `node_id`。
  - 第一版调度策略只做最简单的“显式绑定或单节点默认”，不要提前做复杂均衡。
- Validation：
  - `cd code/backend && go test ./internal/module/runtime/... ./internal/module/practice/... ./internal/module/contest/... -count=1`
- Review focus：
  - `node_id` 是否成为明确 authority，而不是继续靠配置猜测当前宿主机
  - 调度策略是否保持最小正确，不提前引入复杂漂移

### Slice 6：更新事实源、部署文档与开发模式说明

- Goal：把当前“多机代码无需修改 / 通过配置文件切换 DOCKER_HOST”这类过度简化说法改成和真实架构一致的事实源。
- Touched files：
  - `docs/architecture/backend/01-system-architecture.md`
  - `README.md`
  - `docs/operations/*`（若新增 agent 部署说明）
  - 需要时补 `docs/contracts/` 的 agent 协议说明
- Implementation notes：
  - 明确区分：
    - 本地开发单机模式
    - API + 单 remote agent 模式
    - 多 node 正式模式
  - 把文档中的“通过配置文件切换 `DOCKER_HOST`，代码无需修改”修正为：
    - 纯 Docker client 连接目标可通过环境变量切换
    - 但完整多机正式架构需要 agent 承接 ACL / checker / host execution
  - 若新增 agent 配置或 node 管理入口，同步补运维文档与最小部署步骤。
- Validation：
  - `bash scripts/check-consistency.sh`
- Review focus：
  - 文档是否与真实能力边界一致
  - 不再把 dev compose 的临时路径写成正式架构建议

## Expected change surface

- runtime ports / infrastructure / application wiring
- contest checker sandbox execution path
- runtime ACL execution boundary
- 新增 runtime agent 进程与 API -> agent 协议
- runtime node 数据模型、调度入口和配置
- 架构与运维文档

## Data / API / compatibility impact

- 新增 runtime node 数据模型与实例 / AWD service 对 `node_id` 的绑定关系。
- 本地开发模式短期保留，避免一次性让所有现有流程切到 remote agent。
- `DOCKER_HOST` 仍可作为低层 Docker client 的连接参数，但不再是“完整多机方案”的唯一切换开关。
- checker sandbox 配置里的 `host_work_root` 将从“API 本机目录”语义改成“agent 宿主本机目录”语义。
- 对外 API 第一阶段可不暴露 node 细节；等 node 绑定进入业务语义后，再决定是否开放运维 / 管理接口。

## Validation matrix

- API 在不持有本机 Docker socket 的情况下，仍可通过 remote agent 完成实例创建、启动、删除和网络操作。
- ACL 下发与回收发生在容器真实宿主机，而不是 API 所在主机。
- checker sandbox 在远端宿主机模式下无需依赖 API 本地 bind mount。
- 单节点本地 fallback 模式仍可支撑现有开发链路。
- 实例 / AWD service / checker 执行可明确追溯到 `node_id`。
- 架构文档、README、运维说明与真实多机落地路径一致。

## Review fit check

- Owner 清晰：
  - API 负责控制面编排与状态 authority。
  - agent 负责宿主机执行与副作用。
  - Docker / iptables / 宿主临时目录不再散落在 API 各处。
- Reuse 点清晰：
  - 当前 `Engine`、`acl.go`、`docker_checker_runner.go` 尽量复用到 agent 内部。
  - application 语义尽量保持稳定，先换执行 owner，不先改业务流程。
- 结构收敛：
  - 不再依赖“API 进程恰好就在正确宿主机”这个隐含假设。
  - 多机能力建立在显式 node / agent 协议上，而不是局部配置技巧。
- 已知债触达：
  - 当前文档把“切 `DOCKER_HOST` 即可多机”讲得过于乐观，本轮必须同步修正。
  - 当前 checker bind mount 与 ACL 本地执行都是 touched surface 内必须收口的结构债，不能留成 residual risk。

## Rollback / recovery

- 第一阶段的 port 收口与本地 adapter 保留后，可在远端 agent 路径出现问题时临时回退到本地执行模式。
- 只要 node 绑定还未成为强制数据约束，就可以通过配置切回单节点本地 fallback，降低引入 agent 初期的切换风险。
- 一旦开始把实例 / 服务真实绑定到 `node_id`，回退时必须同时处理：
  - 数据模型兼容
  - 文档事实回退
  - 运维部署回退

## Open implementation choices

- Agent 进程形态：
  - 推荐独立二进制 `runtime-agent`
  - 不推荐把 agent 混进现有 `ctf-api` 进程用 mode flag 分叉
- 协议：
  - 推荐 `gRPC + mTLS`
  - 不推荐 SSH 命令执行或 HTTP + 自定义签名这类后补安全模型
- checker 文件注入：
  - 推荐 agent 侧落盘后 `CopyToContainer`
  - 不推荐继续依赖共享文件系统或 NFS 把 bind mount 勉强维持下去
- 节点调度：
  - 推荐先做显式 node 绑定 / 单节点默认
  - 不推荐在边界未收清前先做复杂自动均衡
