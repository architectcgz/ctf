# AWD Defense SSH Gateway Split Independent Review

## Review Target

- Repository: `ctf`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-07-awd-defense-ssh-gateway-split`
- Branch: `task/2026-06-07-awd-defense-ssh-gateway-split`
- Task slug: `2026-06-07-awd-defense-ssh-gateway-split`
- Plan: `docs/plan/archive/impl-plan/2026-06/2026-06-07-awd-defense-ssh-gateway-split-implementation-plan.md`
- Diff source: current uncommitted diff in this worktree
- Files reviewed:
  - `README.md`
  - `code/backend/Dockerfile`
  - `code/backend/cmd/awd-defense-ssh-gateway/main.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway_builder.go`
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router.go`
  - `code/backend/internal/app/composition/runtime_node_execution_router_test.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
  - `code/backend/internal/app/router_composition_structure_test.go`
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go`
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway_test.go`
  - `code/backend/scripts/docker-entrypoint.sh`
  - `docker/ctf/docker-compose.dev.yml`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/operations/runtime-agent-deployment.md`
  - `scripts/lib/check-consistency/architecture.sh`

## Classification Check

同意 `非琐碎任务` / 独立 gate review 判定。

原因：

- 这次改动同时改变了进程边界、bootstrap 生命周期、runtime node 路由、镜像入口和 compose 拓扑。
- `completion-full` 或实现上下文自检只能证明“当前作者看起来没问题”，不能替代独立 reviewer 对 owner 边界和停机语义的判断。

## Gate Verdict

`blocked`

## Findings

### 1. Blocking: gateway 进程启动时把 SSH session 绑定到 `context.Background()`，停机不会取消已建立的 interactive exec

- 位置：
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go:39-67`
  - `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go:116-142`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go:79-110`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go:287-376`
  - 对照：`code/backend/internal/app/http_server.go:55-58,136-147`
  - 对照：`code/backend/internal/app/composition/root.go:99-124`
- 行为：
  - `RunAWDDefenseSSHGateway()` 用 `rootCtx := context.Background()` 调 `process.Start(rootCtx)`。
  - `process.Start()` 原样把这个 ctx 传给 `gateway.Start(ctx)`。
  - `gateway.Start()` 再把这个 ctx 传给 `serve -> handleConn -> handleSessionChannel -> runContainerCommand`。
  - `Shutdown()` 虽然先调用了 `root.Cancel()`，但 session 并没有绑定 `root.Context()`，所以已有的 interactive exec 不会因为进程停机而收到取消信号。
- 触发条件：
  - 已有 SSH 登录并正在跑 shell / `exec` 命令时收到 `SIGINT` / `SIGTERM`，或 supervisor 触发 graceful shutdown。
- 影响：
  - `gateway.Stop()` 只会关 listener 并等待 accept loop 退出，不会等活跃 session 退出。
  - 随后的 `runtimeCloser.Close()` 和 `closeResources()` 会在 session 仍运行时关闭 node client / DB / Redis，造成 SSH 会话被中断、流式 exec 竞争关闭中的 runtime client，或者留下不可控的停机行为。
  - 这和 API 里 background job 通过 `root.Context()` 启动的模式不一致，属于真实的 lifecycle regression。
- 修正方向：
  - gateway 进程应把 session 根上下文绑定到 `BuildRoot()` 生成的 app context，而不是 `context.Background()`。
  - 如果要支持 graceful drain，还需要显式跟踪活跃 session / exec，确保 `Shutdown()` 在超时前等待或取消它们，而不是只等 listener 退出。

## Material Findings

- `code/backend/internal/bootstrap/awd_defense_ssh_gateway.go` / `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
  - 必须把 gateway 的运行上下文改成可取消的 app context，并补上停机时对活跃 SSH session / interactive exec 的取消或等待语义。
  - 修复后至少需要补一条自动化验证，证明“建立中的 SSH interactive exec 在进程 shutdown 时会被取消或有界退出”，而不是仅 listener 停止。

## Non-blocking Suggestions

