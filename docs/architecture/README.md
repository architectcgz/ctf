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
5. 需要过程方案、实现计划或历史评审时，先读 `docs/plan/README.md`，再按需要进入活动 `docs/plan/impl-plan/`、`docs/reviews/`、`practice/`；旧历史计划默认通过 Git 历史追溯。

## 机械化 Guardrail

- `scripts/check-backend-architecture.sh --quick`：快速检查后端模块依赖方向。
- `scripts/check-backend-architecture.sh --full`：在 quick 基础上补充 `internal/app` 的 concrete cross-module import、context architecture，以及后端测试分层 guardrail。
- `scripts/check-frontend-architecture.sh --quick`：快速检查前端分层边界、route view 约束和 `:deep` 存量守卫。
- `scripts/check-frontend-architecture.sh --full`：在 quick 基础上补充前端热点文件增长守卫、feature owner boundary、overlay 结构约束和前端主题 token 检查。
- `scripts/check-architecture.sh --quick`：聚合执行 backend quick + frontend quick。
- `scripts/check-architecture.sh --full`：聚合执行 backend full + frontend full。
- `harness/workflow-plugins/code-workflow/`：项目把本地 guardrail 挂到 `pre-commit-quick`、`completion-full`、`review-governance` 这三个 workflow stage 的注册点。
- `python3 scripts/check-docs-consistency.py`：检查架构文档状态、索引引用与 `## 当前设计` 结构底线。
- `scripts/check-review-governance.sh`：检查 harness 入口、脚本接线与本地 guardrail 是否接入。

主要代码级 guardrail：

- 后端：`code/backend/internal/module/architecture_test.go`
- 进程装配：`code/backend/internal/app/architecture_rules_test.go`
- 后端 context 架构：`code/backend/internal/app/backend_context_architecture_test.go`
- 后端测试架构：`code/backend/tests/architecture/test_architecture_test.go`
- 前端架构策略：`code/frontend/scripts/frontend-architecture-policy.json`
  - 这是前端结构约束的单点事实源；后续若要调整分层、route view 边界或增长守卫，先改策略，再让脚本与测试消费它。
- 前端分层：`code/frontend/src/__tests__/architectureBoundaries.test.ts`
- 路由边界：`code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
- 前端增长守卫：`code/frontend/scripts/check-frontend-growth-guard.mjs`
- 前端深度选择器守卫：`code/frontend/scripts/check-vue-deep-guard.mjs`
- AWD owner 边界：`code/frontend/src/features/contest-awd-admin/model/useAwdOwnerBoundaries.test.ts`
- 共享模板：`code/frontend/src/shared/ui/common/__tests__/ModalTemplates.test.ts`

## 历史迁移说明

- 原 `design-system/ctf-platform/` 下仍有效的最终设计已经并入 `docs/architecture/frontend/`。
- 原 superpowers specs 下仍有效的专题设计已经并入 `docs/architecture/features/`。
- superpowers 的过程资料索引保留在 `practice/superpowers-plan-index.md`，不再作为最终设计入口。
