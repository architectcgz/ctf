# 2026-06-02 runtime ACL authority hardening plan

> 后续：legacy cleanup fallback 已由 `docs/plan/impl-plan/2026-06-03-runtime-acl-legacy-fallback-retirement-plan.md` 继续推进；本文中关于 fallback 窗口的描述仅保留当时的实施背景。

## Objective

- 把 runtime ACL 从“数据库中的可回放规则明细”收口为“runtime 拥有的宿主机防火墙资源”。
- 让 cleanup 不再依赖 `runtime_details.acl_rules` 作为 authority source，而是按实例级 ACL handle / chain 删除。
- 在 `iptables` 执行前对所有进入宿主机防火墙的字段做白名单校验与 canonicalize，避免把半可信数据直接翻译成高权限网络策略。

## Non-goals

- 不把当前 `iptables` 实现整体替换成 `nftables`、Docker plugin 或独立 sidecar firewall service。
- 不顺手重构 topology policy 的业务建模，也不扩大到 `allow_outbound`、静态基础规则初始化脚本的整套治理。
- 不在本轮移除 `runtime_details.acl_rules` 的 JSON 字段；本轮先把它降级为调试快照，避免影响现有运行实例的短期兼容。

## Inputs

- `code/backend/internal/module/runtime/domain/topology_acl.go`
- `code/backend/internal/module/runtime/contracts/runtime_details.go`
- `code/backend/internal/module/runtime/domain/resources.go`
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
- `code/backend/internal/module/runtime/infrastructure/acl.go`
- `code/backend/internal/module/runtime/infrastructure/engine_provisioning.go`
- `code/backend/internal/module/runtime/ports/container_runtime.go`
- `code/backend/internal/module/runtime/service_test.go`
- `docs/architecture/backend/03-container-architecture.md`
- `docs/todos/2026-06-02-security-review-findings.md`

## Current problem statement

- 当前 `ResolveTopologyACLRules()` 在 domain 层把拓扑策略解析成 `InstanceRuntimeACLRule`。
- provisioning 在实例创建成功后调用 `ApplyACLRules()` 下发规则，并把整份 `resolvedACLRules` 写入 `runtime_details.acl_rules`。
- cleanup 通过 `ExtractManagedResources()` 从 `runtime_details` 中读回 `ACLRules`，再调用 `RemoveACLRules()` 按明细回放删除。
- 这导致 `InstanceRuntimeACLRule` 同时承担“领域意图 / 运行时事实 / 可执行防火墙参数”三种职责，并把数据库里的 `runtime_details.acl_rules` 变成了宿主机 `iptables` 的 authority source。

## Working design

- `TopologyTrafficPolicy` 和 `InstanceRuntimeACLRule` 继续承担 ACL 语义与解析结果，但不再作为 cleanup authority。
- 新增实例级 `ACL handle` 概念，初始形态使用实例级专用 chain 名称。
- 细粒度 ACL 的宿主机资源模型改为：
  - 创建时：创建实例级 chain，向 chain 写入规则，再把 jump 挂到 `DOCKER-USER`
  - 清理时：按 handle 删除 jump、flush chain、delete chain
- `runtime_details.acl` 保存 cleanup authority；`runtime_details.acl_rules` 过渡期保留为调试快照，不参与正常 cleanup 决策。
- 对过渡期老实例（只有 `acl_rules`，没有 `acl` handle），cleanup 继续支持一次带校验的 legacy fallback，避免本轮上线后遗留脏规则。

## Task slices

### Slice 1：引入 ACL handle，切换 cleanup authority

