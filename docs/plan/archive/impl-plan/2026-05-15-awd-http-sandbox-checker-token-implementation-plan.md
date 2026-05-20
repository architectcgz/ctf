# AWD HTTP Sandbox Checker Token Implementation Plan

**Goal:** 修复 `http_standard` 在 sandbox 网络路径下未渲染 `headers` 模板的问题，确保 `{{CHECKER_TOKEN}}` 在 `awd-c...` 这类 alias target 场景也能正确下发。

**Architecture:** 复用现有 HTTP checker 模板渲染结果，不在 sandbox 脚本内再次解析模板；由 `runAWDHTTPCheckerAction` 统一生成 rendered headers，并同时供 sandbox / 非 sandbox 分支使用。

**Tech Stack:** Go, AWD checker jobs, Go unit tests, live AWD contest regression

---

## Objective

- 为 sandbox HTTP checker 增加 header 渲染回归测试
- 让 sandbox 分支消费已渲染的 headers，而不是原始模板
- 用比赛 35 的当前轮次手动重跑结果验证现场故障修复

## Non-goals

- 不修改 checker token 算法
- 不调整 HTTP sandbox 脚本的协议字段
- 不改 AWD checker readiness / 审计流程

## Inputs

- `.harness/reuse-decisions/awd-http-sandbox-checker-token.md`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_request.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_sandbox.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_sandbox_test.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_template.go`

## Ownership Boundary

- `awd_http_checker_request.go`
  - 负责：统一构造 rendered headers / body，并分发给 sandbox / 非 sandbox 执行路径
  - 不负责：sandbox 脚本 HTTP 细节
- `awd_http_checker_sandbox.go`
  - 负责：接收已渲染 headers，并通过 sandbox 脚本发起请求
  - 不负责：自己再做模板替换
- `awd_http_checker_sandbox_test.go`
  - 负责：锁定 sandbox env 中的 `AWD_HTTP_HEADERS` 已经是渲染后的结果

## Change Surface

- Add: `.harness/reuse-decisions/awd-http-sandbox-checker-token.md`
- Add: `docs/plan/impl-plan/2026-05-15-awd-http-sandbox-checker-token-implementation-plan.md`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_http_checker_request.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_http_checker_sandbox.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_http_checker_sandbox_test.go`

## Task 1: 先写失败测试

- [x] 为 alias target 的 sandbox checker 增加 `{{CHECKER_TOKEN}}` header 渲染测试
- [x] 运行定向 `go test`，确认测试先失败

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -run 'TestAWDHTTPStandardSandboxRendersCheckerTokenHeaders' -count=1 -timeout 300s`

Review focus：

- 测试是否真的走了 sandbox 分支，而不是普通 HTTP runtime stub
- 断言是否直接检查 `AWD_HTTP_HEADERS`，避免只看最终健康状态

## Task 2: 最小实现修复

- [x] 在 `runAWDHTTPCheckerAction` 里统一渲染 headers
- [x] sandbox 分支改消费 rendered headers
- [x] 非 sandbox 分支继续复用同一份 rendered headers

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -run 'TestAWDHTTPStandardSandboxRendersCheckerTokenHeaders|TestAWDRoundUpdater(PreviewHTTPStandardUsesCheckerToken|HTTPStandardDerivesCheckerTokenForRuntimeChecks|RunAWDHTTPCheckerAction|ProbeServiceInstance)' -count=1 -timeout 300s`
- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -count=1 -timeout 300s`

Review focus：

- 不要让 sandbox / 非 sandbox 再次出现两套 header 处理逻辑
- 不要改变已有 sandbox network selection 行为

## Task 3: 现场回归比赛 35

- [x] 触发 `POST /api/v1/admin/contests/35/awd/current-round/check`
- [x] 回查 `awd_team_services` 当前轮次状态
- [x] 回查题目容器日志，确认 `PUT /api/flag` 不再因为 token 缺失返回 404

验证：

- `curl -sS http://127.0.0.1:8080/health`
- 管理员登录后手动触发当前轮次检查
- `docker exec ctf-postgres psql -U postgres -d ctf -c "select ... from awd_team_services ..."`
- `docker logs --since 3m ctf-instance-challenge-c35-t68-s30`

## Risks

- 如果只改 sandbox 分支而不让普通分支共用同一份 rendered headers，后续还会再漂移
- 现场验证依赖本地后端热重载成功；如果运行进程没吃到新代码，数据库状态仍会误导

## Verification Plan

1. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -run 'TestAWDHTTPStandardSandboxRendersCheckerTokenHeaders' -count=1 -timeout 300s`
2. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -count=1 -timeout 300s`
3. `curl -sS http://127.0.0.1:8080/health`
4. 管理员触发比赛 35 当前轮次检查并回查数据库 / 容器日志

## Architecture-Fit Evaluation

- owner 明确：template 渲染仍在 request builder，sandbox 只消费结果
- reuse point 明确：继续复用现有 `renderAWDHTTPCheckerTemplate` 和 header map，而不是给 sandbox 单独造模板逻辑
- 结构收敛明确：HTTP checker 的 header 渲染 owner 收口到单点，sandbox 只是执行介质

## Execution Notes

- 失败用例先暴露出 sandbox env 中 `AWD_HTTP_HEADERS` 仍保留字面量 `{{CHECKER_TOKEN}}`
- 手动管理员重检接口在客户端侧超时后记录了 `context canceled`，但数据库与题目容器日志确认检查实际已执行完成
- 比赛 `35` 当前轮次最终状态恢复为 team `68/69/70` 全部 `up`，`put_flag` / `get_flag` 均返回 `200`
