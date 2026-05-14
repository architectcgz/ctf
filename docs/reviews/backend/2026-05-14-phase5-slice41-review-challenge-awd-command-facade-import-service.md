# Phase5 Slice41 Review

## Scope

- 目标：收口 `challenge/application/commands/awd_challenge_command_facade.go` 对 `gorm.io/gorm` 的直接依赖
- 评审对象：
  - `challenge/application/commands/awd_challenge_command_facade.go`
  - `challenge/application/commands/awd_challenge_command_facade_test.go`
  - `challenge/runtime/module.go`
  - `architecture_allowlist_test.go`
  - 对应 reuse / implementation plan 文档

## Independent Review

- Reviewer：`Reviewer`
- Classification check：`agree`
- Gate verdict：`pass`
- 结论：`no findings`

## Notes

- 这刀只把 facade 的 `*gorm.DB` 装配职责收回到 runtime，没有声称已经完成 `AWDChallengeImportService` 自己的 `gorm` / tx 迁移。
- review 认可当前切法：先清 facade surface，再把 import service 内部事务面留给后续更大的 slice。
- `challenge/runtime/module.go` 仍是已知 oversized surface，但这次只是让 import service owner 更明确，没有新增职责扩散，因此不构成 blocker。

## Verification

```bash
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/commands -run 'TestNewAWDChallengeCommandFacade' -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/runtime -run '^$' -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module -run 'TestApplicationConcreteDependencyAllowlistIsCurrent' -count=1 -timeout 300s
```

## Review Result

- 当前 diff 已通过独立 correctness / regression review。
- slice41 可以提交，并作为后续清理 `AWDChallengeImportService` / `ChallengeService` 内部 gorm 依赖的前置收口。
