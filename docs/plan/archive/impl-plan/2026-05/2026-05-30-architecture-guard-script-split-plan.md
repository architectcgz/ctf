> 状态：Current
> 事实源：架构守卫脚本拆分与 workflow gate 分流方案
> 替代：无

# Architecture Guard Script Split Plan

## 目标

- 把前后端架构守卫拆成各自独立入口
- 保留 `scripts/check-architecture.sh` 作为聚合入口，避免现有调用点失效
- 让 workflow gate 按 backend / frontend touched surface 分别触发守卫
- 拆分时补齐后端 `internal/app` 架构守卫，不弱化后端边界

## 非目标

- 不重写后端架构测试内容
- 不扩大到路由集成 / 全链路 HTTP smoke 全量校验
- 不调整前端架构策略约束本身

## 输入依据

- `scripts/check-architecture.sh`
- `scripts/check-workflow-complete.sh`
- `scripts/check-consistency.sh`
- `scripts/doctor-local-harness.sh`
- `docs/architecture/README.md`
- `docs/architecture/backend/README.md`
- `docs/architecture/frontend/README.md`
- `code/backend/internal/module/architecture_test.go`
- `code/backend/internal/app/architecture_rules_test.go`
- `code/backend/internal/app/backend_context_architecture_test.go`
- `code/frontend/scripts/frontend-architecture-policy.json`

## 当前结论

- 前端规则源已经分离到 `code/frontend/scripts/frontend-architecture-policy.json`，但顶层执行入口仍与后端混在一个总脚本里。
- 后端当前在总脚本里只显式跑了 `./internal/module`，没有把 `internal/app` 这层架构守卫作为后端入口的一部分明确接入。
- `scripts/check-workflow-complete.sh` 仍把架构检查视为一个总步骤，不利于后续继续独立演进前后端 guardrail。

## 设计边界

### Backend script 负责

- 模块依赖方向
- 进程级 concrete cross-module import 禁令
- backend context architecture 约束

### Frontend script 负责

- 前端分层边界
- route view 边界
- growth guard
- feature owner boundary
- overlay / theme 边界

### Aggregator script 负责

- 保留统一入口
- 串行调度 backend / frontend 子脚本

### Workflow gate 负责

- 根据 backend / frontend touched surface 分别触发对应守卫
- 文档或聚合脚本变化时同时触发两侧守卫

## 任务切片

### Slice 1：新增独立脚本

- 新增：
  - `scripts/check-backend-architecture.sh`
  - `scripts/check-frontend-architecture.sh`
- 更新：
  - `scripts/check-architecture.sh`

### Slice 2：分流 workflow gate 与 consistency wiring

- 更新：
  - `scripts/check-workflow-complete.sh`
  - `scripts/check-consistency.sh`
  - `scripts/doctor-local-harness.sh`

### Slice 3：同步架构文档与 review

- 更新：
  - `docs/architecture/README.md`
  - `docs/architecture/backend/README.md`
  - `docs/architecture/frontend/README.md`
  - `docs/reviews/architecture/archive/2026-05/2026-05-30-architecture-guard-script-split-review.md`

## 验证

- `bash scripts/check-backend-architecture.sh --full`
- `bash scripts/check-frontend-architecture.sh --full`
- `bash scripts/check-architecture.sh --full`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`
- `git diff --check`

## 残余风险

- 后端 `router_test.go` / `full_router_integration_test.go` 仍是独立验证面，本轮不纳入 backend architecture script 的默认执行范围。
- workflow gate 分流后，如果未来新增第三类架构守卫面，还需要继续扩充 touched-surface pattern。
