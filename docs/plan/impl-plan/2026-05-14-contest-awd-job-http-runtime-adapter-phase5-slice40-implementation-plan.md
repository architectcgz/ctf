# Contest AWD Job HTTP Runtime Adapter Phase 5 Slice 40 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `contest/application/jobs/awd_http_checker_request.go`、`awd_http_target_client.go`、`awd_probe_runtime.go`、`awd_round_runtime_bridge.go` 与 `awd_round_updater.go` 对 `net/http` 的直接依赖，并把 AWD jobs 的 HTTP runtime concrete 收口到 infrastructure adapter。

**Architecture:** 在 `contest/ports` 定义 job-only HTTP execution port，在 `contest/infrastructure` 增加默认 HTTP runtime adapter；`AWDRoundUpdater` 只传请求参数、读取响应结果和 dial-override 语义，不再创建 `http.Request` / `http.Client` 或暴露 `SetHTTPClient(*http.Client)`。

**Tech Stack:** Go, net/http, modular monolith ports/infrastructure, job contract tests

---

## Objective

- 删除 `contest/application/jobs/awd_http_checker_request.go -> net/http`
- 删除 `contest/application/jobs/awd_http_target_client.go -> net/http`
- 删除 `contest/application/jobs/awd_probe_runtime.go -> net/http`
- 删除 `contest/application/jobs/awd_round_runtime_bridge.go -> net/http`
- 删除 `contest/application/jobs/awd_round_updater.go -> net/http`

## Non-goals

- 不修改 AWD script / TCP checker 路径
- 不修改 challenge module 剩余 allowlist
- 不改业务状态判断与 probe 结果语义

## Inputs

- `.harness/reuse-decisions/contest-awd-job-http-runtime-adapter-phase5-slice40.md`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_request.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_target_client.go`
- `code/backend/internal/module/contest/application/jobs/awd_probe_runtime.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_runtime_bridge.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_updater_test.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_sandbox.go`
- `code/backend/internal/module/contest/runtime/module.go`

## Ownership Boundary

- `contest/application/jobs/awd_http_checker_request.go`
  - 负责：构造 checker 行为所需的语义化请求参数
  - 不负责：直接创建 `http.Request` / 调 `client.Do`
- `contest/application/jobs/awd_probe_runtime.go`
  - 负责：解释 probe 结果
  - 不负责：直接创建 `http.Request` / 调 `client.Do`
- `contest/application/jobs/awd_round_updater.go`
  - 负责：依赖 HTTP execution port
  - 不负责：持有 `*http.Client`
- `contest/infrastructure/awd_http_runtime_adapter.go`
  - 负责：默认 HTTP client、request 构造、dial override、transport clone 和测试替身注入
- `contest/runtime/module.go`
  - 负责：只给 `AWDRoundUpdater` 注入默认 HTTP runtime adapter

## Change Surface

- Add: `.harness/reuse-decisions/contest-awd-job-http-runtime-adapter-phase5-slice40.md`
- Add: `docs/plan/impl-plan/2026-05-14-contest-awd-job-http-runtime-adapter-phase5-slice40-implementation-plan.md`
- Add: `code/backend/internal/module/contest/ports/http_runtime.go`
- Add: `code/backend/internal/module/contest/infrastructure/awd_http_runtime_adapter.go`
- Add: `code/backend/internal/module/contest/infrastructure/awd_http_runtime_adapter_test.go`
- Add: `code/backend/internal/module/contest/application/jobs/awd_http_runtime_contract_test.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_http_checker_request.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_http_target_client.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_probe_runtime.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_runtime_bridge.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_updater_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`

## Task 1: 先写 contract tests

**Files:**

- Add: `code/backend/internal/module/contest/application/jobs/awd_http_runtime_contract_test.go`
- Add: `code/backend/internal/module/contest/infrastructure/awd_http_runtime_adapter_test.go`

- [ ] 为 HTTP checker action 走 port 而不是直接 `net/http` 补 contract test
- [ ] 为 service probe 走 port 而不是直接 `net/http` 补 contract test
- [ ] 为 HTTP runtime adapter 的 dial override / response 行为补 adapter tests
- [ ] 为测试替身注入路径补 contract test，替代现有 `SetHTTPClient(*http.Client)` 语义

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -run 'AWDRoundUpdater.*HTTP|AWD.*Probe' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -run 'AWDHTTPRuntimeAdapter' -count=1 -timeout 300s`

Review focus：

- application/jobs tests 是否真正约束 “只消费 port”，而不是仍然间接依赖 `*http.Client`
- adapter tests 是否覆盖 dial override 和默认 timeout 行为

## Task 2: 实现 port、adapter 与 runtime wiring

**Files:**

- Add: `code/backend/internal/module/contest/ports/http_runtime.go`
- Add: `code/backend/internal/module/contest/infrastructure/awd_http_runtime_adapter.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_http_checker_request.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_http_target_client.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_probe_runtime.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_runtime_bridge.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_updater_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`

- [ ] 设计并落地 AWD jobs 使用的窄 HTTP execution port
- [ ] `AWDRoundUpdater` 改持有 port，不再持有 `*http.Client`
- [ ] `SetHTTPClient` 改成面向 port 的测试注入方式，或替换为等价 test-only adapter hook
- [ ] 默认 HTTP runtime adapter 收口 request 构造、transport clone 和 dial override
- [ ] runtime 只改 `AWDRoundUpdater` 的 wiring
- [ ] allowlist 删掉这 5 条 `net/http` 依赖

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module -run 'TestApplicationConcreteDependencyAllowlistIsCurrent' -count=1 -timeout 300s`

Review focus：

- `AWDRoundUpdater` 是否已完全脱离 `net/http` concrete
- `SetHTTPClient` 相关测试钩子是否被等价替换，没有让 application 留下 concrete test seam
- runtime 是否只改了 `AWDRoundUpdater` 的 HTTP adapter 注入

## Risks

- 现有 tests 大量通过 `SetHTTPClient(server.Client())` 注入行为，如果迁移方式不稳，会出现大面积 test churn
- dial override 语义集中在 `httpClientForAWDTarget()`，如果只迁走请求发送、不迁走 transport 选择，application 仍会残留 `net/http`
- sandbox 路径和直接 HTTP 路径并存，port 设计要避免把 script / sandbox checker 一起耦合进去

## Verification Plan

1. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -count=1 -timeout 300s`
2. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -count=1 -timeout 300s`
3. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 300s`
4. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module -run 'TestApplicationConcreteDependencyAllowlistIsCurrent' -count=1 -timeout 300s`

## Architecture-Fit Evaluation

- owner 明确：HTTP concrete 由 `contest/infrastructure/awd_http_runtime_adapter.go` 收口，jobs application 只处理 probe/checker 语义
- reuse point 明确：不复用过宽的 runtime / checker runner 接口，而是定义 AWD jobs 自己消费的窄 HTTP port
- 结构收敛明确：这刀同时清理 `AWDRoundUpdater.httpClient`、`httpClientForAWDTarget()` 和 `SetHTTPClient()` 这同一块 boundary leak，而不是只删 allowlist
