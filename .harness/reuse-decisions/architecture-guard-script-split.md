# Reuse Decision

## Change type
architecture guardrail script split / backend + frontend architecture entry separation

## Existing code searched
- scripts/check-architecture.sh
- scripts/check-workflow-complete.sh
- scripts/check-consistency.sh
- scripts/doctor-local-harness.sh
- docs/architecture/README.md
- docs/architecture/backend/README.md
- docs/architecture/frontend/README.md
- code/backend/internal/module/architecture_test.go
- code/backend/internal/app/architecture_rules_test.go
- code/backend/internal/app/backend_context_architecture_test.go
- code/frontend/scripts/frontend-architecture-policy.json

## Similar implementations found
- 当前已经存在前端单点策略源：`code/frontend/scripts/frontend-architecture-policy.json`
- 当前已经存在后端独立的代码级守卫：`code/backend/internal/module/architecture_test.go`、`code/backend/internal/app/architecture_rules_test.go`
- 当前缺口不在“没有前后端各自的架构测试”，而在“顶层脚本和 workflow gate 仍把两边混在一个总入口里，且后端入口没有把 `internal/app` 这层装配守卫显式跑起来”

## Decision
refactor_existing

## Reason
本轮不是重写架构规则，而是把现有前后端 guardrail 拆成更清晰的入口：

- 新增 `scripts/check-backend-architecture.sh`
- 新增 `scripts/check-frontend-architecture.sh`
- 保留 `scripts/check-architecture.sh` 作为聚合入口
- 让 `scripts/check-workflow-complete.sh` 按 backend / frontend touched surface 分别触发对应守卫

同时顺手补齐后端入口，至少显式覆盖：

- `internal/module` 模块边界
- `internal/app` 进程级 concrete cross-module import / context architecture 边界

本轮不做：

- 重新设计后端 `router_test.go` / `full_router_integration_test.go` 的全量编排范围
- 变更前端架构策略本身
- 新增 allowlist 或弱化现有架构测试

## Files to modify
- .harness/reuse-decisions/architecture-guard-script-split.md
- docs/plan/impl-plan/2026-05-30-architecture-guard-script-split-plan.md
- docs/reviews/architecture/2026-05-30-architecture-guard-script-split-review.md
- scripts/check-backend-architecture.sh
- scripts/check-frontend-architecture.sh
- scripts/check-architecture.sh
- scripts/check-workflow-complete.sh
- scripts/check-consistency.sh
- scripts/doctor-local-harness.sh
- docs/architecture/README.md
- docs/architecture/backend/README.md
- docs/architecture/frontend/README.md

## After implementation
- 后端和前端架构守卫会有各自独立入口，聚合脚本只负责调度
- workflow gate 会根据 touched surface 分别触发 backend / frontend architecture checks
- 后端不会因为脚本分离而丢掉 `internal/app` 这层装配守卫
