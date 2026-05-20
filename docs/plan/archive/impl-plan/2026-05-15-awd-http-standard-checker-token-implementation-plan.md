# AWD HTTP Standard Checker Token Implementation Plan

**Goal:** 修复 `http_standard` 对 `{{CHECKER_TOKEN}}` 的支持缺口，让 runtime check 和 preview check 都能在 header / body / expected substring 中拿到真实 checker token。

**Architecture:** 沿用现有 `resolveAWDCheckerToken` 能力，不新增新的 checker 渲染器；只扩展 `awdHTTPCheckerTemplateData`、HTTP runner 和 preview builder。

**Tech Stack:** Go, AWD checker jobs, Go contract tests

---

## Objective

- 为 HTTP checker 模板渲染补上 `{{CHECKER_TOKEN}}`
- 为 runtime HTTP checker 补上 checker token 解析与透传
- 为 preview HTTP checker 补上显式 checker token 透传
- 用最小 contract tests 锁定 runtime / preview 两条链路

## Non-goals

- 不修改 `{{FLAG}} / {{ROUND}} / {{TEAM_ID}} / {{AWD_CHALLENGE_ID}} / {{CHALLENGE_ID}}` 现有语义
- 不改 `script_checker` 或 `tcp_standard` 的现有实现
- 不调整 AWD checker token 生成算法

## Inputs

- `.harness/reuse-decisions/awd-http-standard-checker-token.md`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_template.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_runner.go`
- `code/backend/internal/module/contest/application/jobs/awd_checker_preview.go`
- `code/backend/internal/module/contest/application/jobs/awd_checker_token_support.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_runtime_contract_test.go`
- `docs/plan/impl-plan/2026-05-07-awd-checker-tokenized-management-implementation-plan.md`

## Ownership Boundary

- `awd_http_checker_template.go`
  - 负责：HTTP checker 占位符渲染
  - 不负责：生成 checker token
- `awd_http_checker_runner.go`
  - 负责：runtime HTTP checker 所需 flag / checker token 的解析与透传
  - 不负责：改变 token 生成算法
- `awd_checker_preview.go`
  - 负责：preview HTTP checker 的显式 token 透传
  - 不负责：派生 preview token
- `awd_http_runtime_contract_test.go`
  - 负责：覆盖 token 的 runtime / preview 渲染契约

## Change Surface

- Add: `.harness/reuse-decisions/awd-http-standard-checker-token.md`
- Add: `docs/plan/impl-plan/2026-05-15-awd-http-standard-checker-token-implementation-plan.md`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_http_checker_template.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_http_checker_runner.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_checker_preview.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_http_runtime_contract_test.go`

## Task 1: 先写失败测试

- [x] 为 HTTP preview checker 增加 `{{CHECKER_TOKEN}}` header 渲染测试
- [x] 为 HTTP runtime checker 增加派生 checker token 后再渲染 header 的测试
- [x] 运行定向 `go test`，确认新增测试先失败

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -run 'TestAWDRoundUpdater(PreviewHTTPStandardUsesCheckerToken|HTTPStandardDerivesCheckerTokenForRuntimeChecks)' -count=1 -timeout 300s`

Review focus：

- 测试是否覆盖了 runtime token 派生和 preview token 透传两种来源
- 失败点是否确实来自 token 缺失，而不是 HTTP stub 或 fixture 问题

## Task 2: 最小实现修复

- [x] 扩展 `awdHTTPCheckerTemplateData` 支持 `CheckerToken`
- [x] 在 HTTP template 渲染器里补上 `{{CHECKER_TOKEN}}`
- [x] 在 runtime HTTP checker builder 中按需解析 checker token 并传入 target runtime
- [x] 在 preview HTTP checker builder 中传入显式 checker token

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -run 'TestAWDRoundUpdater(PreviewHTTPStandardUsesCheckerToken|HTTPStandardDerivesCheckerTokenForRuntimeChecks|RunAWDHTTPCheckerAction|ProbeServiceInstance)' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -count=1 -timeout 300s`

Review focus：

- 不要把 HTTP checker 模板体系改成只剩 `{{CHECKER_TOKEN}}`
- 不要让未使用 checker token 的 HTTP checker 因为缺 secret 而误报失败

## Risks

- 如果 runtime builder 无条件解析 token，会让未使用 `{{CHECKER_TOKEN}}` 的 HTTP checker 在 secret 缺失时也报错
- 如果只补模板渲染、不补 builder 透传，header 仍会保留字面量占位符
- preview path 和 runtime path 的 token 来源不同，容易只修一半

## Verification Plan

1. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -run 'TestAWDRoundUpdater(PreviewHTTPStandardUsesCheckerToken|HTTPStandardDerivesCheckerTokenForRuntimeChecks)' -count=1 -timeout 300s`
2. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -count=1 -timeout 300s`

## Architecture-Fit Evaluation

- owner 明确：token 仍由 `resolveAWDCheckerToken` 生成，HTTP checker 只消费结果
- reuse point 明确：复用现有 HTTP 模板数据结构和 script / tcp 的 token 解析模式，不新建渲染器
- 结构收敛明确：同一类 checker token 能力在三种 AWD checker 中保持一致，而不是继续让 `http_standard` 维持特例
