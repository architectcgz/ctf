# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/contest/application/jobs/awd_http_checker_request.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_target_client.go`
- `code/backend/internal/module/contest/application/jobs/awd_probe_runtime.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_runtime_bridge.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_updater_test.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_sandbox.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/checker_runner.go`

## Similar implementations found

- `code/backend/internal/module/contest/infrastructure/docker_checker_runner.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_sandbox.go`

## Decision

refactor_existing

## Reason

这次 debt 不是 AWD jobs 需要新的 checker 能力，而是 `AWDRoundUpdater` 还把 HTTP 请求构造、client 选择、dial override 和测试注入钩子都留在 application/jobs。最小正确做法是：

- 在 `contest/ports` 增加 job 使用的窄 HTTP 执行 port
- 在 `contest/infrastructure` 增加默认 HTTP adapter，把 `http.NewRequestWithContext`、`*http.Client`、transport clone、dial override 收口出去
- `AWDRoundUpdater` 只传语义化请求参数和读取响应结果
- runtime 只给 `AWDRoundUpdater` 注入默认 HTTP adapter

这样可以把 touched surface 上的结构债一并收口：当前 `AWDRoundUpdater.httpClient`、`httpClientForAWDTarget()`、`SetHTTPClient()` 是同一块 boundary leak，如果只替换一半，application 层仍会残留 `net/http` concrete。

## Files to modify

- `.harness/reuse-decisions/contest-awd-job-http-runtime-adapter-phase5-slice40.md`
- `docs/plan/impl-plan/2026-05-14-contest-awd-job-http-runtime-adapter-phase5-slice40-implementation-plan.md`
- `code/backend/internal/module/contest/ports/http_runtime.go`
- `code/backend/internal/module/contest/infrastructure/awd_http_runtime_adapter.go`
- `code/backend/internal/module/contest/infrastructure/awd_http_runtime_adapter_test.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_request.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_target_client.go`
- `code/backend/internal/module/contest/application/jobs/awd_probe_runtime.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_runtime_bridge.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_runtime_contract_test.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_updater_test.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`

## After implementation

- `awd_http_checker_request.go -> net/http` 收口到 HTTP runtime adapter
- `awd_http_target_client.go -> net/http` 下沉到 infrastructure adapter
- `awd_probe_runtime.go -> net/http` 收口到 HTTP runtime adapter
- `awd_round_runtime_bridge.go -> net/http` 改成注入 port / test double，不再暴露 `*http.Client`
- `awd_round_updater.go -> net/http` 收口到 port 字段，不再持有 concrete client
- `contest/runtime/module.go` 只在 `AWDRoundUpdater` wiring 上注入默认 HTTP adapter
