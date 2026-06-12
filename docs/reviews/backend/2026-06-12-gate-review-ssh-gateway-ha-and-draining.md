# SSH Gateway HA And Draining Gate Re-review

日期：2026-06-12

## Review Target

- Repository: `/home/azhi/workspace/projects/ctf`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-12-ssh-gateway-ha-and-draining`
- Branch: `task/2026-06-12-ssh-gateway-ha-and-draining`
- Task: `2026-06-12-ssh-gateway-ha-and-draining`
- Plan: `docs/plan/archive/impl-plan/2026-06/2026-06-12-ssh-gateway-ha-and-draining-implementation-plan.md`
- Diff source: 当前 worktree 未提交 diff（含 untracked `code/backend/internal/module/instance/infrastructure/proxy_ticket_store_test.go`）
- Files reviewed:
  - `README.md`
  - `code/backend/configs/config.prod.yaml`
  - `code/backend/configs/config.yaml`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
  - `code/backend/internal/app/composition/runtime_module_test.go`
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go`
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway_test.go`
  - `code/backend/internal/module/instance/infrastructure/proxy_ticket_store_test.go`
  - `docker/docker-compose.dev.yml`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
  - `docs/operations/runtime-agent-deployment.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-ssh-gateway-ha-and-draining-implementation-plan.md`

## Classification Check

同意当前任务按 `非琐碎任务` 处理。该 diff 同时触达 SSH gateway 生命周期、bootstrap shutdown 语义、跨副本 ticket 校验、host key 合约和运维事实源，属于需要独立 gate 的后端行为变更。

## Gate Verdict

Pass with minor issues.

Review archive path: `docs/reviews/backend/2026-06-12-gate-review-ssh-gateway-ha-and-draining.md`

## Findings

- 无 material findings。上一轮独立 review 的两条 blocker 已收口：
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go` 现在会在 `Drain()` 返回错误后继续执行 `cancel + Stop()`，并仅在 `Stop()` 失败时阻止后续 runtime / DB / Redis 关闭；对应 `code/backend/internal/bootstrap/awd_defense_ssh_gateway_test.go` 已更新为验证这条 hard-stop 语义。
  - `docker/docker-compose.dev.yml` 现在通过一次性的 `ctf-awd-defense-ssh-host-key` service 预置共享 host key，消除了默认 dev / compose 路径在 `load-only` 合约下的 fresh 启动回归；README 与运维文档也同步明确了该 owner。

## Material Findings

- 无。当前独立 re-review 未发现需要阻塞 completion 的问题。

## Non-blocking Suggestions

- `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
  - `Ready()` / `State()` 目前仍是进程内状态，当前需求下足够支撑 TCP LB 摘流，但如果后续要接入非 TCP 健康探测或统一 health service，最好显式补一个 owner，而不是继续只靠 getter 留在组合层。

## Required Re-validation

本轮独立 re-review 复用并确认了以下验证证据：

```bash
cd code/backend && go test ./internal/bootstrap -run 'TestRunAWDDefenseSSHGatewayProcessShutdown' -count=1
cd code/backend && go test ./internal/app/composition -run 'TestAWDDefenseSSHGateway(ReadyStateTracksDrainAndStop|DrainStopsNewTCPConnections)|TestLoadAWDDefenseSSHHostKeySigner(RequiresExistingSharedFile|UsesSharedFileAcrossGateways)|TestAWDDefenseSSHGatewayAuthenticateAcceptsTicketIssuedByAnotherReplica' -count=1
cd code/backend && go test ./internal/module/instance/infrastructure ./internal/config -run 'ProxyTicket|TestValidate(RejectsEnabledDefenseSSHWithoutHostKeyPath|AllowsEnabledDefenseSSHWithHostKeyPath)' -count=1
docker compose -f docker/docker-compose.dev.yml config
git diff --check
```

如后续继续修改 gateway shutdown、compose dev 启动链或 host-key contract，应至少重跑上面同一组验证。

## Residual Risk

- 当前 review 只覆盖 focused tests 与 compose 配置展开，没有执行真实的多 gateway + TCP LB 实机演练，因此文档里关于“drain before terminate”的操作仍主要依赖代码路径和单元测试。
- `ctf-awd-defense-ssh-host-key` 只解决了开发态预置 owner；生产或共享环境仍然依赖部署层显式分发同一份 host key 文件，这一点目前已经文档化，但还没有额外的机械检查脚本约束。

## Touched Known-debt Status

- 本轮没有触达已登记的 oversized / owner-mixed 结构债 surface。
- 上一轮独立 review 标出的两条 blocker 已在当前 touched surface 内收口，没有作为 residual risk 遗留。
