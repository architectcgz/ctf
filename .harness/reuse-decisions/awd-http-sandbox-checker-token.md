# Reuse Decision

## Change type

job / runtime / tests / docs

## Existing code searched

- `code/backend/internal/module/contest/application/jobs/awd_http_checker_request.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_sandbox.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_sandbox_test.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_template.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_runtime_contract_test.go`

## Similar implementations found

- `awd_http_checker_request.go` 的非 sandbox 路径已经先渲染 headers，再发起 HTTP runtime 请求
- `awd_http_checker_sandbox.go` 已经复用相同的 `bodyValue` 渲染结果，但 `headers` 仍然直接序列化原始模板
- 比赛 35 的现场验证说明 alias target 走的是 sandbox 分支，而不是普通 HTTP runtime 分支

## Decision

extend_existing

## Reason

这次不是新增 sandbox checker 能力，而是让 sandbox 分支和普通 HTTP runtime 分支消费同一份“已渲染 headers”语义。最小正确方案是在 `runAWDHTTPCheckerAction` 里先统一渲染 headers，再把这份结果同时传给 sandbox 和普通 HTTP runtime，避免两条分支继续各自维护一份 header 构造逻辑。

## Files to modify

- `code/backend/internal/module/contest/application/jobs/awd_http_checker_request.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_sandbox.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_sandbox_test.go`
- `docs/plan/impl-plan/2026-05-15-awd-http-sandbox-checker-token-implementation-plan.md`

## After implementation

- `http_standard` 的 sandbox / non-sandbox 两条执行路径都应使用同一份已渲染 header 数据
- 后续如果 HTTP checker 模板变量继续扩展，优先让公共渲染结果下沉到两条分支共享，而不是在 sandbox 内重复做模板替换
