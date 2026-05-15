# 2026-05-15 AWD until-contest-end runtime cleanup review

## Review target

- Repository: `ctf`
- Branch: `main`
- Diff source: local uncommitted changes around AWD instance expiry and contest-ended cleanup
- Files reviewed:
  - `code/backend/internal/module/practice/application/commands/instance_start_service.go`
  - `code/backend/internal/module/contest/infrastructure/ended_contest_runtime_cleaner.go`
  - `code/backend/internal/module/contest/infrastructure/status_side_effect_store.go`
  - `code/backend/internal/app/composition/contest_module.go`
  - `code/backend/internal/module/contest/runtime/module.go`
  - `code/backend/internal/module/runtime/infrastructure/repository.go`
  - related query / runtime owner code and targeted tests
- Classification check: agree with `HARNESS` + backend independent review
- Gate verdict: `blocked`

## Findings

### [Blocker] 比赛结束清理没有主动回收 AWD defense workspace companion container

- `code/backend/internal/module/contest/infrastructure/ended_contest_runtime_cleaner.go:70` 只把 `instances` 表里的运行时字段转成 `model.Instance` 后交给 `CleanupRuntime`。
- `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go:57` 只会从 `model.Instance` 的 `container_id` / `network_id` / `runtime_details` / `host_port` 提取资源。
- defense workspace companion container 单独存放在 `awd_defense_workspaces.container_id`，并不在实例 runtime 字段里：
  - `code/backend/internal/model/awd_defense_workspace.go:12`
  - `code/backend/internal/module/runtime/infrastructure/awd_target_proxy_repository.go:86`
  - `code/backend/internal/module/runtime/infrastructure/repository.go:818`
- 结果是：比赛结束 side effect 会清主实例 runtime，但不会清 companion workspace container，也不会把 `awd_defense_workspaces.status/container_id` 收口。这个容器最多只能等后续 orphan cleanup 再删，不满足“赛事结束立即清理 runtime”的新约束。
- 修正方向：把 defense workspace 行也纳入 ended cleanup owner，至少要在同一路径里删除 companion container，并把 workspace 状态改成终态、清空 `container_id`。

### [Blocker] 比赛结束后，活跃中的 AWD service operation 会卡在 `requested/provisioning/recovering`

- `code/backend/internal/module/contest/infrastructure/ended_contest_runtime_cleaner.go:70` 只做 runtime cleanup + `ExpireInstanceRuntime`，没有收口 `awd_service_operations`。
- 活跃状态定义在 `code/backend/internal/model/awd_service_operation.go:15`。
- 仓库层已经有专门的收口方法 `code/backend/internal/module/runtime/infrastructure/repository.go:508`，正常实例启动成功/失败路径也会调用它：
  - `code/backend/internal/module/practice/application/commands/instance_provisioning.go:26`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning.go:58`
- 查询层会继续读取每个 team/service 的最新 operation，而且 `GetUserWorkspace` 不会因为比赛 ended 而拒绝查询：
  - `code/backend/internal/module/contest/application/queries/awd_workspace_query.go:148`
  - `code/backend/internal/module/contest/infrastructure/awd_service_operation_repository.go:10`
  - `code/backend/internal/module/contest/application/queries/awd_support.go:13`
- 结果是：比赛结束后，实例已经被标成 `expired`，但 workspace 里仍可能显示一个永远不结束的 restart / provisioning / recovering 操作，状态和历史都不准确。
- 修正方向：ended cleanup 需要把实例上未完成的 AWD operation 统一收口到终态，并写明 `contest_ended` 之类的 reason / error message。

## Material findings

- `B1`: ended cleanup 必须覆盖 defense workspace container 与 workspace 状态收口。
- `B2`: ended cleanup 必须覆盖 active AWD operation 的终态收口。

## Senior implementation assessment

- 把 AWD 实例 `expires_at` 收口到 `contest.end_time`，方向是对的，比继续沿用固定 TTL 更符合比赛语义。
- 现在的问题不在 expiry 策略，而在 ended side effect 只收口了 `instances` 这一个表面。更稳妥的 owner 应该是“结束 AWD 运行态”这个单点，它一次性处理实例 runtime 字段、端口占用、defense workspace、active operation，而不是把这些不变量拆散到多个松耦合 helper 里。

## Validation observed

- `go test ./internal/module/contest/infrastructure -run 'TestContestEndedRuntimeCleanerCleansOnlyCurrentContestAWDInstances' -count=1 -timeout 5m`
  - 结果：通过
- `go test ./internal/module/practice/application/commands -run 'Test(StartContestAWDService|RestartContestAWDService)|TestServiceStartContestAWDServicePersistsServiceIDOnInstance' -count=1 -timeout 5m`
  - 结果：通过
- `go test ./internal/module/runtime/infrastructure -run 'Test(ExpireInstanceRuntime|UpdateStatusAndReleasePort)' -count=1 -timeout 5m`
  - 结果：通过
- `go test ./internal/module/practice/application/commands -count=1 -timeout 5m`
  - 结果：失败，卡在 `TestServiceStartContestAWDServiceResolvesServiceIDAndReusesTeamInstance`

## Required re-validation

- 增加 ended cleanup 测试，覆盖：
  - defense workspace companion container 被清理，`awd_defense_workspaces.status/container_id` 被收口
  - `requested/provisioning/recovering` operation 被收口到终态
  - 其他 contest / team / service 的行不受影响
- 修复后至少重跑：
  - `go test ./internal/module/contest/infrastructure -count=1 -timeout 5m`
  - `go test ./internal/module/contest/... -count=1 -timeout 5m`
  - `go test ./internal/module/runtime/infrastructure -run 'Test(ExpireInstanceRuntime|UpdateStatusAndReleasePort)' -count=1 -timeout 5m`
  - `go test ./internal/module/practice/application/commands -run 'Test(StartContestAWDService|RestartContestAWDService)|TestServiceStartContestAWDServicePersistsServiceIDOnInstance' -count=1 -timeout 5m`

## Residual risk

- 本次 review 只收敛在 AWD 生命周期这组改动；仓库工作区还有大量无关脏改动，没有并入审查结论。
- `go test ./internal/module/practice/application/commands -count=1 -timeout 5m` 当前不绿，至少说明这个 package 还存在独立问题或测试 / 实现不一致，需要在合并前单独澄清。
- Touched known-debt status: 本次触达的是 runtime lifecycle surface。当前 blocker 是这轮变更新增的收口缺口，不是 `docs/reviews/backend/README.md` 里已标记完成的旧 runtime safety debt 回流。
