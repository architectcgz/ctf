# Phase5 Remaining Branch And Worktree Distillation Review

## Review Target

- 仓库：`/home/azhi/workspace/projects/ctf`
- 主分支：`main`
- 清理对象：
  - `challenge-image-command-not-found-contract-phase5-slice33`
  - `challenge-writeup-not-found-contract-phase5-slice28`
  - `contest-challenge-service-lookup-contract-phase5-slice38`
  - `challenge-transaction-heavy-concrete-surface-phase5-slice36`
  - `codex/phase5-slice39`
  - `codex/contest-awd-http-runtime-slice40`
- 评审目的：判断剩余 `phase5` 分支与 worktree 是否还有未回到 `main` 的有效代码，还是已经被主干吸收或被主干更新版覆盖。

## Classification Check

- 任务类型：cleanup / distillation / merge history recovery
- 复杂度：non-trivial
- 结论：`agree`

## Gate Verdict

- `pass`
- 结论：剩余 `phase5` worktree 没有需要继续人工抄回 `main` 的代码；`slice40` 分支也不应再做代码层 cherry-pick，而应保留当前主干实现并只回收分支历史。

## Findings

### 1. `slice33` 的脏改动已经被主干吸收，剩余差异是旧测试与旧计划文案

- 核对结果：
  - `challenge/application/commands/image_service.go` 与 worktree 完全一致。
  - `challenge/infrastructure/image_command_repository.go` 与 worktree 完全一致。
  - `challenge/runtime/module.go`、`image_service_test.go`、`image_command_repository_test.go` 的差异来自主干后续继续收口 image build / challenge core wiring，和更完整的测试辅助写法，不是功能缺失。
  - reuse decision 与 impl plan 在 `main` 上都存在，且主干版本比 worktree 版本更新。
- 影响判断：
  - 继续从该 worktree 抄代码会把 `main` 上已经更进一步的 wiring 和测试写法回退成旧版本。

### 2. `slice28` 的 writeup not-found 收口已经在主干落地，worktree 剩余差异同样是旧 wiring 与未回填计划状态

- 核对结果：
  - `challenge/application/commands/writeup_service.go`、`queries/writeup_service.go`、对应 sentinel 测试、`writeup_service_repository.go` 都与主干一致。
  - 差异集中在 `ports/ports.go`、`runtime/module.go`、`writeup_topology_service_test.go` 和 impl plan：
    - 主干已经继续推进 challenge module 的 adapter/wiring 收口。
    - worktree 里的 impl plan 仍保留未完成 checkbox，明显落后于主干事实。
- 影响判断：
  - 该 worktree 不再包含任何应补回 `main` 的 writeup 行为代码。

### 3. `slice38` 的 contest lookup 合约收口也已经在主干，worktree 只剩旧测试与旧 runtime 组装方式

- 核对结果：
  - `challenge_add_commands.go`、`contest_awd_service_service.go`、`ports/challenge.go`、两个 lookup adapter 文件都与主干一致。
  - 差异文件是测试初始化方式和旧版 runtime 组装；主干已经继续推进 `contest` runtime 依赖注入的后续收口。
  - 该 slice 的 reuse decision、impl plan、independent review 已经在主干。
- 影响判断：
  - 不存在额外需要从 worktree 补回的 command 合约逻辑。

### 4. `slice40` 分支的两条唯一提交已被主干后续实现覆盖，不适合直接 cherry-pick

- 分支唯一提交：
  - `516fff64` `重构(contest): 提取 awd http runtime adapter`
  - `404d0740` `重构(contest): 收口 awd jobs http runtime 依赖`
- 核对结果：
  - `contest/infrastructure/awd_http_runtime_adapter.go`
  - `contest/infrastructure/awd_http_runtime_adapter_test.go`
  - `contest/ports/http_runtime.go`
  - `contest/application/jobs/awd_http_checker_request.go`
  - `contest/application/jobs/awd_http_runtime_contract_test.go`
  - `contest/application/jobs/awd_http_target_client.go`
  - `contest/application/jobs/awd_probe_runtime.go`
  - `contest/application/jobs/awd_testsupport_test.go`
  这些文件在当前 `main` 上都不是“缺少 slice40 代码”，而是已经有一版更新的实现：
  - 主干保留了更细的 `executeAWDHTTPRequest` / `normalizeAWDHTTPRuntimeError` 包装，避免 jobs 侧把 runtime 错误映射散落回各调用点。
  - 主干的 `AWDHTTPRuntimeError.Error()`、timeout 归一化和 adapter 测试覆盖都比 `slice40` 分支版本更完整。
  - 对该分支执行 cherry-pick 会产生大面积冲突，而且冲突内容主要是在主干与旧版本之间二选一；保留主干版本是正确选择。
- 影响判断：
  - `slice40` 需要回收的是 branch history，不是 branch tree。

### 5. `slice36` 与 `slice39` 的提交历史已经是 `main` 祖先，可以直接清理

- `challenge-transaction-heavy-concrete-surface-phase5-slice36`
- `codex/phase5-slice39`

这两条分支没有额外脏改动，且已被 `main` 包含，不需要额外提炼。

## Material Findings

- 无需要继续修改 `main` 代码的 blocker。
- 需要执行的清理动作：
  - 为 `codex/contest-awd-http-runtime-slice40` 回收历史而不是强行 cherry-pick tree。
  - 删除所有剩余 `phase5` worktree 和对应本地分支，避免旧分支继续制造“看起来未合并”的噪音。

## Required Re-validation

```bash
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/... -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/... ./internal/module -run 'TestApplicationConcreteDependencyAllowlistIsCurrent' -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf && python3 scripts/check-docs-consistency.py
cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh
```

## Residual Risk

- `slice40` review 里记录过一个非阻塞观察：`resolveAWDHTTPDialOverride` 在 `application/jobs` 与 `infrastructure` 各保留一份实现。当前清理不改变这点，它属于后续继续收口时再处理的独立技术债，不影响这次 branch/worktree 回收结论。

## Touched Known-Debt Status

- 本次只清理历史分支与 worktree，不新增 touched surface。
- `slice33` / `slice28` / `slice38` / `slice40` 对应的 phase5 debt 已经由当前 `main` 的实现继续收口；剩余差异不是“主干漏掉了代码”，而是“旧 worktree 落后于主干”。