- Goal：把 runtime details / managed resources / cleanup owner 从 `ACLRules` 切到 `ACL handle`。
- Touched files：
  - `code/backend/internal/module/runtime/contracts/runtime_details.go`
  - `code/backend/internal/module/runtime/domain/resources.go`
  - `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
  - `code/backend/internal/module/runtime/ports/container_runtime.go`
  - `code/backend/internal/module/runtime/infrastructure/engine_provisioning.go`
  - `code/backend/internal/module/runtime/service_test.go`
- Implementation notes：
  - 在 `InstanceRuntimeDetails` 中新增 `ACL *InstanceRuntimeACLHandle` 字段，保存实例级 ACL resource handle。
  - `ManagedResources` 改为提取 `ACLHandle`，不再把 `ACLRules` 作为 cleanup 主路径输入。
  - `ContainerCleanupRuntime` 的删除能力改名或改签名为按 handle 删除，例如 `RemoveACL(ctx, handle)`。
  - cleanup 主路径先走 `ACLHandle`；仅当没有 handle 时才考虑 legacy fallback。
- Validation：
  - `cd code/backend && go test ./internal/module/runtime/contracts ./internal/module/runtime/domain ./internal/module/runtime/application/commands`
- Review focus：
  - cleanup authority 是否已经明确从 `acl_rules` 切到 `acl handle`
  - 旧实例兼容是否只作为 fallback，而不是继续保留为主路径

### Slice 2：实例级 chain 模型落地到 infrastructure

- Goal：把 ACL 资源从“按规则明细删除”改成“按实例级 chain 生命周期管理”。
- Touched files：
  - `code/backend/internal/module/runtime/infrastructure/acl.go`
  - `code/backend/internal/module/runtime/infrastructure/engine_provisioning.go`
  - `code/backend/internal/module/runtime/ports/container_runtime.go`
  - `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
  - `code/backend/internal/module/runtime/service_test.go`
- Implementation notes：
  - 使用实例级 chain 命名，例如 `CTF-INS-<instanceID>`。
  - `OwnerInstanceID <= 0` 且存在细粒度 ACL 时直接报错，不允许创建没有稳定 owner 的 ACL 资源。
  - `ApplyACLRules()` 负责：
    - ensure / reset chain
    - 向 chain 写入规则
    - 挂接 `DOCKER-USER -> chain` jump
  - `RemoveACL()` 负责：
    - 删除 jump
    - flush chain
    - delete chain
  - `runtime_details.ACL` 在 provisioning 成功后写入；`runtime_details.ACLRules` 只保留调试快照。
- Validation：
  - `cd code/backend && go test ./internal/module/runtime/infrastructure ./internal/module/runtime/application/commands ./internal/module/runtime -run 'TestServiceCreateTopology|TestServiceCleanupRuntime' -count=1`
- Review focus：
  - `iptables` 资源 owner 是否已经收口到实例级 chain
  - provisioning / cleanup 是否不再依赖“重新拼出同一条 rule”这类脆弱匹配

### Slice 3：执行前白名单校验与 canonicalize

- Goal：即使走 legacy fallback，也不允许半可信 ACL 明细直接驱动 `iptables` 参数。
- Touched files：
  - `code/backend/internal/module/runtime/infrastructure/acl.go`
  - `code/backend/internal/module/runtime/domain/topology_acl.go`
  - `code/backend/internal/module/runtime/service_test.go`
  - 新增 `code/backend/internal/module/runtime/infrastructure/acl_test.go`（如当前目录没有更合适的测试文件）
- Implementation notes：
  - `SourceIP` / `TargetIP`：使用 `net/netip` 解析，只允许单 IP。
  - `Action`：白名单 `allow` / `deny`。
  - `Protocol`：白名单 `any` / `tcp` / `udp`。
  - `Ports`：1-65535、去重排序、`multiport` 上限 15，`protocol=any` 时禁止携带端口。
  - `Comment`：执行前统一系统重建，不信任持久化值。
  - legacy fallback 若遇到无效 ACL 明细，应记录结构化警告并跳过 / fail fast，不能把原值直接交给 `iptables`。
- Validation：
  - `cd code/backend && go test ./internal/module/runtime/infrastructure -count=1`
- Review focus：
  - trust boundary 是否真的落在 infrastructure
  - comment / ip / protocol / ports 是否都做了明确 owner 收口，而不是只 trim string

### Slice 4：补兼容回归测试并更新事实源

