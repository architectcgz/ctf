# Runtime Node 资源池与访问地址实施计划

> **面向 agent 执行者:** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务执行本计划。步骤使用 checkbox（`- [ ]`）跟踪。

**Goal:** 将题目实例端口 / 网段分配改为按 runtime node 管理的数据库资源池，并补齐 node 级访问地址与实例启动进度事件。

**Architecture:** `container_runtime` 负责 runtime node 元数据、按 node 分片的资源池和分配正确性。`practice` 负责实例启动 / provisioning 编排，并通过窄接口消费 node-scoped allocation 能力。`instance` 负责持久化实例生命周期、当前 provisioning stage 和 provisioning event history。`runtime_nodes.endpoint` 只表示控制面地址；`runtime_nodes.public_host/access_host` 表示数据面访问地址。

**技术栈:** Go, GORM, PostgreSQL, Docker Engine SDK, runtime-agent, Vue presentation tests, 既有 `container_runtime` / `instance` / `practice` 模块。

---

## 任务绑定

- 建议 task slug：`runtime-node-resource-pool-public-hosts`
- 实现前运行：`bash scripts/start-implementation.sh runtime-node-resource-pool-public-hosts`
- 本计划是跨模块正式实施计划，需要经过项目 code-workflow startup gate、plan review、implementation review 和 completion validation。

## Task Metadata

- Task slug：`2026-06-24-runtime-node-resource-pool-public-hosts`
- 分支：`task/2026-06-24-runtime-node-resource-pool-public-hosts`
- Worktree：`.worktrees/ctf/2026-06-24-runtime-node-resource-pool-public-hosts`
- Plan path：`docs/plan/impl-plan/2026-06-24-runtime-node-resource-pool-public-hosts-implementation-plan.md`

## Task Classification

- 类型：后端结构性实现，包含 schema、repository、composition wiring 和契约测试。
- 复杂度：非琐碎，跨 `container_runtime`、`instance`、`app/composition` 和 migration。
- 风险面：运行节点资源唯一性、启动期 seed 写放大、访问地址 fallback、实例 provisioning 可观测性。

## Files

- 数据库：`code/backend/migrations/000001_init_schema.up.sql`。
- Runtime node：`code/backend/internal/module/container_runtime/**` 与 `code/backend/internal/app/composition/**`。
- Instance：`code/backend/internal/module/instance/entity/**`。
- 测试：runtime node migration、node repository、resource pool repository、runtime module composition。

## 复用与 Owner 决策

- 资源池 owner 放在 `container_runtime`，复用既有 runtime node repository、GORM entity 和 `subnetCandidates` CIDR 生成逻辑。
- `instance` 只持有实例生命周期与 provisioning event 数据结构，不反向承担 runtime node 分配决策。
- Composition 只负责默认节点启动 seed 接线，不把 schema owner 或资源分配规则放入启动组装层。

## Intake Analysis Gate

- 当前任务已绑定项目 code-workflow startup gate，并按正式实施计划推进。
- 实现前置判断采用项目 `AGENTS.md`、`ctf-backend-patterns`、后端测试分层说明和资源池计划的 owner 划分。
- Review 反馈的 P1 写放大问题已作为任务 4 收口修复点处理，未扩大到任务 5 之后的 provisioning flow。

## Validation

- 资源池 repository：`go test ./internal/module/container_runtime/infrastructure -run TestResourcePool -count=1`。
- Runtime module seed：`go test ./internal/app/composition -run TestBuildContainerRuntimeModule -count=1`。
- 提交前最小充分验证：`go test ./internal/module/container_runtime/infrastructure ./internal/app/composition -count=1`。
- Pre-commit gate：`bash scripts/run-workflow-stage.sh pre-commit-quick` 或正常 `git commit` hook。

## 背景

当前行为：

- 宿主端口通过扫描全局 `[container.port_range_start, container.port_range_end)` 并写入 `port_allocations(port)` 分配。
- 网络子网通过扫描配置里的 CIDR 候选并写入 `network_allocations(subnet)` 分配。
- Access URL 只从全局 `container.public_host/access_host` 生成。
- `runtime_nodes` 表示 Docker daemon / runtime-agent 执行宿主，但还没有保存学生侧访问地址元数据。
- 实例启动只暴露 `pending / creating / running / failed` 这类粗粒度状态。

