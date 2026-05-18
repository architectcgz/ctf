# practice contest registration scope contract 收口实现方案

## Objective

把 `practice` 模块对 `ContestRegistration` 的跨模块依赖，从全局持久化实体收成模块内只读参与信息视图，先去掉 `practice -> model.ContestRegistration` 这条直接耦合。

## Non-goals

- 不迁移 `ContestRegistration` 持久化实体 owner
- 不修改 contest 注册写路径和审批流程
- 不改变 contest 实例启动的行为和错误语义

## Inputs

- `internal/model/contest_registration.go`
- `internal/module/practice/ports/ports.go`
- `internal/module/practice/infrastructure/repository.go`
- `internal/module/practice/infrastructure/contest_scope_repository.go`
- `internal/module/practice/application/commands/contest_instance_scope.go`
- `.harness/reuse-decisions/practice-contest-registration-scope-contract-localization.md`

## Ownership Evaluation

- `ContestRegistration` 持久化实体 owner 不在 `practice`
- `practice` 当前只消费 `status` 和 `team_id`
- 最小收口点是 `practice` 自己的 scope view，而不是 contest 的 ORM 实体
- 这刀完成后，`practice` 的行为依赖会收敛到受控只读 contract

## Task slices

1. 在 `practice/ports` 定义参与信息只读视图
2. 更新 `practice` repository / scope repository 返回新视图
3. 更新 `contest_instance_scope` 和相关测试
4. 确认 `practice` ports 不再暴露 `model.ContestRegistration`

## Validation

- `go test ./internal/module/practice/... -count=1`
- `go test ./internal/app -run 'TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `practice` 是否已经不再直接依赖 `ContestRegistration` 实体
- 参与状态和队伍 ID 语义是否保持不变
- 这刀是否只做 contract 收口，没有偷渡 owner 迁移

## Rollback

本刀无 schema 变更，如有回归可把 `practice` ports 临时切回 `model.ContestRegistration`。
