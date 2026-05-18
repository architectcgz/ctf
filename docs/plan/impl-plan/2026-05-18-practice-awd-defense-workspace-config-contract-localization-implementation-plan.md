# practice awd defense workspace config contract 收口实现方案

## Objective

把 `practice` 模块在 `awd_defense_workspace_support.go` 中对 `ContestAWDServiceSnapshot` 的直接解析，收进现有 `ContestAWDServiceRuntimeSubject` contract，消除工作区配置读取对 contest snapshot 原始存储形状的应用层依赖。

## Non-goals

- 不修改 defense workspace 编排行为
- 不迁移 `ContestAWDService` 持久化实体 owner
- 不处理 `PortAllocation` 读写和 runtime 事务 owner

## Inputs

- `internal/module/practice/application/commands/awd_defense_workspace_support.go`
- `internal/module/practice/infrastructure/contest_awd_runtime_subject_mapper.go`
- `internal/module/practice/ports/ports.go`
- `.harness/reuse-decisions/practice-awd-defense-workspace-config-contract-localization.md`

## Ownership Evaluation

- `defense_workspace` 仍属于 contest AWD service snapshot 的持久化内部形状
- `practice` 真正需要的是已解析好的 workspace config、checker token env 和 seed signature
- 最小收口点是扩展 `practice` 本地 runtime subject contract，而不是继续暴露 raw snapshot

## Task slices

1. 扩展 `ContestAWDServiceRuntimeSubject` 表达 workspace config / seed signature
2. 在 `practice/infrastructure` 完成 snapshot -> typed config 映射
3. 更新 `prepareAWDDefenseWorkspacePlan` 和相关测试
4. 确认 `practice` 应用层不再直接 decode contest AWD snapshot

## Validation

- `go generate ./internal/module/practice/...`
- `go test ./internal/module/practice/... -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `go test ./internal/app -run 'TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- defense workspace config 语义是否保持不变
- `practice` 应用层是否彻底退出 raw snapshot decode
- 这刀是否仍保持在 contract 收口范围，没有偷渡 owner / tx 重构

## Rollback

本刀无 schema 变更，如有回归可临时把 `prepareAWDDefenseWorkspacePlan` 切回原始 snapshot decode 路径。
