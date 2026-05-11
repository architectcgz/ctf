# 后端专题设计

> 状态：Current
> 事实源：`docs/architecture/backend/design/` 当前 adopted 专题索引
> 替代：无

## 定位

这里收录已经采用、但暂时不适合直接并入 `01` 到 `07` 总览文档的后端专题事实。

- 这些文档依然属于当前事实源，不是 Draft。
- 如果某个专题已经被总览文档完整吸收，应从这里移除活动入口。
- 仍在推演的后端方案不放这里，进入 `docs/design/` 或 `docs/plan/impl-plan/`。

## 当前专题

- `awd-engine-migration.md`
  - 负责：说明 AWD 运行态引擎化迁移后的当前边界和组件落点
  - 不负责：承载尚未采用的替代运行时方案

- `contest-status-state-machine.md`
  - 负责：说明竞赛状态机、定时推进与幂等 guardrail
  - 不负责：承载一次性迁移步骤或执行清单

- `instance-sharing.md`
  - 负责：说明题目实例共享策略、适用范围和限制条件
  - 不负责：替代 `03-container-architecture.md` 的整体容器事实

## 使用方式

1. 先读 `../README.md` 和 `../01-system-architecture.md`，确认整体边界。
2. 当总览文档需要跳转到 adopted 专题时，再进入这里。
3. 如果专题已经稳定并成为全局约束，优先考虑并回 owning 总览文档。
