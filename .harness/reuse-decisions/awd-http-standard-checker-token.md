# Reuse Decision

## Change type

job / runtime / tests / docs

## Existing code searched

- `code/backend/internal/module/contest/application/jobs/awd_http_checker_template.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_runner.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_request.go`
- `code/backend/internal/module/contest/application/jobs/awd_checker_preview.go`
- `code/backend/internal/module/contest/application/jobs/awd_checker_token_support.go`
- `code/backend/internal/module/contest/application/jobs/awd_script_checker_runner.go`
- `code/backend/internal/module/contest/application/jobs/awd_tcp_checker_runner.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_runtime_contract_test.go`

## Similar implementations found

- `awd_script_checker_runner.go` 已经在 runtime check 和 preview check 两条链路中解析并传递 `checkerToken`
- `awd_tcp_checker_runner.go` 复用了同样的 token 解析模式，并在 runtime tests 中覆盖了派生 token 的行为
- `awd_http_checker_request.go` 已经统一通过 `awdHTTPCheckerTemplateData` 渲染 header / body / expected substring，适合直接扩字段而不是新建第二套 HTTP checker 渲染逻辑

## Decision

extend_existing

## Reason

当前问题不是缺少新的 checker 体系，而是 `http_standard` 这条既有链路落后于题包契约：模板渲染器没支持 `{{CHECKER_TOKEN}}`，runtime builder 也没有像 script / tcp 那样解析并透传 token。最小正确方案是在现有 HTTP checker 模板数据和 runner 上补齐 token 支持，并沿用已有的 token 解析 helper 与测试风格。

## Files to modify

- `code/backend/internal/module/contest/application/jobs/awd_http_checker_template.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_runner.go`
- `code/backend/internal/module/contest/application/jobs/awd_checker_preview.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_runtime_contract_test.go`
- `docs/plan/impl-plan/2026-05-15-awd-http-standard-checker-token-implementation-plan.md`

## After implementation

- `http_standard` 与 `script_checker` / `tcp_standard` 的 checker token 契约应保持一致
- 若后续还有新的 checker 类型消费模板占位符，应优先复用现有 token 解析 helper，而不是在各自链路里重新拼接 token