目标行为：

- 必须先选择 runtime node，再分配端口 / 子网。
- 端口和 Docker bridge 子网按 `runtime_node_id` 唯一，不再要求全局唯一。
- 不同 node 复用同一端口时，使用 node 级 `public_host/access_host` 生成正确访问地址。
- 资源池分配使用数据库行锁（`FOR UPDATE SKIP LOCKED`），替代反复 insert-conflict 探测。
- 实例启动进度通过当前阶段字段和 append-only 事件表可见。

## 范围

本次包含：

- runtime node host、资源池和 provisioning events 相关数据库 schema。
- GORM entity 和 repository。
- 基于现有配置生成 node-scoped pool。
- node-scoped 端口 / 子网 reserve、bind、release、quarantine 流程。
- 基于选中 runtime node host metadata 生成 Access URL。
- Provisioning stage 字段和事件历史。
- 前端展示当前 provisioning stage，包括“正在重新调度”。
- 聚焦的后端和前端测试。

v1 不包含：

- Runtime node 管理 UI/API。
- Node host 字段的管理员 CRUD。
- 按 instance id 路由的公共 ingress / gateway。
- 跨宿主机 overlay network。
- AWD 单 service 自动漂移到其他 node。
- 超出“过滤 healthy/schedulable node + 从该 node pool 分配资源”之外的容量感知调度。

不做 runtime node 管理 UI/API 的原因：

- 第一版必须先稳定 runtime 正确性：node-scoped resources、access URL correctness、retry / quarantine 行为和 provisioning observability。
- 管理 UI/API 会引入独立 surface：权限、校验 UX、审计日志、前端表单、列表 / 详情页和运维工作流。
- v1 中 node hosts 通过 config / bootstrap / ops seed path 维护。这是有意取舍，不是遗漏。后续可在 runtime allocation model 稳定后单独增加 UI/API。

## 数据模型

### `runtime_nodes`

新增：

```sql
public_host varchar(255) NOT NULL DEFAULT ''
access_host varchar(255) NOT NULL DEFAULT ''
```

语义：

- `endpoint`：API/gateway 访问 runtime-agent 的控制面地址。
- `public_host`：学生访问该 node 发布端口时使用的 host。
- `access_host`：API/gateway/checker 内部访问该 node 发布端口时使用的 host。
- fallback 顺序：
  - public：node `public_host` -> global `container.public_host`
  - access：node `access_host` -> node `public_host` -> global `container.access_host` -> global `container.public_host`

### `runtime_port_pool`

```sql
CREATE TABLE runtime_port_pool (
  runtime_node_id bigint NOT NULL REFERENCES runtime_nodes(id) ON DELETE CASCADE,
  port integer NOT NULL,
  status varchar(16) NOT NULL DEFAULT 'available',
  instance_id bigint REFERENCES instances(id) ON DELETE SET NULL,
  reserved_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (runtime_node_id, port)
);
```

索引：

```sql
CREATE INDEX idx_runtime_port_pool_available
  ON runtime_port_pool(runtime_node_id, status, port);

CREATE INDEX idx_runtime_port_pool_instance
  ON runtime_port_pool(instance_id)
  WHERE instance_id IS NOT NULL;
```

状态：`available / reserved / bound / quarantined`。

### `runtime_subnet_pool`

```sql
CREATE TABLE runtime_subnet_pool (
  runtime_node_id bigint NOT NULL REFERENCES runtime_nodes(id) ON DELETE CASCADE,
  pool_kind varchar(32) NOT NULL,
  subnet text NOT NULL,
  status varchar(16) NOT NULL DEFAULT 'available',
  instance_id bigint REFERENCES instances(id) ON DELETE SET NULL,
  network_key varchar(128) NOT NULL DEFAULT '',
  reserved_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (runtime_node_id, subnet)
);
```

索引：

```sql
CREATE INDEX idx_runtime_subnet_pool_available
  ON runtime_subnet_pool(runtime_node_id, pool_kind, status, subnet);

CREATE UNIQUE INDEX uk_runtime_subnet_pool_instance_network
  ON runtime_subnet_pool(instance_id, network_key)
  WHERE instance_id IS NOT NULL AND status IN ('reserved', 'bound');
```