- `code/backend/scripts/docker-entrypoint.sh:48-56`
  - 现在只有“零参数”时才默认补成 `/app/ctf-api`。这让旧的“只传 API 参数、不显式写二进制路径”调用方式从 `exec /app/ctf-api "$@"` 退化成 `exec "$@"`。
  - 如果仓库外已有 `docker run image --flag` 或 compose `command: ["--flag"]` 之类调用，会直接失败。没有仓库内证据证明当前已受影响，所以先记为非阻塞，但建议把 args-only API 启动兼容回来，或在文档里明确“覆盖 command 时必须显式写 `/app/ctf-api` / `/app/ctf-awd-defense-ssh-gateway`”。

- `README.md:55-67`
  - 全容器联调的风险说明仍主要强调 `ctf-api` 持有宿主 Docker 控制权。现在 `ctf-awd-defense-ssh-gateway` 在单机 compose 里同样挂了 `/var/run/docker.sock`，而且对外暴露了独立 `2222` 端口，建议把这层威胁模型一起写清楚。

## Missing Validation

- 没有自动化验证覆盖“gateway 收到 shutdown 时，已建立 SSH session / interactive exec 的取消或有界退出”。
- 没有 shell 级验证覆盖新的 `docker-entrypoint.sh` 双入口语义：
  - API 默认启动时仍会跑 migration
  - gateway 启动时不会误跑 migration
  - 只传 API 参数的旧调用方式是否仍兼容，或已被明确废弃

## Senior Implementation Assessment

整体拆分方向是对的：

- `InstanceModule` 已经不再注册 `awd_defense_ssh_gateway` background job，owner boundary 比以前清楚。
- `BuildAWDDefenseSSHGateway()` 复用了现有 proxy ticket / scope / runtime 组合，而不是再造一套 SSH 鉴权链路。
- `runtimeNodeExecutionRouter` 给 interactive exec 补上 `container_id -> node_id` 路由，和文件写入、cleanup 的 node authority 一致，属于最小正确扩展。

但当前实现把“独立进程”只拆到了监听和装配层，还没有把停机语义一起拆干净。对于一个持有长连接和 interactive stream 的 SSH gateway，这不是可接受的剩余风险，应该在本次 touched surface 内收口。

## Required Re-validation

修复 blocker 后，至少重新执行：

```bash
cd code/backend
go test ./internal/app/composition -run 'TestRuntimeNodeExecutionRouter|TestAWDDefenseSSHGateway|TestBuildAWDDefenseSSHGateway' -count=1
go test ./internal/bootstrap -run 'TestRunAWDDefenseSSHGateway|TestShutdownGracefully' -count=1
```

并新增一条针对 shutdown 的自动化验证，证明：

- gateway 启动后建立的 interactive exec 会在 shutdown 时收到取消或在超时内退出
- `runtimeCloser.Close()` 不会早于活跃 SSH session 的生命周期结束

本次 review 额外独立复跑结果：

```bash
cd code/backend
go test ./internal/app/composition -run 'TestRuntimeNodeExecutionRouter|TestAWDDefenseSSHGateway|TestBuildAWDDefenseSSHGateway' -count=1
go test ./internal/bootstrap -run 'TestRunAWDDefenseSSHGateway|TestShutdownGracefully' -count=1
```

结果：两条命令均 `PASS`。这能证明现有测试仍为绿色，但不能覆盖上面的 shutdown regression。

## Residual Risk

- 用户提供的 API health check 失败 `"no migration found for version 13"` 被当作既有问题处理；本 review 没把它升级成本 diff 的 blocker。
- 我没有看到真实 SSH 联调或多 node 远端 agent 下的 session teardown 证据，所以对“修复后是否真的按预期 drain/cancel”仍需要补 runtime 级验证。
- 对 `docker-entrypoint.sh` 的 args-only 兼容性，我只在本仓库内检查到 compose dev 当前用法没有直接踩中；仓库外部署是否依赖旧契约未知。

## Touched Known-debt Status

- 本次 touched surface 没有命中当前 fact source 里要求“只要触达就必须顺手关闭”的存量结构债。
- 但 `awd-defense-ssh-gateway` 作为新独立进程，shutdown / stream lifecycle 属于这次新引入的 owner surface；这里的未收口问题已经按 blocker 处理，没有降级成 residual risk。
