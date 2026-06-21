<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# AWD 运行节点身份与 Placement Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to execute this plan task-by-task in this task worktree. Steps use checkbox (`- [ ]`) syntax for tracking. Production code changes require TDD: write each failing test, run it red, then implement the smallest code to pass.

**Goal:** 把平台运行节点身份从泛化的 `node_id` 收敛为明确的 `runtime_node_id`，并让一场 AWD 比赛在非 emergency recreate 的情况下始终绑定到同一个平台 runtime node。

**Architecture:** `runtime_nodes.id` 继续作为平台控制面的 runtime node 主键；`instances.runtime_node_id`、checker metadata、AWD workspace / service routing 使用这个平台主键路由到对应 runtime-agent / Docker daemon。AWD 增加 contest 级轻量 placement 表，首次启动时从健康可调度 runtime node 中选择一个并持久化，后续防守重启、desired reconcile 和 pending scheduler 只复用该绑定；绑定 node 不可用时返回 `ErrRuntimeNodeUnavailable` 并保持等待，不回退到默认 selector。

**Tech Stack:** Go, GORM, PostgreSQL migrations, existing `container_runtime` runtime node repository / selector, `practice` instance lifecycle services, `contest` AWD runtime records, package-level Go tests.

---

## Task Metadata

- Task Slug: `2026-06-21-awd-runtime-node-identity-and-placement`
- Created At: `2026-06-21T19:10:11+08:00`
- Draft Branch: `multi-instance`
- Implementation Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-21-awd-runtime-node-identity-and-placement`
- Implementation Branch: `task/2026-06-21-awd-runtime-node-identity-and-placement`
- Plan Type: `formal implementation plan`
- Related TODO: `docs/todos/2026-06-21-awd-runtime-node-identity-and-placement.md`

## Plan Status

- Status: `ready-for-implementation`
- Coding may start only after:
  - [x] 运行 `bash scripts/check-task-intake.sh` 确认当前 task 入口状态。
  - [x] 运行 `bash scripts/start-implementation.sh 2026-06-21-awd-runtime-node-identity-and-placement` 绑定 task slug、worktree / branch 和 startup gate。
  - [x] 由执行者复读本 plan，并确认本轮只做已确认范围。
  - [x] 进入代码前完成一次独立 plan review 或人工 review。

启动记录：

- `2026-06-21`：执行者已读取本 plan、项目规则、测试分层说明、reuse policy、相关架构文档和 runtime node schedulable 反馈；本轮只实现本 plan 已列范围，不包含 emergency recreate、reservation / capacity preflight 或 Docker / Swarm 内部 node identity 建模。
- `2026-06-21`：用户已明确要求开始实现；执行者完成 architecture-fit 复核，未发现会改变实现方向的 blocker。当前环境没有独立 subagent 工具，完成实现后仍需按 `code-workflow` 执行独立 review gate 或明确标注该 gate 未满足。

## Intake Analysis Gate

- `2026-06-21`：已完成 task intake，确认本任务属于非琐碎后端 / migration / 架构文档改动，并已绑定 task slug、implementation plan 和当前 worktree。
- `2026-06-21`：实现前已完成 architecture-fit 复核；确认本轮只处理 runtime node identity 命名收敛与 AWD contest runtime placement，不包含 emergency recreate、reservation / capacity preflight 或 Docker / Swarm 内部 node identity 建模。
- `2026-06-21`：复用与 owner 决策已收口在本 plan 的 `## 复用与 Owner 决策`，practice 通过 port 消费 contest placement 与 container_runtime health lookup，具体接线放在 `app/composition`。

## 已确认术语

- `runtime_node` / `runtime_node_id`
  - 含义：CTF 平台登记的执行节点，持久化在 `runtime_nodes` 表；通常对应一台运行 Docker daemon / runtime-agent 的宿主。
  - 主键：`runtime_nodes.id`。
  - 外键字段：`instances.runtime_node_id`，以及需要执行路由的 checker / AWD metadata 中的 `runtime_node_id`。
- `runtime_node_name`
  - 含义：部署和运维可读的稳定逻辑名，对应 `runtime_nodes.name`。
  - 不替代数据库主键；用于配置、日志和排障。
- `container_id` / `network_id`
  - 含义：Docker / runtime engine 对象 id。
  - 不具备跨 runtime node 的全局唯一语义；执行定位必须按 `(runtime_node_id, container_id)` 或 `(runtime_node_id, network_id)` 理解。
- `docker_node_id` / `engine_node_id`
  - 含义：Docker daemon / Swarm / engine 内部自己的节点身份。
  - 本计划不新增该字段；未来如果需要建模，必须使用 `docker_node_id` 或 `engine_node_id`，不得再叫 `node_id`，也不得和平台 `runtime_node_id` 混用。
- `contest_runtime_placement`
  - 含义：一场 AWD 比赛绑定到哪个平台 runtime node 的轻量记录。
  - 约束：每场 AWD contest 只有一个 active placement；多场 AWD 比赛可以各自绑定不同 runtime node，也可以在没有 reservation 的情况下被选到同一 runtime node。

## Objective And Non-Goals

- Objective:
  - 把 `instances.node_id` 迁移为 `instances.runtime_node_id`，并同步 Go entity、repository、DTO、测试、日志和架构文档命名。
  - 为 `instances(runtime_node_id, container_id)` 和 `instances(runtime_node_id, network_id)` 增加辅助索引。
  - 新增 AWD contest runtime placement 的轻量持久化。
  - 调整 AWD runtime node 选择：一场 AWD 比赛首次启动时从多个健康可调度 node 中选一个；后续所有该 contest 的 AWD service 启动、重启和 desired reconcile 只使用该 node。
  - 当绑定 runtime node 不可用时，AWD instance 保持 pending / backoff，不自动漂移到其他 runtime node。