`pool_kind`: `single_container / topology`.

### `instances`

新增：

```sql
provisioning_stage varchar(64) NOT NULL DEFAULT ''
provisioning_attempt integer NOT NULL DEFAULT 0
last_provisioning_error text NOT NULL DEFAULT ''
```

`status` 仍然是粗粒度生命周期。`provisioning_stage` 是 `pending / creating / failed` 阶段的当前展示状态。

### `instance_provisioning_events`

```sql
CREATE TABLE instance_provisioning_events (
  id bigserial PRIMARY KEY,
  instance_id bigint NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
  attempt integer NOT NULL DEFAULT 0,
  stage varchar(64) NOT NULL,
  message varchar(255) NOT NULL DEFAULT '',
  severity varchar(16) NOT NULL DEFAULT 'info',
  runtime_node_id bigint REFERENCES runtime_nodes(id) ON DELETE SET NULL,
  detail jsonb NOT NULL DEFAULT '{}',
  created_at timestamptz NOT NULL DEFAULT now()
);
```

索引：

```sql
CREATE INDEX idx_instance_provisioning_events_instance
  ON instance_provisioning_events(instance_id, created_at DESC);

CREATE INDEX idx_instance_provisioning_events_stage
  ON instance_provisioning_events(stage, created_at DESC);
```

## 阶段契约

后端 stage 值：

- `queued`
- `selecting_node`
- `allocating_port`
- `allocating_network`
- `creating_network`
- `creating_container`
- `starting_container`
- `probing_readiness`
- `cleaning_previous`
- `rescheduling`
- `failed`

学生侧展示文案：

- `queued`: `排队中`
- `selecting_node`: `正在选择运行节点`
- `allocating_port`: `正在分配访问端口`
- `allocating_network`: `正在分配隔离网络`
- `creating_network`: `正在创建隔离网络`
- `creating_container`: `正在创建靶机容器`
- `starting_container`: `正在启动靶机`
- `probing_readiness`: `正在检查服务可用性`
- `cleaning_previous`: `正在清理上一次启动残留`
- `rescheduling`: `正在重新调度`
- `failed`: `启动失败`

响应字段：

```json
{
  "status": "creating",
  "provisioning_stage": "allocating_port",
  "provisioning_message": "正在分配访问端口",
  "provisioning_attempt": 1
}
```

## 文件地图

数据库：

- 修改：`code/backend/migrations/000001_init_schema.up.sql`
- 修改：`code/backend/internal/app/` 下 runtime 相关 migration contract tests

Container runtime：

- 修改：`code/backend/internal/module/container_runtime/entity/runtime_node.go`
- 新建：`code/backend/internal/module/container_runtime/entity/runtime_port_pool.go`
- 新建：`code/backend/internal/module/container_runtime/entity/runtime_subnet_pool.go`
- 修改：`code/backend/internal/module/container_runtime/contracts/runtime_node.go` 或最近的 runtime node contract 文件
- 修改：`code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
- 新建：`code/backend/internal/module/container_runtime/infrastructure/resource_pool_repository.go`
- 测试：`code/backend/internal/module/container_runtime/infrastructure/resource_pool_repository_test.go`
- 测试：`code/backend/internal/module/container_runtime/infrastructure/node_repository_test.go`

Instance：

- 修改：`code/backend/internal/module/instance/entity/instance.go`
- 新建：`code/backend/internal/module/instance/entity/provisioning_event.go`
- 修改：`code/backend/internal/module/instance/contracts/instance_output.go`
- 修改：`code/backend/internal/module/instance/infrastructure/repository.go`
- 测试：`code/backend/internal/module/instance/infrastructure/repository_test.go`

Practice：

- 修改：`code/backend/internal/module/practice/ports/ports.go`
- 修改：`code/backend/internal/module/practice/infrastructure/repository.go`
- 修改：`code/backend/internal/module/practice/application/commands/instance_start_service.go`
- 修改：`code/backend/internal/module/practice/application/commands/instance_provisioning.go`
- 修改：`code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- 新建：`code/backend/internal/module/practice/application/commands/provisioning_progress.go`
- 测试：`code/backend/internal/module/practice/application/commands/instance_start_service_test.go`
- 测试：`code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`

