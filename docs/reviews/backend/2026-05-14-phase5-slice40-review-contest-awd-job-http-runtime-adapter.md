# Phase5 Slice40 Review

## Scope

- 目标：收口 `contest/application/jobs/awd_http_checker_request.go`、`awd_http_target_client.go`、`awd_probe_runtime.go`、`awd_round_runtime_bridge.go` 与 `awd_round_updater.go` 对 `net/http` 的直接依赖
- 评审对象：
  - `contest/ports/http_runtime.go`
  - `contest/infrastructure/awd_http_runtime_adapter.go`
  - `contest/infrastructure/awd_http_runtime_adapter_test.go`
  - `contest/application/jobs/awd_http_checker_request.go`
  - `contest/application/jobs/awd_http_target_client.go`
  - `contest/application/jobs/awd_probe_runtime.go`
  - `contest/application/jobs/awd_round_runtime_bridge.go`
  - `contest/application/jobs/awd_round_updater.go`
  - `contest/application/jobs/awd_http_runtime_contract_test.go`
  - `contest/application/jobs/awd_round_updater_test.go`
  - `contest/application/jobs/awd_testsupport_test.go`
  - `contest/runtime/module.go`
  - `architecture_allowlist_test.go`

## Independent Review

- Reviewer：`Code Reviewer`
- Classification check：`agree`
- Gate verdict：`pass`
- 结论：`no findings`

## Notes

- 这刀把 AWD jobs 的 HTTP concrete 正式收口到 `contest/infrastructure/awd_http_runtime_adapter.go`，`AWDRoundUpdater` 只保留 `ports.AWDHTTPRuntime` 依赖，runtime 只负责默认 adapter 注入。
- review 额外记录了一个非阻塞观察：`resolveAWDHTTPDialOverride` 目前在 `application/jobs` 与 `infrastructure` 各保留一份实现。当前 slice40 没有行为回退，但后续如果 alias 解析规则继续演进，需要留意两处 helper 的漂移风险。
- 这次 slice40 已经把 touched surface 上的 `httpClient` / `SetHTTPClient` / `httpClientForAWDTarget` boundary leak 一起收口，没有把同一块 debt 留到后续再拆。

## Verification

```bash
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -run 'TestAWDRoundUpdaterRunAWDHTTPCheckerActionUsesHTTPRuntime|TestAWDRoundUpdaterRunAWDHTTPCheckerActionMapsReadError|TestAWDRoundUpdaterProbeServiceInstanceUsesHTTPRuntime' -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -run 'TestAWDHTTPRuntimeAdapter' -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/jobs -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module -run 'TestApplicationConcreteDependencyAllowlistIsCurrent' -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh
```

## Review Result

- 当前 diff 已通过独立 correctness / regression review。
- slice40 可以提交，并作为 phase5 后续 challenge module allowlist 迁移的基础继续推进。