- Non-Goals:
  - 不实现管理员 emergency recreate API / UI / runbook。
  - 不实现完整 runtime reservation、容量预占、容量可视化或跨 contest 容量规划。
  - 不实现 Docker / Swarm 内部 `docker_node_id` / `engine_node_id` 建模。
  - 不实现容器级跨 node 放置；同一实例 / 拓扑 / AWD contest 仍以单 runtime node 为执行单元。
  - 不为旧字段名 `node_id` 增加长期兼容层；开发阶段按项目规则迁移到新命名。

## Problem Statement

- 当前 `instances.node_id` 实际引用 `runtime_nodes.id`，但命名无法区分平台 runtime node 和 Docker / engine 内部 node identity。
- 当前执行面有不少只凭 `container_id` 反查 node 的路径；在多 Docker daemon 场景下，`container_id` / `network_id` 必须带上所属 `runtime_node_id` 才是完整执行身份。
- 当前 AWD 防守重启在 scheduler enabled 时会把实例重置为 `pending`；`instance_provisioning_scheduler.go` 后续重新调用 `SelectRuntimeNode()`。
- 当前 `instance_practice_runtime_node_selector_adapter.go` 明确忽略 `InstanceScope`，所以 AWD scope 不会复用 contest 绑定 node，而是走默认 selector。
- 当前 runtime node offline failover 会清空旧 `node_id` 并触发 desired reconcile；如果 selector 后续选择了其他健康 node，AWD runtime 会发生静默跨 node 漂移。
- AWD 的 bridge alias、防守工作区和 checker token 场景假设同一场比赛的 service runtime 在同一 Docker 宿主上；静默漂移会破坏该假设，也会让赛中排障不可预测。

## Inputs

- Source docs:
  - `AGENTS.md`
  - `docs/文档规范.md`
  - `docs/plan/README.md`
  - `docs/plan/impl-plan/README.md`
  - `docs/todos/2026-06-21-awd-runtime-node-identity-and-placement.md`
  - `code/backend/tests/README.md`
  - `harness/policies/reuse-first.yaml`
