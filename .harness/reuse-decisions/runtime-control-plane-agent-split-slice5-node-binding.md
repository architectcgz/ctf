# Reuse Decision

## Change type
entity / repository / migration / selection / application / composition

## Existing code searched
- code/backend/internal/app/composition
- code/backend/internal/module/runtime/runtime
- code/backend/internal/module/runtime/infrastructure
- code/backend/internal/module/runtime/entity
- code/backend/internal/module/practice/runtime
- code/backend/internal/module/practice/application/commands
- code/backend/internal/module/practice/infrastructure
- code/backend/internal/module/instance/entity
- docs/plan/impl-plan/2026-06-02-runtime-control-plane-agent-split-plan.md

## Similar implementations found
- `code/backend/internal/app/composition/runtime_module.go` 已经把 local / remote runtime execution 的选择收口到 composition，说明 phase5 不该再让 practice / contest 自己判断节点来源。
- `code/backend/internal/module/practice/application/commands/instance_start_service.go` 已经是实例创建事务 owner，`instance` 记录的持久化字段应继续在这里一次性写入，而不是事后补写节点绑定。
- `code/backend/internal/module/practice/runtime/module.go` 已经负责 practice 侧 wiring，适合在这里注入“如何选择 runtime node”的窄依赖，而不是让 application 直接 new runtime repository。
- `code/backend/internal/module/runtime/infrastructure/repository.go` 与 `code/backend/internal/module/practice/infrastructure/repository.go` 已经是 `instances` 表的主要持久化入口，新增 `node_id` 应沿现有持久化模型扩展，不另起第二套实例表或影子状态。

## Decision
extend_existing

## Reason
phase5 的第一批最小正确目标不是一次做完多节点调度，而是先把“节点是明确数据模型”落地，并让新建实例从创建时就绑定 `node_id`。如果继续只在配置里保留一根 remote endpoint：

- 后续 checker / cleanup / AWD 调度仍然没有稳定 authority；
- `instances` 依旧只知道容器和网络细节，不知道自己属于哪个 runtime node；
- 之后再补 node binding 时会变成二次迁移，而不是在实例 owner 入口一次收口。

因此这批改动先交付：

- `runtime_nodes` 持久化模型与最小 selector；
- `instances.node_id` 字段与实例创建时的绑定；
- 当前单节点 local / remote 模式下的默认节点同步，避免因为引入 node 模型直接打断现有开发链路。

本轮不提前做复杂均衡、管理 API 或 checker / AWD 全量节点编排；这些在 node authority 落地后再沿同一模型继续推进。

## Files to modify
- .harness/reuse-decisions/runtime-control-plane-agent-split-slice5-node-binding.md
- code/backend/migrations/000012_create_runtime_nodes.up.sql
- code/backend/migrations/000012_create_runtime_nodes.down.sql
- code/backend/internal/module/runtime/entity/**
- code/backend/internal/module/runtime/ports/**
- code/backend/internal/module/runtime/infrastructure/**
- code/backend/internal/module/runtime/runtime/module.go
- code/backend/internal/app/composition/runtime_module.go
- code/backend/internal/app/composition/instance_module.go
- code/backend/internal/app/composition/practice_module.go
- code/backend/internal/module/instance/entity/instance.go
- code/backend/internal/module/practice/ports/ports.go
- code/backend/internal/module/practice/runtime/module.go
- code/backend/internal/module/practice/application/commands/**
- code/backend/internal/module/practice/infrastructure/**

## After implementation
- runtime node 会成为正式持久化事实，而不是隐藏在配置里的 endpoint。
- 新创建的实例会在落表时带上明确 `node_id`。
- 当前单节点 local / remote 模式仍然能通过默认节点同步继续工作。
- 后续 checker / AWD / cleanup 迁到按节点解析执行面时，可以直接复用这次落下来的 node authority，而不是再补第二套绑定逻辑。
