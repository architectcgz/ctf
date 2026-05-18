# contest core persistence entity owner localization 实施计划

## Objective

把 contest 主干持久化实体的真实定义收回 `internal/module/contest/entity`，并让 `contest` 模块内部改为直接依赖模块内实体。

## Non-goals

- 不处理 `AWDScopeControl`
- 不处理 `AWDServiceOperation`
- 不在本刀删除 `internal/model` 下对应兼容入口
- 不一次清理 `practice / runtime / assessment / app` 对这些兼容入口的外部引用

## Inputs

- `internal/model/contest.go`
- `internal/model/contest_registration.go`
- `internal/model/team.go`
- `internal/model/submission.go`
- `internal/model/contest_challenge.go`
- `internal/model/contest_awd_service.go`
- `internal/model/contest_awd_service_snapshot.go`
- `internal/model/awd.go`
- `internal/module/contest/...`
- `.harness/reuse-decisions/contest-core-persistence-entity-owner-localization.md`

## Ownership evaluation

- owner 明确：上述实体都由 `contest` 模块定义和维护业务语义
- landing zone 明确：`internal/module/contest/entity`
- 兼容策略明确：`internal/model` 只保留 alias 过渡，不再保留真实定义
- 结构收敛目标明确：先让 `contest` 模块内部完全切到新 owner，再继续清理外部模块残留依赖

## Task slices

1. 新增 `contest/entity` 下的实体定义文件
2. 把 `internal/model` 同名文件改成兼容 alias / 转发
3. 更新 `contest` 模块内部 ports / application / domain / infrastructure / testsupport / tests
4. 运行最小充分验证，确认 contest 内部 owner 已切换且外部兼容未破坏

## Validation

- `go test ./internal/module/contest/... -count=1`
- `go test ./internal/module/practice/... ./internal/module/runtime/... -count=1`
- `go test ./internal/app -run 'TestFullRouter_AuthorizedSmokeMatrix|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
- `go test ./internal/module -run TestModuleArchitectureBoundaries -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `contest` 模块内部是否还残留对 contest 自有 `internal/model` 实体的直接依赖
- `internal/model` 是否已经只剩兼容 alias / 转发语义
- 外部模块经由兼容入口是否仍可通过编译与测试

## Rollback

本刀无 schema 变更。如有回归，可恢复 `internal/model` 中的真实定义并撤回 `contest/entity` 引用切换。