Composition 与 access URL：

- 修改：`code/backend/internal/module/container_runtime/contracts/access_url.go`
- 修改：`code/backend/internal/module/container_runtime/application/commands/provisioning_service.go`
- 修改：`code/backend/internal/app/composition/runtime_node_execution_router.go`
- 测试：`code/backend/internal/module/container_runtime/application/commands/provisioning_service_test.go`
- 测试：`code/backend/internal/app/composition/runtime_node_execution_router_test.go`

前端：

- 修改现有 instance / AWD workspace presentation code 中的 instance 状态文案映射。
- 测试：更新相关 presentation tests，覆盖 `provisioning_stage`。

文档：

- 实现后更新 `docs/architecture/backend/03-container-architecture.md`。
- 更新 `docs/operations/runtime-agent-deployment.md`，说明 `public_host/access_host`。

## 实施任务

### 任务 1：Runtime Node Host 元数据

**文件：**
- 修改：`code/backend/internal/module/container_runtime/entity/runtime_node.go`
- 修改：`code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
- 修改：runtime node contract/config bootstrap files
- 测试：`code/backend/internal/module/container_runtime/infrastructure/node_repository_test.go`

- [x] **步骤 1：添加 `public_host/access_host` 持久化失败测试**

断言 `EnsureDefaultNode` 会创建并更新这两个字段。

- [x] **步骤 2：运行失败测试**

运行：

```bash
cd code/backend
go test ./internal/module/container_runtime/infrastructure -run TestRuntimeNodeRepository -count=1
```

预期：FAIL。

- [x] **步骤 3：添加 entity、contract 和 repository 字段**

不要改变 `endpoint` 语义。

- [x] **步骤 4：运行聚焦测试**

运行：

```bash
cd code/backend
go test ./internal/module/container_runtime/infrastructure -run TestRuntimeNodeRepository -count=1
```

预期：PASS。

### 任务 2：资源池与 Provisioning Events Schema

**文件：**
- 修改：`code/backend/migrations/000001_init_schema.up.sql`
- 新建：`code/backend/internal/module/container_runtime/entity/runtime_port_pool.go`
- 新建：`code/backend/internal/module/container_runtime/entity/runtime_subnet_pool.go`
- 修改：`code/backend/internal/module/instance/entity/instance.go`
- 新建：`code/backend/internal/module/instance/entity/provisioning_event.go`
- 测试：`code/backend/internal/app/*migration*_test.go` 或最近的 baseline schema test

- [x] **步骤 1：添加 schema 失败断言**

断言 baseline 包含：

- `runtime_nodes.public_host`
- `runtime_nodes.access_host`
- `runtime_port_pool`
- `runtime_subnet_pool`
- `instances.provisioning_stage`
- `instances.provisioning_attempt`
- `instances.last_provisioning_error`
- `instance_provisioning_events`
- port/subnet pool 的 per-node primary keys

- [x] **步骤 2：运行失败 migration test**

运行：

```bash
cd code/backend
go test ./internal/app -run 'Test.*Runtime.*Schema|Test.*Migration' -count=1
```

预期：FAIL。

- [x] **步骤 3：添加 SQL schema 和 entities**

迁移期间保留现有 `port_allocations/network_allocations`。本切片不要删除旧表。

- [x] **步骤 4：运行 migration test**

运行：

```bash
cd code/backend
go test ./internal/app -run 'Test.*Runtime.*Schema|Test.*Migration' -count=1
```

预期：PASS。

### 任务 3：按 Runtime Node 分片的资源池 Repository

**文件：**
- 新建：`code/backend/internal/module/container_runtime/infrastructure/resource_pool_repository.go`
- 测试：`code/backend/internal/module/container_runtime/infrastructure/resource_pool_repository_test.go`

- [x] **步骤 1：添加不同 node 复用同一端口的失败测试**

为 node A 和 node B 都 seed 端口 `30000`。两个 reservation 必须能独立成功。

- [x] **步骤 2：添加同一 node 并发分配的失败测试**

同一 node 上的并发 reservation 不能返回同一个端口。

- [x] **步骤 3：添加不同 node 复用同一网段的失败测试**

node A 和 node B 上的同一 subnet 必须能独立 reserve。

- [x] **步骤 4：实现 repository 方法**

必需方法：

- `EnsurePoolsForNode(ctx, nodeID, cfg) error`
- `ReserveAvailablePortForNode(ctx, nodeID, instanceID int64) (int, error)`
- `ReserveAvailableSubnetForNode(ctx, nodeID int64, poolKind string, instanceID int64, networkKey string) (string, error)`
- `BindResourcesForInstance(ctx, instanceID int64) error`
- `ReleaseResourcesForInstance(ctx, instanceID int64) error`
- `QuarantinePort(ctx, nodeID int64, port int, reason string) error`
- `QuarantineSubnet(ctx, nodeID int64, subnet string, reason string) error`

- [x] **步骤 5：运行资源池测试**

运行：

```bash
cd code/backend
go test ./internal/module/container_runtime/infrastructure -run TestResourcePool -count=1
```

预期：PASS。

### 任务 4：从配置生成资源池

**文件：**
- 修改：`code/backend/internal/module/container_runtime/infrastructure/resource_pool_repository.go`
- 按需修改：`code/backend/internal/app/composition/container_runtime_module.go` 中的 runtime module bootstrap
- 测试：`code/backend/internal/module/container_runtime/infrastructure/resource_pool_repository_test.go`
- 测试：`code/backend/internal/app/composition/runtime_module_test.go`

- [x] **步骤 1：添加幂等性失败测试**

连续调用两次 `EnsurePoolsForNode` 不能重复插入 rows，也不能覆盖 `reserved/bound` rows。

- [x] **步骤 2：添加配置生成数量失败测试**

对于 `30000-30003`，预期生成 3 条 port rows。对于 topology `10.10.0.0/16 + /24`，预期生成 256 条 subnet rows。

- [x] **步骤 3：实现资源池 seed**

使用现有配置：

- `container.port_range_start`
- `container.port_range_end`
- `container.network.single_container_subnet_base`
- `container.network.single_container_subnet_mask`
- `container.network.topology_subnet_base`
- `container.network.topology_subnet_mask`

- [x] **步骤 4：运行测试**

运行：

```bash
cd code/backend
go test ./internal/module/container_runtime/infrastructure ./internal/app/composition -run 'TestResourcePool|TestBuildContainerRuntimeModule|TestRuntimeNode' -count=1
```

预期：PASS。

### 任务 5：Provisioning Stage Repository 与事件

**文件：**
- 修改：`code/backend/internal/module/instance/contracts/instance_output.go`
- 修改：`code/backend/internal/module/instance/infrastructure/repository.go`
- 新建：`code/backend/internal/module/practice/application/commands/provisioning_progress.go`
- 测试：`code/backend/internal/module/instance/infrastructure/repository_test.go`

- [ ] **步骤 1：添加 stage 更新与 event 追加原子性的失败测试**

Repository 必须在同一个 DB transaction 中更新 `instances.provisioning_stage` 并追加 `instance_provisioning_events`。

- [ ] **步骤 2：添加响应字段失败测试**

`InstanceResp` 和 `InstanceInfo` 暴露：

- `provisioning_stage`
- `provisioning_message`
- `provisioning_attempt`

- [ ] **步骤 3：实现 repository 和 mapping**

不要把内部原始错误直接暴露到学生侧 `provisioning_message`。

- [ ] **步骤 4：运行聚焦测试**

运行：

```bash
cd code/backend
go test ./internal/module/instance/infrastructure ./internal/module/instance/contracts -run 'Test.*Provisioning|Test.*Instance' -count=1
```

预期：PASS。

### 任务 6：Practice 启动流程中的 Node-Scoped Allocation

**文件：**
- 修改：`code/backend/internal/module/practice/ports/ports.go`
- 修改：`code/backend/internal/module/practice/infrastructure/repository.go`
- 修改：`code/backend/internal/module/practice/application/commands/instance_start_service.go`
- 测试：`code/backend/internal/module/practice/application/commands/instance_start_service_test.go`

- [ ] **步骤 1：添加按选中 node reserve 端口的失败测试**

当 selector 返回 node B 时，instance 必须保存 node B，并从 node B 的资源池 reserve 端口。

- [ ] **步骤 2：添加两个 node 复用同一端口的失败测试**

不同 node 上的两个实例可以同时获得 `30000`；它们生成的 access hosts 必须不同。

- [ ] **步骤 3：更新 practice ports 和 repository adapter**

通过窄接口暴露 node-scoped reservation。不要把 GORM 或 pool table 细节泄漏到 application service。

- [ ] **步骤 4：更新启动流程**

顺序必须是：

```text
select healthy node -> update provisioning stage -> reserve node port -> create instance -> bind reservation
```

- [ ] **步骤 5：运行聚焦测试**

运行：

```bash
cd code/backend
go test ./internal/module/practice/application/commands -run 'Test.*RuntimeNode.*Port|Test.*HostPort|Test.*Start.*Instance' -count=1
```

预期：PASS。

### 任务 7：Node-Aware Access URL 生成

**文件：**
- 修改：`code/backend/internal/module/container_runtime/contracts/access_url.go`
- 修改：`code/backend/internal/module/container_runtime/application/commands/provisioning_service.go`
- 修改：`code/backend/internal/app/composition/runtime_node_execution_router.go`
- 测试：`code/backend/internal/module/container_runtime/application/commands/provisioning_service_test.go`
- 测试：`code/backend/internal/app/composition/runtime_node_execution_router_test.go`

- [ ] **步骤 1：添加 node public host 失败测试**

在 node B 上 provision，`public_host=node-b.ctf.local` 且端口为 `30000` 时，学生侧 URL 必须是 `http://node-b.ctf.local:30000`。

- [ ] **步骤 2：添加 node access host 失败测试**

内部 probe/proxy 使用 `access_host`，学生响应使用 `public_host`。

- [ ] **步骤 3：实现 resolver**

只有 node 字段为空时才 fallback 到全局配置。

- [ ] **步骤 4：通过 router 传递 node host metadata**

选中 node 的 host metadata 必须传到 provisioning，且不能改变 `runtime_nodes.endpoint` 语义。

- [ ] **步骤 5：运行聚焦测试**

运行：

```bash
cd code/backend
go test ./internal/module/container_runtime/application/commands ./internal/app/composition -run 'Test.*AccessHost|Test.*PublicHost|TestRuntimeNodeExecutionRouter' -count=1
```

预期：PASS。

### 任务 8：Runtime Conflict、Cleanup、Rescheduling 与 Quarantine

**文件：**
- 修改：`code/backend/internal/module/container_runtime/application/commands/provisioning_service.go`
- 修改：`code/backend/internal/module/practice/application/commands/instance_provisioning.go`
- 修改：`code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- 测试：`code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
- 测试：`code/backend/internal/module/container_runtime/application/commands/provisioning_service_test.go`

- [ ] **步骤 1：添加可重试 Docker failure 的失败测试**

普通实例遇到可重试 provisioning failure 后：

- 旧 resources 被 cleaned 或 quarantined
- 当前 stage 变为 `rescheduling`
- instance 保持 `creating`
- 写入一条 provisioning event

- [ ] **步骤 2：添加 subnet conflict quarantine 失败测试**

Docker subnet overlap 只 quarantine `(runtime_node_id, subnet)`，并在同一 node pool 内重试。

- [ ] **步骤 3：添加 host port conflict 失败测试**

Docker host port conflict 不能在同一次 attempt 里复用同一个端口。

- [ ] **步骤 4：实现 failure classification**

使用现有错误：

- `ErrRuntimeNetworkSubnetConflict`
- `ErrPublishedHostPortConflict`
- `ErrRuntimeNodeUnavailable`

普通实例可以在 retry budget 内 reschedule。AWD service instances 不能静默移动到另一个 node。

- [ ] **步骤 5：运行聚焦测试**

运行：

```bash
cd code/backend
go test ./internal/module/practice/application/commands ./internal/module/container_runtime/application/commands -run 'Test.*Conflict|Test.*Quarantine|Test.*Rescheduling|Test.*Provisioning.*Failure' -count=1
```

预期：PASS。

### 任务 9：前端 Provisioning Progress 展示

**文件：**
- 修改相关 frontend instance presentation model/components。
- 如果 AWD workspace 会展示实例启动 label，则同步修改 AWD workspace presentation。
- 测试：`code/frontend/src/**/__tests__` 或 `*.test.ts` 下相关 Vitest 文件。

- [ ] **步骤 1：添加 `rescheduling` presentation 失败测试**

预期 label：`正在重新调度`。

- [ ] **步骤 2：添加 allocation stages presentation 失败测试**

至少覆盖：

- `allocating_port`
- `allocating_network`
- `creating_container`

- [ ] **步骤 3：实现 stage label mapping**

存在 backend `provisioning_message` 时优先使用它。为增强韧性，fallback 到前端已知 stage labels。

- [ ] **步骤 4：运行前端测试**

运行：

```bash
cd code/frontend
npm test -- --run
```

如果该命令对本地迭代过宽，先运行被修改的测试文件；完成前再运行项目要求的 frontend guard。

预期：PASS。

### 任务 10：文档与运维说明

**文件：**
- 修改：`docs/architecture/backend/03-container-architecture.md`
- 修改：`docs/operations/runtime-agent-deployment.md`
- 可选修改：如果 dev compose host 行为变化，修改 `README.md`

- [ ] **步骤 1：更新架构事实**

记录：

- per-node `public_host/access_host`
- node-scoped port/subnet pools
- provisioning events
- AWD no silent per-service drift

- [ ] **步骤 2：更新运维说明**

记录 v1 没有 management UI/API 时如何配置 node public/access host。

- [ ] **步骤 3：运行文档检查**

运行：

```bash
python3 scripts/check-docs-consistency.py
```

预期：PASS。

## 验证

聚焦后端：

```bash
cd code/backend
go test ./internal/module/container_runtime/infrastructure -run 'TestRuntimeNodeRepository|TestResourcePool' -count=1
go test ./internal/module/instance/infrastructure -run 'Test.*Provisioning.*Event|Test.*Provisioning.*Stage' -count=1
go test ./internal/module/practice/application/commands -run 'Test.*RuntimeNode.*Port|Test.*HostPort|Test.*Rescheduling|Test.*Provisioning' -count=1
go test ./internal/module/container_runtime/application/commands ./internal/app/composition -run 'Test.*AccessHost|Test.*PublicHost|TestRuntimeNodeExecutionRouter|Test.*Conflict|Test.*Quarantine' -count=1
go test ./internal/app -run 'Test.*Runtime.*Schema|Test.*Migration' -count=1
```

较宽后端：

```bash
cd code/backend
go test ./internal/module/container_runtime/... ./internal/module/instance/... ./internal/module/practice/... ./internal/app/composition -count=1
```

前端：

```bash
cd code/frontend
npm test -- --run
```

项目检查：

```bash
bash scripts/check-frontend-test-guard.sh
python3 scripts/check-docs-consistency.py
```

## 风险

- **Access URL 歧义：** 每 node 复用端口只有在存在 node-specific public hosts 或等价 routing 时才安全。否则应保持端口全局唯一，或拒绝该配置。
- **写放大：** Provisioning events 会增加写入。只记录有意义的生命周期边界，不记录每一条 low-level log line。
- **残留 Docker 状态：** Pool state 可能和真实 Docker host state 不一致。必须保留 runtime conflict detection 和 quarantine。
- **现有 active instances：** 不要静默重写 active `instances.access_url`。
- **AWD placement：** AWD contest services 不能逐个 service 静默漂移到其他 nodes。
- **v1 没有 management UI/API：** 这是有意取舍。直到后续 dedicated management task 落地前，operator 通过 config/bootstrap/ops seed path 维护 node host fields。

## 架构适配评估

- 边界明确：`container_runtime` 拥有 node resources；`instance` 拥有持久化 lifecycle/progress；`practice` 拥有 provisioning orchestration。
- Access URL 正确性不延期处理；per-node data-plane host metadata 属于本次实现的一部分。
- Resource pool allocation 和 provisioning progress 一起实现，因为用户可见的 rescheduling 依赖 allocation/retry 语义。
- Runtime node management UI/API 被有意识地排除在范围外，并已记录原因。
