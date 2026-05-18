# contest status transition model 模块内化实现方案

## Objective

把 `internal/model/contest_status_transition.go` 中的 `ContestStatusTransition` 持久化实体收回 `internal/module/contest/entity`，让 contest 状态机的仓储、重放链路和测试 schema 直接依赖模块内实体。

## Non-goals

- 不处理 `Contest`
- 不处理 `ContestRegistration`
- 不改变状态推进语义、副作用重放顺序或表结构

## Inputs

- `internal/model/contest_status_transition.go`
- `internal/module/contest/...`
- `internal/app/full_router_integration_test.go`
- `.harness/reuse-decisions/contest-status-transition-model-localization.md`

## Ownership Evaluation

- owner 明确：`ContestStatusTransition` 只属于 `contest` 状态机
- landing zone 明确：`internal/module/contest/entity/status_transition.go`
- 触达边界可控：外围只涉及 contest testsupport 和 app integration schema
- 结构收敛目标明确：删除旧全局模型文件

## Task slices

1. 新增 `contest/entity/status_transition.go`
2. 更新 contest repository / tests / testsupport
3. 更新 app integration schema
4. 删除 `internal/model/contest_status_transition.go`

## Validation

- `go test ./internal/module/contest/... -count=1`
- `go test ./internal/app -run 'TestFullRouter_AuthorizedSmokeMatrix|TestFullRouter_AccessControlMatrix' -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `ContestStatusTransition` 是否完全收敛到 `contest/entity`
- contest repository / testsupport 是否没有残留全局依赖
- app schema 是否同步切换

## Rollback

本刀无 schema 变更，如有回归可直接恢复到 `internal/model/contest_status_transition.go`。
