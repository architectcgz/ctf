# Reuse Decision

## Change type
job / service

## Existing code searched
- code/backend/internal/module/contest/application/commands/awd_service_run_commands.go
- code/backend/internal/module/contest/application/jobs/awd_checker_preview.go
- code/backend/internal/module/contest/application/jobs/awd_http_checker_request.go
- code/backend/internal/module/contest/application/jobs/awd_probe_runtime.go
- code/backend/internal/module/contest/application/jobs/awd_http_runtime_contract_test.go

## Similar implementations found
- `code/backend/internal/module/contest/application/jobs/awd_probe_runtime.go`
  - 已有基于 `healthPath` 的 HTTP 可用性探测，可复用为 preview 前置就绪等待的单次探测原语。
- `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
  - 已有 preview 三轮聚合与进度上报 owner，不应把运行时就绪细节塞回 command 层。

## Decision
extend_existing

## Reason
问题发生在 `http_standard` preview 刚拉起临时实例后立刻执行 `put_flag`，并不是题包 checker 配置缺失，也不是 command 层聚合逻辑错误。最小正确改动是在 `AWDRoundUpdater` 的 preview HTTP checker 执行路径中复用现有 `/health` 探测语义，增加短暂就绪等待，而不是改题包、改前端或把重试散落到更上层。

## Files to modify
- code/backend/internal/module/contest/application/jobs/awd_checker_preview.go
- code/backend/internal/module/contest/application/jobs/awd_http_runtime_contract_test.go
- code/backend/internal/module/contest/application/jobs/awd_probe_runtime.go
- code/backend/internal/module/contest/application/jobs/awd_http_checker_request.go

## After implementation
- 如果这次修复沉淀出稳定的 preview readiness 模式，再评估是否补到 `.harness/reuse-index/`。
