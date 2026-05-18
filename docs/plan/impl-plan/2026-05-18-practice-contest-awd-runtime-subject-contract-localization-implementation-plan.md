# practice contest awd runtime subject contract 收口实现方案

## Objective

把 `practice` 模块对 `ContestAWDService.ServiceSnapshot` 的直接应用层解析，收成模块内只读 runtime subject contract，先去掉 `practice` 应用层对 contest snapshot 存储形状的直接耦合。

## Non-goals

- 不迁移 `ContestAWDService` 持久化实体 owner
- 不修改 admin AWD 服务列表 / 预热逻辑仍在使用的原始服务实体读取
- 不处理 defense workspace 配置解析，这部分留给下一刀

## Inputs

- `internal/model/contest_awd_service.go`
- `internal/module/practice/ports/ports.go`
- `internal/module/practice/infrastructure/repository.go`
- `internal/module/practice/infrastructure/contest_scope_repository.go`
- `internal/module/practice/application/commands/contest_instance_scope.go`
- `internal/module/practice/application/commands/contest_awd_runtime_subject.go`
- `.harness/reuse-decisions/practice-contest-awd-runtime-subject-contract-localization.md`

## Ownership Evaluation

- `ContestAWDService` 持久化实体 owner 不在 `practice`
- `practice` 启动实例只需要 service 可见性、challenge id、运行 challenge 视图和 topology 视图
- 最小收口点是 `practice` 自己的 runtime subject read contract，而不是 contest 的 ORM 实体或 snapshot DTO

## Task slices

1. 在 `practice/ports` 定义 AWD runtime subject 只读视图
2. 在 `practice/infrastructure` 把 `ContestAWDService + ServiceSnapshot` 映射成该视图
3. 更新 `contest_scope_repository`、`contest_instance_scope` 和相关测试
4. 确认 `practice` 应用层不再直接解 `ContestAWDServiceSnapshot`

## Validation

- `go generate ./internal/module/practice/...`
- `go test ./internal/module/practice/... -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `go test ./internal/app -run 'TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `practice` 应用层是否已经不再直接依赖 contest snapshot decode
- 可见性、challenge id、虚拟 challenge / topology 语义是否保持不变
- 这刀是否只做 runtime subject contract 收口，没有偷渡 owner 迁移

## Rollback

本刀无 schema 变更，如有回归可把 `contest_instance_scope` 临时切回原始 `FindContestAWDService + DecodeContestAWDServiceSnapshot` 路径。
