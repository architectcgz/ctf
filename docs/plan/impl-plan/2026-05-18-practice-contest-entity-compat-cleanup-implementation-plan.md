# practice contest entity 兼容层清理实现方案

## Objective

把 `practice` 模块里仍通过 `internal/model` 访问的 contest-owned 持久化实体，改为直接依赖
`internal/module/contest/entity`，去掉 `practice -> internal/model contest alias` 这层过渡依赖。

## Non-goals

- 不迁移共享 owner 类型
- 不删除 `internal/model` 中的 contest 兼容 alias
- 不处理 `runtime`、`assessment`、`app` 等其他模块残留引用
- 不重新定义 `AWDScopeControl`、`AWDServiceOperation` 的 owner

## Inputs

- `.harness/reuse-decisions/practice-contest-entity-compat-cleanup.md`
- `code/backend/internal/module/contest/entity/*`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/application/commands/*`
- `code/backend/internal/module/practice/infrastructure/*`

## Ownership Evaluation

- `Contest / ContestRegistration / ContestChallenge / ContestAWDService / ContestAWDServiceSnapshot / Team / TeamMember / Submission`
  的 owner 已经明确在 `contest/entity`
- `practice` 只是消费这些实体，不应继续通过共享兼容层读取
- `Challenge / ChallengeTopology / Instance / AWDScopeControl / AWDServiceOperation / AWDDefenseWorkspace / User`
  这批类型本刀仍保留现状，不混入 owner 迁移

## Task slices

1. 更新 `practice/ports`，把 contest-owned repository contract 类型改成 `contestentity`
2. 更新 `practice/application/commands` 与 goverter 源文件，替换对应实体类型和常量引用
3. 更新 `practice/infrastructure` 仓储与 mapper，使 ORM 查询直接落到 `contestentity`
4. 更新 `practice` 测试、stub、testsupport，保证测试夹具也直接使用真实 owner
5. 运行 `go generate` 和受影响模块验证，确认没有遗留 `practice -> internal/model` contest alias

## Validation

- `go generate ./internal/module/practice/...`
- `go test ./internal/module/practice/... -count=1`
- `go test ./internal/module/runtime/... -count=1`
- `go test ./internal/app -run 'TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
- `go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestMapperWrappersFollowGlobalDelegationPolicy' -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `practice` 是否只替换了 contest-owned 类型，没有误伤共享 owner
- repository / service / test 中的常量和 snapshot encode/decode 是否都切到 `contestentity`
- goverter 生成代码是否随源接口同步更新

## Rollback

本刀无 schema 变更。如有回归，可以把 `practice` 里的 contest-owned 类型引用临时切回 `internal/model`
兼容 alias，再按文件逐步复核。
