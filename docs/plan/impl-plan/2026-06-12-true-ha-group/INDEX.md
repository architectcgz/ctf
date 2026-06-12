<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# 真正高可用控制面与运行时恢复 Task Group

**Task Group Slug:** `2026-06-12-true-ha-group`

**Goal:** 建立控制面真正高可用能力，支持 Redis Sentinel、PostgreSQL HA、共享存储、分布式事件总线、SSH Gateway draining 和 runtime 节点健康检查与 failover

**Status:** `in-progress`（T1 / T2 / T3 已合入，T4 / T5 仍待实现）

**Created:** `2026-06-12T03:00:00Z`

---

## Overview

- Background: 当前控制面各组件为单点，需要建立真正 HA 能力
- Motivation: 支持多副本部署、节点失效自动恢复、无单点故障
- Scope: Redis Sentinel 接入、PostgreSQL HA 连接、共享存储收口、分布式事件总线、SSH Gateway draining、runtime 健康检查
- Non-Goals: 本阶段不涉及跨地域容灾、实时热迁移

## Slices

### Slice 1: Redis Sentinel 与 PostgreSQL HA 接入

- Task Slug: `2026-06-12-redis-sentinel-and-postgres-ha-connectivity`
- Status: `completed`
- Plan: [implementation-plan](redis-sentinel-and-postgres-ha-connectivity.md)
- Depends On: 无
- Notes: 已合入 `3fc5667f4 feat(backend): 支持 Redis Sentinel 接入`；当前 plan 正文仍保留在 task group 下，待后续统一归档

### Slice 2: 共享存储 owner 收口

- Task Slug: `2026-06-12-shared-storage-owner-convergence`
- Status: `completed`
- Plan: [implementation-plan](shared-storage-owner-convergence.md)
- Depends On: `2026-06-12-redis-sentinel-and-postgres-ha-connectivity`
- Notes: 报告/附件 shared storage 与密钥共源约束已合入；主要代码提交为 `ba35ea9c9`，后续补修为 `a0d1ccc5b`

### Slice 3: 跨副本事件总线与 outbox relay

- Task Slug: `2026-06-12-distributed-event-bus-and-outbox-relay`
- Status: `not-started`
- Plan: [implementation-plan](distributed-event-bus-and-outbox-relay.md)
- Depends On: `2026-06-12-redis-sentinel-and-postgres-ha-connectivity`
- Notes: 并行于 Slice 2，同样依赖基础设施 HA

### Slice 4: SSH Gateway HA 与 draining

- Task Slug: `2026-06-12-ssh-gateway-ha-and-draining`
- Status: `completed`
- Plan: [implementation-plan](../../archive/impl-plan/2026-06/2026-06-12-ssh-gateway-ha-and-draining-implementation-plan.md)
- Depends On: `2026-06-12-redis-sentinel-and-postgres-ha-connectivity`
- Notes: 已合入 `7a9ab91a8 feat(backend): 支持 SSH gateway 多副本摘流`；实施计划已归档到 `docs/plan/archive/impl-plan/2026-06/`

### Slice 5: Runtime 节点健康检查与 failover 重建

- Task Slug: `2026-06-12-runtime-node-health-and-failover-rebuild`
- Status: `not-started`
- Plan: [implementation-plan](runtime-node-health-and-failover-rebuild.md)
- Depends On: `2026-06-12-shared-storage-owner-convergence, 2026-06-12-distributed-event-bus-and-outbox-relay, 2026-06-12-ssh-gateway-ha-and-draining`
- Notes: 依赖 Slice 2/3/4 完成

## Dependencies Graph

```
Slice 1 (Redis/PostgreSQL HA 基础设施)
  ├─> Slice 2 (共享存储)
  ├─> Slice 3 (事件总线)
  └─> Slice 4 (SSH Gateway)
       └─> Slice 5 (Runtime failover，依赖 2/3/4)
```

## Integration Validation

整个 task group 完成后的集成验证：

- [ ] 多副本控制面部署验证
- [ ] Redis Sentinel failover 演练
- [ ] PostgreSQL HA 切换演练
- [ ] SSH Gateway draining 与恢复演练
- [ ] Runtime 节点失效与自动重建演练
- [ ] 文档完整性检查

## Completion Criteria

Task group 视为完成的条件：

- [ ] 所有 5 个 slice 的 implementation plan 已归档
- [ ] 所有 slice 的代码已合并到 main
- [ ] 集成验证全部通过
- [ ] 架构文档已更新：`docs/architecture/backend/01-system-architecture.md`
- [ ] 运维文档已更新：`docs/operations/runtime-agent-deployment.md`
- [ ] 无 blocker 级别的 residual risk

## Progress Tracking

| Slice | Status | Started | Completed | Notes |
|-------|--------|---------|-----------|-------|
| 1 | completed | 2026-06-12 | 2026-06-12 | 已合入 `3fc5667f4` |
| 2 | completed | 2026-06-12 | 2026-06-12 | 已合入 `ba35ea9c9` / `a0d1ccc5b` |
| 3 | not-started | - | - | 事件总线与 outbox relay |
| 4 | completed | 2026-06-12 | 2026-06-12 | 已合入 `7a9ab91a8`；plan 已归档 |
| 5 | not-started | - | - | Runtime 健康检查与 failover |

## Notes

- Task-group-level decisions: 优先保证基础设施层 HA 可用，再逐步推进业务层 HA
- Cross-slice coordination: 当前剩余实现可以先独立推进 Slice 3；Slice 5 仍需等待 Slice 3 完成后再进入主实现
- Known risks: 真实 Sentinel/PostgreSQL HA 演练需要运维环境支持
