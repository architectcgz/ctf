# CTF 架构与最终设计入口

> 状态：Current
> 事实源：`docs/architecture/` 当前入口与索引
> 替代：无

## 定位

`docs/architecture/` 是当前项目的最终架构与最终设计事实源。

- 架构边界、模块职责、接口协作、页面与工作台最终设计，都从这里读取。
- `practice/superpowers-plan-index.md`、`docs/reviews/` 只保留过程与证据，不覆盖当前事实。
- 单题题包、题面和解法不作为本目录的常驻入口；已落地内容回到题包目录，仍在推演的方案进入 `docs/design/`。

## 当前活动入口

| 入口 | 说明 |
| --- | --- |
| [backend/README.md](./backend/README.md) | 后端总览、数据库、容器、API、关键流程与已采用专题入口 |
| [frontend/README.md](./frontend/README.md) | 前端分层、路由、状态管理、组件体系、主题与间距规则入口 |
| [features/专题架构索引.md](./features/专题架构索引.md) | 业务专题、产品能力与跨模块边界专题入口 |
| [backend/design/README.md](./backend/design/README.md) | 已采用但暂未并入总览的后端专题入口 |

## 当前非活动入口

| 文档 | 状态 | 说明 |
| --- | --- | --- |
| [00-hard-points-and-solutions.md](./00-hard-points-and-solutions.md) | Superseded | 2026-03 的实现难点清单，仅保留早期任务拆分背景，不再作为当前架构事实入口 |

## 读取顺序

1. 当前先读本文件，确认入口和状态归属。
2. 再按需要进入 [backend/README.md](./backend/README.md) 或 [frontend/README.md](./frontend/README.md)。
3. 需要专题边界时，进入 [features/专题架构索引.md](./features/专题架构索引.md)。
4. 需要接口与字段契约时，进入 `docs/contracts/`。
5. 需要过程方案、实现计划或历史评审时，回到 `docs/plan/impl-plan/`、`docs/reviews/`、`practice/`。

## 机械化 Guardrail

- `scripts/check-architecture.sh --quick`：快速检查后端模块依赖方向、前端分层边界和关键历史腐蚀基线。
- `scripts/check-architecture.sh --full`：在 quick 基础上补充 overlay 结构约束和前端主题 token 检查。
- `python3 scripts/check-docs-consistency.py`：检查架构文档状态、索引引用与 `## 当前设计` 结构底线。
- `scripts/check-consistency.sh`：检查 harness 入口、脚本接线与本地 guardrail 是否接入。

主要代码级 guardrail：

- 后端：`code/backend/internal/module/architecture_test.go`
- 进程装配：`code/backend/internal/app/architecture_rules_test.go`
- 前端分层：`code/frontend/src/__tests__/architectureBoundaries.test.ts`
- 路由边界：`code/frontend/src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- 共享模板：`code/frontend/src/components/common/__tests__/ModalTemplates.test.ts`

## 历史迁移说明

- 原 `design-system/ctf-platform/` 下仍有效的最终设计已经并入 `docs/architecture/frontend/`。
- 原 superpowers specs 下仍有效的专题设计已经并入 `docs/architecture/features/`。
- superpowers 的过程资料索引保留在 `practice/superpowers-plan-index.md`，不再作为最终设计入口。