- Related architecture docs:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
- Related code:
  - `code/backend/migrations/000001_init_schema.up.sql`
  - `code/backend/internal/app/runtime_node_migration_test.go`
  - `code/backend/internal/module/instance/entity/instance.go`
  - `code/backend/internal/module/instance/infrastructure/repository.go`
  - `code/backend/internal/module/instance/infrastructure/container_inventory_repository.go`
  - `code/backend/internal/module/contest/infrastructure/awd_container_inventory_repository.go`
  - `code/backend/internal/module/contest/infrastructure/awd_service_instance_repository.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/runtime_state_providers.go`
  - `code/backend/internal/app/composition/instance_practice_runtime_node_selector_adapter.go`
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - `code/backend/internal/module/practice/ports/ports.go`
  - `code/backend/internal/module/practice/application/commands/instance_start_service.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
  - `code/backend/internal/app/composition/runtime_node_failover.go`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 涉及数据库 migration、后端实体命名、跨模块 ports / adapters、AWD 生命周期和运行时调度语义。
  - 错误实现会导致赛中实例被调度到错误 Docker 宿主，或使旧容器清理 / checker / 文件写入路由到错误 node。
  - 必须 TDD、按 slice 提交，并在完成后更新架构事实源。

## Files

- Create:
  - `code/backend/migrations/000020_rename_instance_runtime_node_id_and_add_awd_runtime_placements.up.sql`
  - `code/backend/migrations/000020_rename_instance_runtime_node_id_and_add_awd_runtime_placements.down.sql`
  - `code/backend/internal/module/contest/entity/contest_runtime_placement.go`
  - `code/backend/internal/module/contest/infrastructure/contest_runtime_placement_repository.go`
  - `code/backend/internal/module/contest/infrastructure/contest_runtime_placement_repository_test.go`
- Modify:
  - `code/backend/migrations/000001_init_schema.up.sql`
  - `code/backend/internal/app/runtime_node_migration_test.go`
  - `code/backend/internal/module/instance/entity/instance.go`
  - `code/backend/internal/module/instance/infrastructure/repository.go`
  - `code/backend/internal/module/instance/infrastructure/repository_test.go`
  - `code/backend/internal/module/instance/infrastructure/container_inventory_repository.go`
  - `code/backend/internal/module/instance/infrastructure/acl_migration_state_repository.go`
  - `code/backend/internal/module/contest/ports/awd.go`
  - `code/backend/internal/module/contest/ports/checker_runner.go`
  - `code/backend/internal/module/contest/infrastructure/awd_container_inventory_repository.go`
  - `code/backend/internal/module/contest/infrastructure/awd_service_instance_repository.go`
  - `code/backend/internal/module/contest/application/jobs/awd_*checker*.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/runtime_state_providers.go`
  - `code/backend/internal/app/composition/runtime_cleanup_target_adapter.go`
  - `code/backend/internal/app/composition/instance_practice_runtime_node_selector_adapter.go`
  - `code/backend/internal/app/composition/runtime_node_failover.go`
  - `code/backend/internal/module/practice/ports/ports.go`
  - `code/backend/internal/module/practice/application/commands/instance_start_service.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
  - `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
  - `code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go`
  - `code/backend/internal/module/practice/application/commands/*_test.go`
  - `code/backend/internal/testutil/systemapp/practice_flow.go`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/todos/2026-06-21-awd-runtime-node-identity-and-placement.md`
- Review:
  - `code/backend/internal/module/container_runtime/infrastructure/node_repository.go`
  - `code/backend/internal/module/container_runtime/ports/node.go`
  - `code/backend/internal/module/container_runtime/contracts/runtime_node.go`
  - `code/backend/internal/app/router_composition_typed_deps_test.go`
  - `docs/architecture/backend/platform-overall-architecture.drawio`
  - `docs/architecture/backend/platform-overall-architecture.svg`
- Test:
  - `code/backend/internal/app/runtime_node_migration_test.go`
  - `code/backend/internal/module/contest/infrastructure/contest_runtime_placement_repository_test.go`
  - `code/backend/internal/module/instance/infrastructure/repository_test.go`
  - `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`
  - `code/backend/internal/module/practice/application/commands/instance_restart_awd_service_test.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - `code/backend/internal/app/composition/runtime_state_providers_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - Runtime node registry and health selector: `RuntimeNodeRepository`, `FindHealthyByID`, `ListSchedulableHealthyNodes`, `NewDefaultRuntimeNodeSelector`。
  - Practice runtime selection seam: `practiceports.RuntimeNodeSelector` and `practiceRuntimeNodeSelectorAdapter`。
  - AWD lifecycle owner: `RestartContestAWDService`, `restartOrStartScopedAWDService`, `ReconcileDesiredAWDInstances`, `processPendingInstance`。
  - Runtime routing owner: `runtimeNodeExecutionRouter`, `FindRuntimeNodeIDByContainerID`, AWD workspace / checker metadata。
  - Migration baseline guardrail: `runtime_node_migration_test.go`。
- Reuse / extend / split / create-new decision:
  - `extend_existing`: 复用 `RuntimeNodeRepository.FindHealthyByID()` 验证 bound node 可用性。
  - `extend_existing`: 扩展 `practiceports.RuntimeNodeSelector.SelectRuntimeNode(ctx, scope)` 的 adapter 逻辑，让 AWD scope 进入 placement 分支。
  - `create_new`: 新建 `contest_runtime_placements` 表，因为 contest 级 placement 是比赛运行策略，不属于单个 instance。
  - `create_new`: 新建 contest-owned repository，practice 只通过 port 消费 placement，不直接依赖 contest infrastructure。
  - `rename_existing`: 将平台执行身份字段从 `NodeID` / `node_id` 收敛为 `RuntimeNodeID` / `runtime_node_id`；`RuntimeNode` entity 自身的 `ID` 保持不变。
- Owner boundary:
  - `container_runtime` owns runtime node registry、health status、default node selector and explicit healthy-by-id lookup。
  - `contest` owns contest-level AWD runtime placement persistence。
  - `practice` owns instance lifecycle and asks a port for runtime placement / selection; it does not query contest tables directly.
  - `instance` owns per-instance runtime attempt identity, including `runtime_node_id`, `container_id`, `network_id`, cleanup and requeue state.
  - `app/composition` wires contest placement repository + container runtime selector into the practice port.
- Why this is the narrowest safe surface:
  - AWD stickiness needs one contest-level record; it does not require full reservation or capacity scoring.
  - Explicit healthy-by-id lookup already exists, so bound-node enforcement does not require a second selector implementation inside practice.
  - Composite auxiliary indexes make existing reverse lookups safer without replacing business primary keys.

## Data Model

### `instances.runtime_node_id`

Migration target:

```sql
ALTER TABLE public.instances DROP CONSTRAINT IF EXISTS instances_node_id_fkey;
ALTER TABLE public.instances RENAME COLUMN node_id TO runtime_node_id;
ALTER INDEX IF EXISTS public.idx_instances_node_id RENAME TO idx_instances_runtime_node_id;
ALTER TABLE public.instances
    ADD CONSTRAINT instances_runtime_node_id_fkey
    FOREIGN KEY (runtime_node_id) REFERENCES public.runtime_nodes(id);

CREATE INDEX idx_instances_runtime_node_container_id
    ON public.instances USING btree (runtime_node_id, container_id)
    WHERE runtime_node_id IS NOT NULL AND container_id <> '';

CREATE INDEX idx_instances_runtime_node_network_id
    ON public.instances USING btree (runtime_node_id, network_id)
    WHERE runtime_node_id IS NOT NULL AND network_id <> '';
```

Go entity target:

```go
type Instance struct {
    RuntimeNodeID *int64 `gorm:"column:runtime_node_id;index"`
    ContainerID   string `gorm:"size:64;not null"`
    NetworkID     string `gorm:"size:64"`
}
```

### `contest_runtime_placements`

Migration target:

```sql
CREATE TABLE public.contest_runtime_placements (
    id bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    contest_id bigint NOT NULL,
    runtime_node_id bigint NOT NULL,
    status varchar(16) NOT NULL DEFAULT 'active',
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    released_at timestamp with time zone,
    CONSTRAINT contest_runtime_placements_contest_id_fkey
        FOREIGN KEY (contest_id) REFERENCES public.contests(id),
    CONSTRAINT contest_runtime_placements_runtime_node_id_fkey
        FOREIGN KEY (runtime_node_id) REFERENCES public.runtime_nodes(id),
    CONSTRAINT contest_runtime_placements_status_check
        CHECK (status IN ('active', 'released'))
);

CREATE UNIQUE INDEX idx_contest_runtime_placements_active_contest
    ON public.contest_runtime_placements (contest_id)
    WHERE status = 'active';

CREATE INDEX idx_contest_runtime_placements_runtime_node_status
    ON public.contest_runtime_placements (runtime_node_id, status);
```

Go entity target:

```go
type ContestRuntimePlacement struct {
    ID            int64      `gorm:"primaryKey"`
    ContestID     int64      `gorm:"column:contest_id;not null;index"`
    RuntimeNodeID int64      `gorm:"column:runtime_node_id;not null;index"`
    Status        string     `gorm:"size:16;not null;default:'active'"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
    ReleasedAt    *time.Time `gorm:"column:released_at"`
}
```

## Target Runtime Selection Behavior

- 非 AWD instance:
  - 继续使用默认 runtime node selector。
  - 如果默认 selector 找不到健康可调度 node，返回 `ErrRuntimeNodeUnavailable`，pending instance 等待下一轮。
- AWD instance 且 contest 没有 active placement:
  - 先调用默认 selector，从多个健康可调度 runtime node 中选一个。
  - 用唯一 active contest index 创建 placement。
  - 若并发创建冲突，重新读取 active placement，并验证该 bound node 是否健康。
- AWD instance 且 contest 已有 active placement:
  - 只按 placement 的 `runtime_node_id` 查询健康 node。
  - 健康则返回该 binding。
  - 不健康、offline、心跳过期、node 不存在时返回 `ErrRuntimeNodeUnavailable`。
  - 不 fallback 到 `SelectDefaultNode()`，不把同一 contest 静默漂移到其他 runtime node。
- Defense restart:
  - 清理旧容器仍按旧 `runtime_node_id + container_id` 路由。
  - scheduler enabled 时实例可以回到 `pending`，但后续 bind 只能绑定 contest placement 的 runtime node。
  - 除非管理员未来显式 emergency recreate，否则比赛中的 AWD runtime node 不改变。
- Desired reconcile / runtime node offline:
  - node offline 后 existing runtime attempt 可以清空 `container_id` / `network_id` / runtime details 并回到 `pending`。
  - contest placement 保持 active。
  - 如果 bound node 仍不可用，desired reconcile 记录失败 / backoff 或保持 pending，不自动改绑。

## Execution Slices

### Slice 1: 数据库命名迁移和辅助索引

- Goal:
  - 将 `instances.node_id` 改为 `instances.runtime_node_id`。
  - 增加 `(runtime_node_id, container_id)` 和 `(runtime_node_id, network_id)` 辅助索引。
  - 新建 `contest_runtime_placements` 表。
- Dependencies:
  - None.
- Files:
  - Create: `code/backend/migrations/000020_rename_instance_runtime_node_id_and_add_awd_runtime_placements.up.sql`
  - Create: `code/backend/migrations/000020_rename_instance_runtime_node_id_and_add_awd_runtime_placements.down.sql`
  - Modify: `code/backend/migrations/000001_init_schema.up.sql`
  - Modify: `code/backend/internal/app/runtime_node_migration_test.go`
- Steps:
  - [x] 写失败测试，断言 baseline schema 不再包含 `instances.node_id`，而是包含 `runtime_node_id`、`idx_instances_runtime_node_id`、`instances_runtime_node_id_fkey` 和两个 composite index。
  - [x] 写失败测试，断言 migration 列表包含 `000020`，并检查 up/down 中 `contest_runtime_placements` 的 active 唯一索引和外键。
  - [x] 运行：`cd code/backend && go test ./internal/app -run 'TestRuntimeNodeMigration|TestInstanceRuntimeNodeIDMigration|TestAWDRuntimePlacementMigration' -count=1`
  - [x] 确认 RED：测试因为 schema / migration 仍使用 `node_id` 或缺少 placement 表而失败。
  - [x] 新增 `000020` up/down migration，按 Data Model 中的 SQL 完成 rename、外键、索引和 placement 表。
  - [x] 更新 `000001_init_schema.up.sql` baseline，使全新库直接创建 `runtime_node_id` 和 placement 表。
  - [x] 重新运行 focused migration 测试并确认 GREEN。
  - [ ] 运行：`cd code/backend && go test ./internal/app -count=1`
  - [ ] 按 `committing-changes` skill 和仓库 commit policy 提交本 slice。
- Validation:
  - `cd code/backend && go test ./internal/app -run 'TestRuntimeNodeMigration|TestInstanceRuntimeNodeIDMigration|TestAWDRuntimePlacementMigration' -count=1`
  - `cd code/backend && go test ./internal/app -count=1`
- Validation evidence:
  - `cd code/backend && go test ./internal/app -run 'TestRuntimeNodeContractInBaseline|TestRuntimeNodeMigration|TestInstanceRuntimeNodeIDMigration|TestAWDRuntimePlacementMigration' -count=1`：PASS。
  - `cd code/backend && go test ./internal/app -count=1`：FAIL；失败集中在既有 full-router / typed-deps / 缺表测试，例如 `contest_realtime_outbox`、`platform_event_outbox` 缺表和 typed deps marker，不属于本 slice migration diff。该命令暂不作为 Slice 1 可归因 GREEN。
- Review focus:
  - Down migration 能完整恢复 `node_id`、旧 index 和旧 FK。
  - Partial index 条件不遗漏空字符串 `container_id` / `network_id`。
  - `contest_runtime_placements` 没有引入 reservation / capacity 字段。
- Done criteria:
  - 数据库层不再把平台 runtime node 叫作 `node_id`。
  - 复合执行身份查询有明确辅助索引。
  - AWD contest placement 有最小持久化表和唯一 active 约束。

### Slice 2: Instance runtime identity 代码命名收敛

- Goal:
  - 把 instance 侧平台执行身份从 `NodeID` 收敛为 `RuntimeNodeID`。
  - 所有 SQL / GORM column 引用使用 `runtime_node_id`。
- Dependencies:
  - Slice 1 migration.
- Files:
  - Modify: `code/backend/internal/module/instance/entity/instance.go`
  - Modify: `code/backend/internal/module/instance/infrastructure/repository.go`
  - Modify: `code/backend/internal/module/instance/infrastructure/container_inventory_repository.go`
  - Modify: `code/backend/internal/module/instance/infrastructure/acl_migration_state_repository.go`
  - Modify: `code/backend/internal/module/instance/infrastructure/repository_test.go`
  - Modify: `code/backend/internal/testutil/systemapp/practice_flow.go`
  - Modify: dependent compile errors surfaced by `go test`.
- Steps:
  - [x] 写失败 repository 测试，断言 `BindRuntimeNode` 写入 `runtime_node_id`，`RequeueLostRuntime` / `RequeueLostRuntimesByNode` 清理的是 `runtime_node_id`。
  - [x] 写失败 inventory 测试，断言按 `runtime_node_id + container_id` 查找运行节点。
  - [x] 运行：`cd code/backend && go test ./internal/module/instance/infrastructure -run 'Test.*RuntimeNodeID|TestRequeueLostRuntimesByNode|TestContainerInventory' -count=1`
  - [x] 确认 RED：测试或编译因 `NodeID` / `node_id` 仍存在而失败。
  - [x] 将 `instanceentity.Instance.NodeID` 重命名为 `RuntimeNodeID`，GORM tag 改为 `column:runtime_node_id;index`。
  - [x] 更新 instance repository 中所有 `node_id` map key、where 条件、select 列和返回字段为 `runtime_node_id`。
  - [x] 更新 testutil 和 instance infrastructure tests 中的字段名。
  - [x] 运行：`cd code/backend && go test ./internal/module/instance/... -count=1`
  - [x] 修复仅由重命名导致的编译错误，直到 instance module GREEN。
  - [ ] 按 `committing-changes` skill 和仓库 commit policy 提交本 slice。
- Validation:
  - `cd code/backend && go test ./internal/module/instance/... -count=1`
- Validation evidence:
  - RED：`cd code/backend && go test ./internal/module/instance/infrastructure -run 'TestRepositoryUsesRuntimeNodeIDColumnForBindingAndRequeue|TestContainerInventoryRepositoryReturnsRuntimeNodeID|Test.*RuntimeNodeID|TestRequeueLostRuntimesByNode|TestContainerInventory' -count=1` 因 `Instance` 缺少 `RuntimeNodeID` 编译失败。
  - GREEN：同 focused command 通过。
  - GREEN：`cd code/backend && go test ./internal/module/instance/... -count=1` 通过。
- Review focus:
  - `runtime_node_id` 只表示 `runtime_nodes.id`，不表示 Docker 内部 node。
  - 非 AWD requeue 仍允许重新选择健康 node。
  - 不引入旧 `node_id` alias 字段。
- Done criteria:
  - instance 模块内部不再暴露 `NodeID` 作为平台执行身份字段。
  - 旧容器清理、pending requeue、inventory lookup 全部按 `runtime_node_id` 工作。

### Slice 3: 执行路由、checker 和 AWD metadata 命名收敛

- Goal:
  - 将跨模块执行路由里的平台 node identity 改名为 `RuntimeNodeID`。
  - 保持 `container_id` / `network_id` 都通过 runtime node 归属定位。
- Dependencies:
  - Slice 2.
- Files:
  - Modify: `code/backend/internal/module/contest/ports/awd.go`
  - Modify: `code/backend/internal/module/contest/ports/checker_runner.go`
  - Modify: `code/backend/internal/module/contest/infrastructure/awd_container_inventory_repository.go`
  - Modify: `code/backend/internal/module/contest/infrastructure/awd_service_instance_repository.go`
  - Modify: `code/backend/internal/module/contest/application/jobs/awd_http_checker_runner.go`
  - Modify: `code/backend/internal/module/contest/application/jobs/awd_tcp_checker_runner.go`
  - Modify: `code/backend/internal/module/contest/application/jobs/awd_script_checker_runner.go`
  - Modify: `code/backend/internal/module/contest/application/jobs/awd_checker_preview.go`
  - Modify: `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - Modify: `code/backend/internal/app/composition/runtime_state_providers.go`
  - Modify: `code/backend/internal/app/composition/runtime_cleanup_target_adapter.go`
  - Modify: `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - Modify: `code/backend/internal/app/router_composition_typed_deps_test.go`
- Steps:
  - [x] 写失败 router 测试，断言 cleanup / file write / checker job 使用 `RuntimeNodeID` 字段并拒绝缺失 bound node 的旧容器操作漂移。
  - [x] 写失败 inventory provider 测试，断言 SQL select 使用 `runtime_node_id`，并返回 `FindRuntimeNodeIDByContainerID` 的结果。
  - [x] 运行：`cd code/backend && go test ./internal/app/composition -run 'TestRuntimeNodeExecutionRouter|TestRuntimeStateProviders' -count=1`
  - [x] 确认 RED：编译或断言因旧字段名失败。
  - [x] 将 checker metadata、AWD service runtime view、cleanup target、router helper 的平台 node 字段改为 `RuntimeNodeID`。
  - [x] 更新 contest infrastructure SQL projection，将 `inst.runtime_node_id` 投影为 `runtime_node_id`。
  - [x] 更新所有日志字段：平台主键使用 `runtime_node_id`；如果未来出现 Docker 内部身份，必须使用 `docker_node_id` / `engine_node_id`。
  - [x] 运行：`cd code/backend && go test ./internal/app/composition ./internal/module/contest/... -count=1`
  - [x] 修复因重命名暴露的 typed deps / package tests。
  - [ ] 按 `committing-changes` skill 和仓库 commit policy 提交本 slice。
- Validation:
  - `cd code/backend && go test ./internal/app/composition ./internal/module/contest/... -count=1`
- Validation evidence:
  - RED：`cd code/backend && go test ./internal/app/composition -run 'TestRuntimeNodeExecutionRouter|TestRuntimeStateProviders' -count=1` 因 `Instance.NodeID` / checker metadata 旧字段编译失败。
  - GREEN：同 focused command 通过。
  - GREEN：`cd code/backend && go test ./internal/app/composition ./internal/module/contest/... -count=1` 通过。
  - GREEN：`cd code/backend && go test ./internal/module/practice/application/commands -count=1` 通过，用于覆盖 practice port 字段改名的附加编译面。
- Review focus:
  - 不把 `RuntimeNodeID` 写进 runtime-agent / Docker 内部协议字段。
  - `container_id` 反查路径优先使用 `(runtime_node_id, container_id)` 对应索引。
  - 显式旧容器操作没有 fallback 到默认 node。
- Done criteria:
  - checker、AWD workspace、router 和 cleanup target 中的平台执行身份命名一致。
  - 日志和测试能区分平台 `runtime_node_id` 与未来 Docker 内部 id。

### Slice 4: Contest runtime placement repository

- Goal:
  - 新增 contest-owned AWD runtime placement repository。
  - 支持读取 active placement、创建 active placement、并发冲突后重读。
- Dependencies:
  - Slice 1.
- Files:
  - Create: `code/backend/internal/module/contest/entity/contest_runtime_placement.go`
  - Create: `code/backend/internal/module/contest/infrastructure/contest_runtime_placement_repository.go`
  - Create: `code/backend/internal/module/contest/infrastructure/contest_runtime_placement_repository_test.go`
- Steps:
  - [x] 写失败 repository 测试，断言没有 placement 时返回 not found / `exists=false`。
  - [x] 写失败 repository 测试，断言 `EnsureActiveContestRuntimePlacement(contestID, runtimeNodeID)` 创建 active placement。
  - [x] 写失败 repository 测试，断言同一 contest 再次 ensure 不会创建第二条 active placement，并返回已有 placement。
  - [x] 写失败 repository 测试，断言 released placement 不阻止新 active placement。
  - [x] 运行：`cd code/backend && go test ./internal/module/contest/infrastructure -run TestContestRuntimePlacementRepository -count=1`
  - [x] 确认 RED：repository / entity 尚不存在。
  - [x] 实现 `ContestRuntimePlacement` entity，`TableName()` 返回 `contest_runtime_placements`。
  - [x] 实现 repository：`FindActiveContestRuntimePlacement(ctx, contestID)` 和 `EnsureActiveContestRuntimePlacement(ctx, contestID, runtimeNodeID)`。
  - [x] 对唯一冲突使用重读 active placement，而不是吞掉错误后继续选新 node。
  - [x] 运行：`cd code/backend && go test ./internal/module/contest/infrastructure -run TestContestRuntimePlacementRepository -count=1`
  - [ ] 按 `committing-changes` skill 和仓库 commit policy 提交本 slice。
- Validation:
  - `cd code/backend && go test ./internal/module/contest/infrastructure -run TestContestRuntimePlacementRepository -count=1`
- Validation evidence:
  - RED：`cd code/backend && go test ./internal/module/contest/infrastructure -run TestContestRuntimePlacementRepository -count=1` 因 placement entity / repository 不存在编译失败。
  - GREEN：同 focused command 通过。
- Review focus:
  - Repository 位于 contest module，practice 只通过 port / adapter 使用。
  - 表只表达 placement，不表达 reservation。
  - 所有时间使用 UTC。
- Done criteria:
  - Contest module 能持久化和读取一场 AWD contest 的 active runtime node binding。

### Slice 5: AWD runtime node selector adapter

- Goal:
  - 让 practice 的 `SelectRuntimeNode(ctx, scope)` 对 AWD scope 使用 contest placement。
  - 已有 placement 时只验证绑定 node，不 fallback。
  - 没有 placement 时先默认选择，再持久化第一条 active placement。
- Dependencies:
  - Slice 4.
- Files:
  - Modify: `code/backend/internal/module/practice/ports/ports.go`
  - Modify: `code/backend/internal/app/composition/instance_practice_runtime_node_selector_adapter.go`
  - Modify: `code/backend/internal/module/container_runtime/infrastructure/node_repository.go` only if a narrower port wrapper is needed; prefer no change because `FindHealthyByID` already exists.
  - Modify: `code/backend/internal/app/composition/runtime_module_test.go` or add focused adapter tests.
  - Modify: `code/backend/internal/module/practice/application/commands/repository_stub_test.go`
- Target port sketch:

```go
type AWDRuntimePlacementStore interface {
    FindActiveContestRuntimePlacement(ctx context.Context, contestID int64) (*RuntimeNodeBinding, bool, error)
    EnsureActiveContestRuntimePlacement(ctx context.Context, contestID int64, runtimeNodeID int64) (*RuntimeNodeBinding, error)
}

type RuntimeNodeHealthLookup interface {
    FindHealthyRuntimeNodeByID(ctx context.Context, runtimeNodeID int64) (*RuntimeNodeBinding, error)
}
```

- Steps:
  - [x] 写失败 adapter 测试：AWD scope 已有 contest placement 指向 node A，默认 selector 会返回 node B，最终仍返回 node A。
  - [x] 写失败 adapter 测试：AWD scope 已有 placement 但 node A 不健康，返回 `ErrRuntimeNodeUnavailable`，且默认 selector 未被调用。
  - [x] 写失败 adapter 测试：AWD scope 缺少 placement store / health lookup 接线时返回 `ErrRuntimeNodeUnavailable`，且默认 selector 未被调用。
  - [x] 写失败 adapter 测试：AWD scope 无 placement 时调用默认 selector 选择 node A，并创建 active placement。
  - [x] 写失败 adapter 测试：非 AWD scope 继续调用默认 selector。
  - [x] 运行：`cd code/backend && go test ./internal/app/composition -run 'TestPracticeRuntimeNodeSelectorAdapter.*AWD|TestPracticeRuntimeNodeSelectorAdapter.*NonAWD' -count=1`
  - [x] 确认 RED：adapter 仍忽略 scope。
  - [x] 在 `practice/ports` 增加窄 port，表达 active placement store 和 runtime node health lookup。
  - [x] 修改 `practiceRuntimeNodeSelectorAdapter`：先判断 `scope.ContestMode == ContestModeAWD && scope.ContestID != nil`。
  - [x] 修改 AWD 接线缺失分支：返回 `ErrRuntimeNodeUnavailable`，禁止静默 fallback 到默认 selector。
  - [x] 实现已有 placement 分支：调用 healthy-by-id lookup，失败时直接返回 `ErrRuntimeNodeUnavailable`。
  - [x] 实现无 placement 分支：调用默认 selector，ensure placement；并发冲突后以 repository 返回的 active placement 为准。
  - [x] 运行 focused adapter tests 并确认 GREEN。
  - [x] 运行：`cd code/backend && go test ./internal/app/composition ./internal/module/practice/application/commands -count=1`
  - [ ] 按 `committing-changes` skill 和仓库 commit policy 提交本 slice。
- Validation:
  - `cd code/backend && go test ./internal/app/composition -run 'TestPracticeRuntimeNodeSelectorAdapter.*AWD|TestPracticeRuntimeNodeSelectorAdapter.*NonAWD' -count=1`
  - `cd code/backend && go test ./internal/app/composition ./internal/module/practice/application/commands -count=1`
- Validation evidence:
  - RED：focused adapter command 因 `newPracticeRuntimeNodeSelectorAdapter` 还未接收 placement/health 依赖而编译失败。
  - RED：`TestPracticeRuntimeNodeSelectorAdapterRejectsAWDWhenPlacementDependenciesMissing` 首次失败，当前实现对 AWD 接线缺失 fallback 到默认 selector 并返回 nil error。
  - GREEN：focused adapter command 通过。
  - GREEN：`cd code/backend && go test ./internal/app/composition ./internal/module/practice/application/commands -count=1` 通过。
- Review focus:
  - AWD branch 不 fallback 到 default selector when placement exists.
  - First placement creation is idempotent under concurrency.
  - Practice does not import contest infrastructure directly.
- Done criteria:
  - 一场 AWD contest 只从多个 runtime nodes 中选择一个 active placement；不是全平台只允许一个 runtime node 参与所有 AWD 比赛。
  - 有 active placement 后，该 contest 的所有 AWD instance 选择结果一致。

### Slice 6: AWD restart、pending scheduler 和 desired reconcile 行为收敛

- Goal:
  - 防守重启、pending scheduler 和 desired reconcile 都复用 contest placement。
  - 绑定 node 不可用时保持等待 / backoff，不自动漂移。
- Dependencies:
  - Slice 5.
- Files:
  - Modify: `code/backend/internal/module/practice/application/commands/instance_start_service.go`
  - Modify: `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
  - Modify: `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
  - Modify: `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
  - Modify: `code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go`
  - Modify: `code/backend/internal/module/practice/application/commands/instance_restart_awd_service_test.go`
  - Modify: `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
  - Modify: `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
- Steps:
  - [x] 写失败测试：AWD 防守重启原实例在 node A，默认 selector 返回 node B，重启后的 runtime create request 仍使用 node A。
  - [x] 写失败测试：scheduler 处理 AWD pending instance 且 placement node 不可用时，实例回到 `pending`，不绑定 node B，不创建 runtime。
  - [x] 写失败测试：desired reconcile 为缺失 AWD scope 创建 instance 时复用 contest placement node。
  - [x] 写失败测试：非 AWD pending instance 在旧 node 不可用时仍可按默认 selector 重新选择健康 node。
  - [x] 运行：`cd code/backend && go test ./internal/module/practice/application/commands -run 'Test.*AWD.*RuntimeNode|Test.*Desired.*RuntimeNode|Test.*Pending.*RuntimeNode' -count=1`
  - [x] 确认 RED：当前实现会重新选择默认 node 或无法表达 placement。
  - [x] 调整 restart path：清理旧 runtime 后，AWD scope 的下一次 bind 必须来自 contest placement；必要时将 reset options 扩展为可保留 authoritative runtime placement。
  - [x] 调整 `processPendingInstance` 断言：`selectRuntimeNode` 返回 `ErrRuntimeNodeUnavailable` 时只 requeue / wait，不 mark failed。
  - [x] 调整 runtime create request 构造：从 `instance.RuntimeNodeID` 读取 runtime node，禁止继续使用旧 `NodeID`。
  - [x] 调整 desired reconcile failure recording：bound node unavailable 进入现有 backoff / suppress 机制，不自动改绑。
  - [x] 运行 focused practice command tests 并确认 GREEN。
  - [x] 运行：`cd code/backend && go test ./internal/module/practice/... -count=1`
  - [ ] 按 `committing-changes` skill 和仓库 commit policy 提交本 slice。
- Validation:
  - `cd code/backend && go test ./internal/module/practice/application/commands -run 'Test.*AWD.*RuntimeNode|Test.*Desired.*RuntimeNode|Test.*Pending.*RuntimeNode' -count=1`
  - `cd code/backend && go test ./internal/module/practice/... -count=1`
- Validation evidence:
  - `2026-06-21`：新增 `TestProcessPendingAWDInstancePassesPlacementScopeWhenRuntimeNodeUnavailable`，首次运行 focused 命令 RED，原因是 AWD pending 测试未接入 contest runtime subject fixture，selector 未被调用；补齐 fixture 后该 guard 验证 pending scheduler 会把完整 AWD scope 交给 selector，`ErrRuntimeNodeUnavailable` 时实例回到 `pending`、清空 `runtime_node_id` 且不创建 runtime。
  - `2026-06-21`：扩展 `TestReconcileDesiredAWDInstancesCreatesMissingInstance`，断言 desired reconcile 创建缺失 AWD instance 时传递 contest/team/service scope，并持久化 selector 返回的 `runtime_node_id`。
  - `2026-06-21`：Slice 2-5 已完成 `processPendingInstance` requeue、runtime create request 读取 `RuntimeNodeID`、AWD selector adapter sticky placement 和 desired reconcile backoff 接线；本 slice 未再需要生产代码改动。
  - GREEN：`cd code/backend && go test ./internal/module/practice/application/commands -run 'Test.*AWD.*RuntimeNode|Test.*Desired.*RuntimeNode|Test.*Pending.*RuntimeNode|TestProcessPendingAWDInstancePassesPlacementScopeWhenRuntimeNodeUnavailable|TestReconcileDesiredAWDInstancesCreatesMissingInstance' -count=1`
  - GREEN：`cd code/backend && go test ./internal/module/practice/... -count=1`
- Review focus:
  - Defense restart 不引入 emergency recreate 的隐式行为。
  - Pending scheduler 的 non-AWD 行为不被 AWD placement 规则破坏。
  - `ErrRuntimeNodeUnavailable` 是等待信号，不被错误包装成永久失败。
- Done criteria:
  - 比赛过程中防守重启不会把 AWD instance 静默调度到另一个 runtime node。
  - Desired reconcile 在 bound node 不可用时等待，而不是跨 node 补建。

### Slice 7: 文档、图和 TODO 收口

- Goal:
  - 更新架构事实源中的 runtime node identity 和 AWD sticky placement 规则。
  - 将已纳入本 plan 的 todo 项标记为 tracked，保留 future work。
- Dependencies:
  - Slices 1-6 implemented.
- Files:
  - Modify: `docs/architecture/backend/01-system-architecture.md`
  - Modify: `docs/architecture/backend/03-container-architecture.md`
  - Modify: `docs/architecture/backend/05-key-flows.md`
  - Review/optional modify: `docs/architecture/backend/platform-overall-architecture.drawio`
  - Review/optional modify: `docs/architecture/backend/platform-overall-architecture.svg`
  - Modify: `docs/todos/2026-06-21-awd-runtime-node-identity-and-placement.md`
- Steps:
  - [x] 更新系统架构文档：把 `instances.node_id` 改为 `instances.runtime_node_id`，并说明这是平台 runtime node 主键。
  - [x] 更新容器架构文档：补充 `(runtime_node_id, container_id)` / `(runtime_node_id, network_id)` 复合执行身份。
  - [x] 更新关键流程文档：补充 AWD contest placement、defense restart sticky node、desired reconcile no-fallback 行为。
  - [x] 检查 drawio / svg 是否仍显式写 `instances.node_id / container_id`；如仍是活动图，更新为 `runtime_node_id / container_id`。
  - [x] 更新 TODO：本 plan 覆盖的前四项标记为 tracked by impl-plan；保留 emergency recreate 和 full reservation / capacity preflight 作为 Open。
  - [x] 运行：`python3 scripts/check-docs-consistency.py`
  - [x] 运行：`rg -n "instances\\.node_id|\\bnode_id\\b" docs/architecture/backend code/backend/internal -g '*.md' -g '*.go'`
  - [x] 人工审查 `rg` 输出，只保留历史说明、Docker 内部字段或明确不是平台 runtime node 的合法命中。
  - [ ] 按 `committing-changes` skill 和仓库 commit policy 提交本 slice。
- Validation:
  - `python3 scripts/check-docs-consistency.py`
  - `rg -n "instances\\.node_id|\\bnode_id\\b" docs/architecture/backend code/backend/internal -g '*.md' -g '*.go'`
- Validation evidence:
  - GREEN：`python3 scripts/check-docs-consistency.py`
  - `rg -n "instances\\.node_id|\\bnode_id\\b" docs/architecture/backend code/backend/internal -g '*.md' -g '*.go'` 仅剩 `code/backend/internal/app/runtime_node_migration_test.go` 中的 migration guard / down migration 历史字段断言。
  - 额外修正 `container_runtime/application/jobs/node_health_service.go` 中平台主键日志字段：`node_id` -> `runtime_node_id`。
- Review focus:
  - 文档不把 plan 写成当前事实，必须等代码落地后再更新事实源。
  - Future work 仍留在 todo，不混进本轮实现范围。
  - `node_id` 剩余命中都能解释清楚。
- Done criteria:
  - 架构文档、TODO 和代码命名一致。
  - 后续 emergency recreate / reservation 仍有明确 backlog。

## Validation

Run after all implementation slices:

```bash
cd code/backend && go test ./internal/app ./internal/module/instance/... ./internal/module/contest/... ./internal/module/practice/... -count=1
cd code/backend && go test ./tests/... -run 'AWD|RuntimeNode|Instance' -count=1
python3 scripts/check-docs-consistency.py
rg -n "instances\\.node_id|\\bnode_id\\b" docs/architecture/backend code/backend/internal -g '*.md' -g '*.go'
```

预期结果：

- Go tests pass for app composition, instance, contest and practice modules.
- Docs consistency check passes.
- Remaining `node_id` matches are either:
  - Docker / engine internal identity explicitly named `docker_node_id` / `engine_node_id`;
  - generic method names on `RuntimeNode` owner where `ID` is already scoped by type;
  - migration down path or historical explanation;
  - external protocol fields intentionally outside this task and documented as future migration.

实际验证记录：

- `2026-06-21`：`cd code/backend && go test ./internal/app ./internal/module/instance/... ./internal/module/contest/... ./internal/module/practice/... -count=1`
  - `./internal/module/instance/...`、`./internal/module/contest/...`、`./internal/module/practice/...` 通过。
  - `./internal/app` 失败；失败集中在既有 full-router / typed-deps / 路由结构 guardrail 和缺表问题，不属于本任务改动范围。代表性失败包括 `contest_realtime_outbox`、`platform_event_outbox` 缺表，以及 `TestAssessmentModuleUsesTypedCrossModuleDeps`、`TestBuildInstanceModuleDelegatesToSubBuilders`、`TestUserSelfRoutesAreExtractedIntoDedicatedRegistrarFile`。
- `2026-06-21`：`cd code/backend && go test ./tests/... -run 'AWD|RuntimeNode|Instance' -count=1` 通过。
- `2026-06-21`：`cd code/backend && go test ./internal/module/container_runtime/... -count=1` 通过。
- `2026-06-21`：`python3 scripts/check-docs-consistency.py` 通过。
- `2026-06-21`：`rg -n "instances\\.node_id|\\bnode_id\\b" docs/architecture/backend code/backend/internal -g '*.md' -g '*.go'` 仅剩 `code/backend/internal/app/runtime_node_migration_test.go` 中 migration guard / down migration 历史字段断言。

## Architecture-Fit Self-Check

- Target boundary explicit:
  - Yes. `container_runtime` owns runtime node health / registry, `contest` owns contest placement, `practice` owns lifecycle consumption, `instance` owns per-instance runtime attempt identity.
- Shared layers and reuse points named:
  - Yes. Reuses `FindHealthyByID`, default selector, practice selector port, runtime execution router and existing desired reconcile backoff.
- Structural convergence rather than output-only fix:
  - Yes. The plan migrates DB column, Go naming, auxiliary indexes and AWD placement persistence; it does not only patch one restart path.
- Immediate second-round redesign risk:
  - Low for confirmed scope. Full reservation and emergency recreate are intentionally deferred and tracked as separate todo items.
- Deferred structural work captured:
  - Yes. Emergency recreate product shape and full runtime reservation / capacity preflight remain in `docs/todos/2026-06-21-awd-runtime-node-identity-and-placement.md`.

## Execution Handoff

Plan complete only means the implementation path is written. Execution should start in a bound code-workflow task context:

1. Run `bash scripts/start-implementation.sh 2026-06-21-awd-runtime-node-identity-and-placement`.
2. Use `superpowers:executing-plans` or `superpowers:subagent-driven-development`.
3. Complete slices in order because later slices depend on migration and naming changes.
4. Commit each slice separately after running the slice validation and `committing-changes` skill.
