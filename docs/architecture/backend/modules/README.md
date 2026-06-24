# 后端模块设计索引

> 状态：Current
> 事实源：`code/backend/internal/module/`、`code/backend/internal/app/composition/`
> 替代：无

## 定位

本目录参考 `zhicore-go` 的 `docs/architecture/services/` 拆分方式，记录 CTF 后端模块化单体中各业务模块和运行时能力模块的具体设计。

- 负责：说明每个模块的职责边界、入口、主要用例、数据与副作用、跨模块依赖和 Guardrail。
- 不负责：把当前单体进程描述成多进程微服务，或替代 `docs/architecture/backend/01-system-architecture.md` 的总体分层说明。

## 当前设计

- `code/backend/internal/module/<module>/`
  - 负责：承载模块内 `api`、`application`、`domain`、`ports`、`infrastructure`、`runtime` 与 `contracts` 等层；实际存在的子目录以源码为准，不强制每个模块拥有完整层级。
  - 不负责：跨模块直接导入其他模块的 `infrastructure`，或让 `api` / handler 承担业务 owner。
- `code/backend/internal/app/composition/`
  - 负责：把模块 runtime 组合成进程级视图，例如 `ContainerRuntimeModule`、`InstanceModule`、`OpsModule`，并注入跨模块能力。
  - 不负责：实现模块业务规则、持久化细节或状态机。

## 阅读顺序

1. 先读 [../01-system-architecture.md](../01-system-architecture.md)，确认单体部署、Onion 分层和 composition root。
2. 再读本索引，选择具体模块文档。
3. 涉及 API、OpenAPI、题包格式或导入导出契约时，再进入 `docs/contracts/`。
4. 涉及 AWD、题包、复盘、报告等跨模块专题时，再进入 `docs/architecture/features/专题架构索引.md`。

## 模块文档

| 模块 | 文档 | 类型 |
| --- | --- | --- |
| `auth` | [auth.md](./auth.md) | 认证与会话 owner |
| `identity` | [identity.md](./identity.md) | 用户、角色与资料 owner |
| `challenge` | [challenge.md](./challenge.md) | 题目、题包、镜像与 Flag owner |
| `container_runtime` | [container_runtime.md](./container_runtime.md) | 容器运行时能力模块 |
| `instance` | [instance.md](./instance.md) | 实例生命周期与访问 owner |
| `practice` | [practice.md](./practice.md) | 训练、开题、提交与进度 owner |
| `contest` | [contest.md](./contest.md) | 赛事、队伍、榜单与 AWD owner |
| `assessment` | [assessment.md](./assessment.md) | 技能画像、推荐与报告 owner |
| `ops` | [ops.md](./ops.md) | 审计、通知、WebSocket relay 与运营支撑 owner |
| `teaching_analysis` | [teaching_analysis.md](./teaching_analysis.md) | 教师视角只读查询聚合模块 |

## 模块设计模板

每篇模块文档默认包含：

- `定位`
- `事实来源`
- `当前设计`
- `API 入口设计`
- `Application / Service 设计`
- `数据设计`
- `边界`
- `主要用例`
- `数据与副作用`
- `跨模块依赖`
- `Guardrail`
- `已知限制`

## Guardrail

- `code/backend/internal/module/architecture_test.go`：模块通用 Onion 边界、跨模块 private import、runtime 装配约束。
- `code/backend/internal/module/*/architecture_test.go`：模块局部边界和迁移守卫。
- `code/backend/internal/app/architecture_rules_test.go`：进程装配层禁止 concrete cross-module import。
- `code/backend/internal/app/composition/architecture_test.go`：composition 层禁止回退到旧运行时物理模块。
- `python3 scripts/check-docs-consistency.py`：检查架构文档状态、当前设计结构和引用。