- Goal：证明“新实例按 handle 清理、老实例按 fallback 清理、脏 `acl_rules` 不再影响主路径”，同时更新架构事实。
- Touched files：
  - `code/backend/internal/module/runtime/service_test.go`
  - `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service_test.go`（如更适合）
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/todos/2026-06-02-security-review-findings.md`
- Implementation notes：
  - 新增测试覆盖：
    - 新实例创建后写入 `details.ACL`
    - cleanup 优先按 `ACLHandle` 删除
    - 老实例只有 `acl_rules` 时仍可 cleanup
    - `acl_rules` 被污染时，新实例 cleanup 不受影响
    - `OwnerInstanceID <= 0` 且有细粒度 ACL 时创建失败
  - 架构文档把“按 `runtime_details.acl_rules` 精确回收”更新为“按实例级 ACL handle 清理；`acl_rules` 仅为调试快照 / 兼容回退”。
  - todo 若本轮收口完成，更新状态或替换为剩余遗留项；不要让 backlog 与实现结论漂移。
- Validation：
  - `cd code/backend && go test ./internal/module/runtime/... -count=1`
  - `bash scripts/check-consistency.sh`
- Review focus：
  - 文档是否与新 authority model 对齐
  - 兼容测试是否真的覆盖旧实例和污染数据两个高风险路径

## Expected change surface

- runtime contracts / ports / cleanup / provisioning / infrastructure ACL adapter
- runtime service and infrastructure tests
- container architecture fact source
- security todo 跟踪项

## Data / API / compatibility impact

- `runtime_details` JSON 结构会新增 `acl` 字段，保存实例级 ACL handle。
- `acl_rules` 字段本轮不删除，作为调试快照保留；后续 fallback 退场见 `2026-06-03-runtime-acl-legacy-fallback-retirement-plan.md`。
- 细粒度 ACL 创建将要求 `OwnerInstanceID` 存在；如果调用方无法提供稳定实例 owner，需要先回到调用链补 owner，再创建 topology ACL。
- cleanup authority 目标：
  - 新实例：`ACLHandle` 为 authority
  - 旧实例：先通过 runtime startup migration 补齐 `ACLHandle`
  - `ACLRules` 只保留为调试快照，不再作为长期 cleanup owner

## Validation matrix

- 新创建 topology 实例在 `runtime_details.acl` 中记录实例级 handle。
- cleanup 对新实例只按 handle 删除，不依赖 `acl_rules` 明细。
- 旧实例在 startup migration 后也按 `ACLHandle` 完成 cleanup。
- 旧实例 `acl_rules` 含非法字段时，不会把原始污染值直接交给 `iptables`。
- `OwnerInstanceID <= 0` 且存在细粒度 ACL 时，创建流程失败并返回明确错误。
- `docs/architecture/backend/03-container-architecture.md` 与实际 authority model 一致。

## Review fit check

- Owner 清晰：
  - domain 负责 ACL 意图和解析结果
  - application 负责把 ACL 绑定到实例 owner 与运行时事实
  - infrastructure 独占 `iptables` 翻译权和 chain 生命周期
- Reuse 点清晰：
  - 继续复用现有 topology policy 解析逻辑
  - 复用 `OwnerInstanceID` 作为 chain owner
  - 复用 `runtime_details` 作为实例运行时资源句柄的持久化入口
- 结构收敛：
  - 不再把数据库中的 `acl_rules` 当作宿主机防火墙 authority source
  - cleanup 从“按规则回放删除”收口为“按实例级资源删除”
- 已知债触达：
  - 本次正面收口 `iptables` 参数来自数据库字段的安全债
  - touched surface 内不再把这类 debt 继续留成 residual risk

## Rollback / recovery

- 若新 chain 模型上线后出现问题，可回退本次 runtime ACL 改动；`runtime_details.acl` 是附加字段，不涉及 schema migration。
- 回退时必须保留 legacy fallback，确保当前运行中的实例仍可按旧 `acl_rules` 清理。
- 若回退代码，需同时回退架构文档和 todo 状态更新，并重新确认 legacy 实例的 cleanup 策略，避免事实源漂移。

## Open implementation choices

- 推荐 chain 命名：`CTF-INS-<instanceID>`。
- 如果需要更强隔离，可在 chain 名称前加环境前缀，例如 `CTF-DEV-INS-<instanceID>`；但本轮默认先用实例级命名，避免扩大配置面。
- 若 infrastructure 里已经存在静态基础规则初始化逻辑，本轮不合并改造；动态 ACL 只负责实例级细粒度规则。
