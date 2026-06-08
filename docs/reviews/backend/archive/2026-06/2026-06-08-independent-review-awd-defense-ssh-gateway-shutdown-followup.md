# AWD Defense SSH Gateway Shutdown Follow-up Independent Review

## Review Target

- Repository: `ctf`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-07-awd-defense-ssh-gateway-split`
- Branch: `task/2026-06-07-awd-defense-ssh-gateway-split`
- Task slug: `2026-06-07-awd-defense-ssh-gateway-split`
- Plan: `docs/plan/impl-plan/2026-06-07-awd-defense-ssh-gateway-split-implementation-plan.md`
- Diff source: current uncommitted diff in this worktree
- Files reviewed:
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go`
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway_test.go`
  - `code/backend/internal/app/router_composition_structure_test.go`
  - `code/backend/scripts/docker-entrypoint.sh`

## Classification Check

同意 `非琐碎任务` / 独立 gate review 判定。

## Gate Verdict

`blocked`

## Findings

### 1. Blocking: `Shutdown()` 在 `gateway.Stop()` 超时或报错后仍继续关闭 runtime / DB / Redis，session 未收敛时仍有资源关闭竞态

- 位置：
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go:123-141`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go:127-169`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go:496-512`
- 行为：
  - `gateway.Stop(ctx)` 现在会取消 `runCtx`、关闭活跃连接，并在 `ctx` 截止前等待 `workerWG`。
  - 但如果 interactive exec 没能在 shutdown timeout 内退出，`waitWaitGroup()` 会返回 `ctx.Err()`。
  - `awdDefenseSSHGatewayProcess.Shutdown()` 记录这个错误后，仍然无条件继续执行 `runtimeCloser.Close(ctx)` 和 `closeResources(...)`。
- 触发条件：
  - 活跃 SSH interactive session 对取消响应慢、底层 runtime exec 卡住，或 stop path 因任何原因返回错误。
- 影响：
  - `runtimeCloser`、Postgres、Redis 仍可能在 session / exec 尚未退出时被关闭。
  - 这正是本次修复要避免的 shutdown 竞态，当前只把“正常取消路径”做对了，没有把“超时失败路径”收口。
- 修正方向：
  - `Shutdown()` 必须在 `gateway.Stop(ctx)` 成功收敛前阻止 `runtimeCloser.Close()` 和 `closeResources()` 继续执行，或者明确分成“先强制终止 session，再关资源”的两阶段停机。
  - 修复后需要补失败路径测试，证明 stop 超时或 stop error 时不会提前关闭 runtime / DB / Redis。

## Material Findings

- `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go`
  - 必须修正 `Shutdown()` 的错误路径，避免在 `gateway.Stop()` 未完成收敛时继续关闭 runtime 与存储资源。
  - 需要新增自动化验证覆盖 `gateway.Stop()` 返回错误或超时的场景。

## Non-blocking Suggestions

- `code/backend/scripts/docker-entrypoint.sh:48-56`
  - 当前只有“零参数”时才补 `/app/ctf-api`。如果外部仍沿用旧的 args-only 启动方式，例如 `docker run image --flag`，现在会直接 `exec --flag` 失败。
  - 仓库内没有直接证据表明这已经是线上 blocker，但建议补兼容或在运维文档里明确废弃这条旧契约。

## Missing Validation

- `code/backend/internal/bootstrap/awd_defense_ssh_gateway_test.go:60-130`
  - 现有 bootstrap 测试只验证了正常 shutdown 会调用 `gateway.Stop()`、`runtimeCloser.Close()` 并关闭 DB/Redis，没有覆盖 `gateway.Stop()` 返回错误或超时时的资源关闭顺序。
- `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go:342-388`
  - 现有组合测试只证明 `gateway.Stop()` 会取消一个手工构造的 interactive exec，没有覆盖真实 `process.Shutdown()` 的 timeout/error 传播。

## Senior Implementation Assessment

这次拆分的大方向是对的：

- `InstanceModule` 已经不再注册 SSH gateway background job。
- `runtimeNodeExecutionRouter` 也补上了 interactive exec 的 `container_id -> node_id` 路由。
- `AWDDefenseSSHGateway.Stop()` 本身已经开始主动取消 `runCtx`、关闭活跃连接并等待 worker 收敛，这部分比原实现完整。

但停机链路还差最后一道门：只要 `gateway.Stop()` 没能在超时内收敛，bootstrap 仍会继续把 runtime / DB / Redis 关掉。对一个持有长连接和 interactive stream 的独立网关来说，这个失败路径不能留成 residual risk。

## Required Re-validation

修复后至少重新执行：

```bash
cd code/backend
go test ./internal/app/composition -run 'TestRuntimeNodeExecutionRouter|TestAWDDefenseSSHGateway|TestBuildAWDDefenseSSHGateway' -count=1
go test ./internal/bootstrap -run 'TestRunAWDDefenseSSHGateway|TestShutdownGracefully' -count=1
go test ./internal/app -run 'TestBuildInstanceModule|TestBuildAWDDefenseSSHGateway' -count=1
```

并新增一条失败路径验证，证明：

- `gateway.Stop()` 超时或返回错误时，`runtimeCloser.Close()` 不会先执行
- DB / Redis 不会在活跃 session 尚未退出时被关闭

## Residual Risk

- 我没有复跑真实 SSH 会话联调，只复核了 Go 代码与相关单测。
- `docker-entrypoint.sh` 的 args-only 兼容性仍缺少 shell 级验证证据。

## Touched Known-debt Status

- 本次复核没有命中已登记的必须顺手收口的结构性旧债。
- 当前 blocker 是这次 shutdown 修复本身仍未收口的失败路径，不是历史遗留问题。
