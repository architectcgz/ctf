# 后端架构索引

> 状态：Current
> 事实源：`code/backend/` 当前模块布局、路由装配、运行时实现与架构守卫
> 替代：无

## 定位

`docs/architecture/backend/` 只保留当前后端架构事实与已经采用的专题设计。

- 总览文档回答“系统现在怎么组织、谁负责什么、主链路怎么走”。
- `design/` 目录只保留已经采用、但暂时不适合并入总览的后端专题事实。
- 过程方案、迁移步骤和任务分解不放这里，分别回到 `docs/plan/impl-plan/`、`docs/tasks/`、`practice/`。

## 读取顺序

1. [01-system-architecture.md](./01-system-architecture.md)
2. [02-database-design.md](./02-database-design.md)
3. [03-container-architecture.md](./03-container-architecture.md)
4. [04-api-design.md](./04-api-design.md)
5. [05-key-flows.md](./05-key-flows.md)
6. [06-file-storage.md](./06-file-storage.md)
7. [07-modular-monolith-refactor.md](./07-modular-monolith-refactor.md)
8. [design/README.md](./design/README.md)

## 当前活动文档

| 文档 | 说明 |
| --- | --- |
| [01-system-architecture.md](./01-system-architecture.md) | 进程级 composition、模块 owner、主依赖方向与运行时边界 |
| [02-database-design.md](./02-database-design.md) | PostgreSQL 主模型、读写分工、关键表关系与持久化约束 |
| [03-container-architecture.md](./03-container-architecture.md) | Docker runtime、网络隔离、资源限制、端口与 ACL |
| [04-api-design.md](./04-api-design.md) | HTTP / WebSocket 契约、错误码、鉴权与返回结构 |
| [05-key-flows.md](./05-key-flows.md) | 练习、竞赛、实例、题包导入等关键调用链 |
| [06-file-storage.md](./06-file-storage.md) | 文件、导出产物、题包与运行时存储边界 |
| [07-modular-monolith-refactor.md](./07-modular-monolith-refactor.md) | 当前模块化单体边界、拆分约束与历史兼容说明 |

## 已采用专题

| 文档 | 说明 |
| --- | --- |
| [design/awd-engine-migration.md](./design/awd-engine-migration.md) | AWD 运行时引擎化迁移后的当前事实 |
| [design/contest-status-state-machine.md](./design/contest-status-state-machine.md) | 竞赛状态机、定时推进与幂等 guardrail |
| [design/instance-sharing.md](./design/instance-sharing.md) | 题目实例共享策略与限制条件 |

## Guardrail

- 模块依赖方向：`code/backend/internal/module/architecture_test.go`
- 进程级装配边界：`code/backend/internal/app/architecture_rules_test.go`
- 路由与组合装配：`code/backend/internal/app/router_test.go`
- 全链路装配验证：`code/backend/internal/app/full_router_integration_test.go`
