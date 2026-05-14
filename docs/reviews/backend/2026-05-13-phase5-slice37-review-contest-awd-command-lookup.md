# Phase5 Slice37 Review

## Scope

- 目标：收口 `contest` AWD command 这 7 个 application command 文件的 `gorm.ErrRecordNotFound` 依赖
- 评审对象：
  - `contest/ports/awd.go`
  - `contest/infrastructure/awd_command_repository.go`
  - `contest/infrastructure/awd_round_manager_adapter.go`
  - `contest/application/commands/awd_*`
  - `contest/runtime/module.go`
  - 对应 command / infrastructure tests

## Independent Review A

- Reviewer：`Reviewer the 7th`
- 结论：发现 1 条 scope blocker

### Finding 1

- 严重级别：Blocking
- 位置：
  - `code/backend/internal/module/architecture_allowlist_test.go`
  - `docs/design/backend-module-boundary-target.md`
- 问题：worker diff 把 allowlist 与长期设计文档一起带进了 slice37，实现面超出 plan 里声明的 command/ports/infrastructure/runtime/test 收口范围。
- 处理：
  - 已撤回 `docs/design/backend-module-boundary-target.md` 改动。
  - `architecture_allowlist_test.go` 保留为 leader-owned guardrail sync，用于让 `TestApplicationConcreteDependencyAllowlistIsCurrent` 与已落地代码保持一致；不再把它作为 worker 实现面的一部分。

## Independent Review B

- Reviewer：`Code Reviewer the 7th`
- 结论：`no findings`

## Residual Risk

- 目前没有独立的 runtime-level 测试直接锁定 `buildAWDHandler` 对 `AWDCommandRepository` / `AWDRoundManagerAdapter` 的注入。
- 本次通过 command / infrastructure / runtime compile-level 验证覆盖了主要回归面，但 wiring 断言仍主要依赖编译通过和行为测试间接覆盖。

## Verification

```bash
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/application/commands -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/infrastructure -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module -run 'TestApplicationConcreteDependencyAllowlistIsCurrent' -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh
```

## Review Result

- 当前 diff 已通过独立 correctness / regression review。
- spec review 提出的 scope blocker 已通过移除长期设计文档改动解决。
