# Challenge AWD Command Facade Import Service Phase5 Slice41 Implementation Plan

**Goal:** 移除 `challenge/application/commands/awd_challenge_command_facade.go` 对 `gorm.io/gorm` 的直接依赖，把 import service 的构造 owner 收回到 runtime wiring。

## Objective

- 删除 `challenge/application/commands/awd_challenge_command_facade.go -> gorm.io/gorm`

## Non-goals

- 不处理 `awd_challenge_import_service.go` 仍直接依赖 `*gorm.DB`
- 不改 AWD challenge import 的事务实现或 image build 流程
- 不顺手改 challenge 其他 `gorm` allowlist

## Inputs

- `.harness/reuse-decisions/challenge-awd-command-facade-import-service-phase5-slice41.md`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_command_facade.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- `code/backend/internal/module/challenge/runtime/module.go`

## Ownership Boundary

- `challenge/application/commands/awd_challenge_command_facade.go`
  - 负责：组合 core service 与 import service，提供统一 command facade
  - 不负责：自行装配 `*gorm.DB`
- `challenge/runtime/module.go`
  - 负责：构造 `AWDChallengeImportService` 并注入 facade

## Change Surface

- Add: `.harness/reuse-decisions/challenge-awd-command-facade-import-service-phase5-slice41.md`
- Add: `docs/plan/impl-plan/2026-05-14-challenge-awd-command-facade-import-service-phase5-slice41-implementation-plan.md`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_command_facade.go`
- Add: `code/backend/internal/module/challenge/application/commands/awd_challenge_command_facade_test.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`

## Task

- `AWDChallengeCommandFacade` 构造器改成消费 repo + 预构建 import service
- runtime 负责先构造 `AWDChallengeImportService`，再注入 facade
- 补一个 facade 构造单测，保护“传入 import service 被原样复用”的装配契约
- 删掉这 1 条 allowlist

## Validation

```bash
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/commands -run 'TestNewAWDChallengeCommandFacade' -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/runtime -run '^$' -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module -run 'TestApplicationConcreteDependencyAllowlistIsCurrent' -count=1 -timeout 300s
```

## Review focus

- facade 是否只去掉 concrete 装配，没有改变 command 行为
- runtime 是否成为唯一 import service 装配 owner
- 这刀是否只清理 facade surface，没有假装已经完成 `AWDChallengeImportService` 的 `gorm` 迁移
