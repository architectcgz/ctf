# Assessment Report Not Found Contract Phase 5 Slice 14 Implementation Plan

## Objective

继续 phase 5 收窄 `assessment` application concrete allowlist，把 report service 的 GORM not-found 判断下沉到模块 repo contract：

- 去掉 `assessment/application/commands/report_service.go -> gorm.io/gorm`
- 保持“报告不存在 / 竞赛不存在”的现有错误码行为不变

## Non-goals

- 不改报表生成逻辑、异步 worker、导出格式或下载权限校验
- 不处理其他模块剩余的 GORM / HTTP concrete allowlist
- 不改 report repository 的 SQL 查询或数据模型

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository.go`

## Current Baseline

- `report_service.go` 当前直接 import `gorm.io/gorm`
- application 只是为了把 `gorm.ErrRecordNotFound` 映射成 `ErrContestNotFound` 或 `ErrNotFound`
- allowlist 当前保留：
  - `assessment/application/commands/report_service.go -> gorm.io/gorm`

## Chosen Direction

把 report service 的 not-found 语义收口成 `assessment` 自己的 repository contract：

1. 在 `assessment/ports` 暴露 `ErrAssessmentReportNotFound` 和 `ErrAssessmentContestNotFound`
2. `assessment/infrastructure/report_repository.go` 负责把 `gorm.ErrRecordNotFound` 映射成对应 sentinel
3. `report_service.go` 只判断 assessment 自己的 sentinel，不再知道 GORM sentinel
4. 保持 runtime wiring 和 repository 组合方式不变

## Ownership Boundary

- `assessment/application/commands/report_service.go`
  - 负责：报表用例编排和错误码映射
  - 不负责：知道 `gorm.ErrRecordNotFound` 这类 ORM 错误类型
- `assessment/infrastructure/report_repository.go`
  - 负责：把 persistence not-found 映射成 assessment 自己的 repo 契约
  - 不负责：决定 HTTP 错误码或导出业务语义

## Change Surface

- Add: `.harness/reuse-decisions/assessment-report-not-found-contract-phase5-slice14.md`
- Add: `docs/plan/impl-plan/2026-05-13-assessment-report-not-found-contract-phase5-slice14-implementation-plan.md`
- Modify: `code/backend/internal/module/assessment/ports/ports.go`
- Modify: `code/backend/internal/module/assessment/application/commands/report_service.go`
- Modify: `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- Modify: `code/backend/internal/module/assessment/infrastructure/report_repository.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

- [x] Slice 1: 收口 report repo not-found contract
  - 目标：report service 不再 import `gorm.io/gorm`，not-found 行为保持一致
  - 验证：
    - `cd code/backend && go test ./internal/module/assessment/application/commands -run 'ReportService' -count=1 -timeout 5m`
  - Review focus：application 是否已经不再知道 GORM sentinel；contest/report not-found 是否仍然稳定映射到原有错误码

- [x] Slice 2: 删除 allowlist 并同步文档
  - 目标：删除 `assessment/application/commands/report_service.go` 的 GORM allowlist，并更新 phase 5 当前事实
  - 验证：
    - `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus：只删除这条实际收口的 allowlist；文档准确描述 assessment 当前 concrete allowlist 状态

## Risks

- 如果 infrastructure 没有稳定映射 not-found，report service 会把 404 变成 500
- 如果 application 仍保留 GORM import，allowlist 不会真正收口

## Verification Plan

1. `cd code/backend && go test ./internal/module/assessment/application/commands -run 'ReportService' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
3. `python3 scripts/check-docs-consistency.py`
4. `bash scripts/check-consistency.sh`
5. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：report service 继续持有报表用例编排和错误码映射语义，persistence sentinel 映射落回 infrastructure
- reuse point 明确：复用 ops slice13 的 not-found contract 模式，不引入新的宽 repository facade
- 这刀同时解决行为与结构：保持 report not-found 用户行为不变，同时删掉 assessment command surface 的 concrete GORM import
