# AWD Defense SSH Gateway Shutdown Re-review

## Review Target

- Repository: `ctf`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-07-awd-defense-ssh-gateway-split`
- Branch: `task/2026-06-07-awd-defense-ssh-gateway-split`
- Task slug: `2026-06-07-awd-defense-ssh-gateway-split`
- Focus: previous shutdown blocker follow-up
- Files reviewed:
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go`
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway_test.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`

## Classification Check

同意继续按 `非琐碎任务` 的独立 gate review 处理。

## Gate Verdict

`pass`

## Findings

- 无 blocker。上一轮 blocker 已解除：
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go:123-142` 现在在 `gateway.Stop(ctx)` 返回错误时立刻返回，不再继续执行 `runtimeCloser.Close(ctx)` 和 `closeResources(...)`。
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway_test.go:132-179` 新增测试覆盖了失败路径，断言 runtime closer 被跳过，且 DB / Redis 保持打开。

## Non-blocking Suggestions

- `code/backend/scripts/docker-entrypoint.sh:48-56`
  - args-only 启动方式的兼容性风险仍在，但这不是本轮 shutdown blocker 的残留问题。

## Required Re-validation

本次独立复跑：

```bash
cd code/backend
go test ./internal/bootstrap -run 'TestRunAWDDefenseSSHGateway|TestShutdownGracefully' -count=1
go test ./internal/app/composition -run 'TestAWDDefenseSSHGateway|TestBuildAWDDefenseSSHGateway' -count=1
```

结果：均 `PASS`。

## Residual Risk

- 本次只复核 shutdown blocker 相关路径，没有重新审全量 compose / entrypoint / docs 变更。

## Touched Known-debt Status

- 上一轮 review 标出的 shutdown 失败路径 blocker 已在当前 touched surface 内收口。
