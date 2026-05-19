# teaching shared kernel 边界纯化实施计划

## Objective

把当前 `teaching` 共享教学内核里的三处边界污染先收成一个可审阅切片：

- 抽出共享 taxonomy，解除 `internal/teaching` 对 `challenge/contracts` 的反向依赖
- 删除 `challenge/contracts` 对共享 taxonomy 的兼容出口，只保留 challenge owner 自己的契约
- 将 `teaching_query/ports` 中的 GORM tag 下沉到 `infrastructure`
- 移除 `classwindow.Parse` 的隐式 `time.Now()` fallback，恢复纯规则输入

## Non-goals

- 不把 `internal/teaching` 升格成完整 `internal/module/teaching`
- 不改教师端 API shape、推荐排序语义或班级复盘输出结构
- 不处理 `assessment` 中是否继续下沉 `BuildRecommendationPlan` 的后续拆分

## Inputs

- `internal/teaching/advice/advice.go`
- `internal/teaching/classwindow/window.go`
- `internal/module/challenge/contracts/*.go`
- `internal/module/teaching_query/ports/query.go`
- `internal/module/teaching_query/infrastructure/repository.go`
- `internal/module/teaching_query/architecture_test.go`
- `.harness/reuse-decisions/teaching-shared-kernel-boundary.md`

## Ownership Evaluation

- owner 明确：教学事实评估规则仍由 `internal/teaching` 共享内核持有
- 共享语义落点明确：维度/难度枚举进入 `internal/shared/taxonomy`
- 兼容出口删除：共享 taxonomy 由调用方直接依赖，不再经 `challenge/contracts` 中转
- ORM 细节 owner 明确：GORM tag 只允许停留在 `teaching_query/infrastructure`

## Task slices

1. 新增 `internal/shared/taxonomy`，迁移维度与题目难度共享语义；`teaching/advice` 改依赖共享 taxonomy。
2. 将外围生产代码、seed 命令与测试里的共享维度/难度引用切到 `internal/shared/taxonomy`。
3. 删除 `challenge/contracts` 中的共享 taxonomy 兼容出口，并补禁止回流的架构守卫。
4. 清理 `teaching_query/ports/query.go` 中的 GORM tag，在 `infrastructure/repository.go` 增加 row struct 与转换。
5. 收紧 `classwindow.Parse` 输入约束，移除零值时间 fallback，并调整调用与测试。
6. 为新边界补守卫与回归测试。

## Validation

- `cd code/backend && go test ./internal/teaching/... -count=1`
- `cd code/backend && go test ./internal/module/challenge/... ./internal/module/teaching_query/... ./internal/module/assessment/... -count=1`
- `cd code/backend && go test ./internal/app/... ./internal/module/practice/... ./internal/module/runtime/... ./internal/module/contest/... ./cmd/seed-teaching-review-data/... -count=1`
- `cd code/backend && go test ./internal/module/teaching_query -run 'TestPortsDoNotDependOnDTOGinOrGORM|TestPortsDoNotDeclareGORMTags' -count=1`

## Review focus

- `internal/teaching` 是否已不再反向依赖 `module/*/contracts`
- `challenge/contracts` 是否已不再暴露共享维度/难度语义
- `ports` 是否只保留纯读模型与接口，不再携带持久化注解
- 时间窗规则是否完全由调用方显式传入 `now`

## Rollback

本刀无 schema 变更。若出现回归，可先回退共享 taxonomy 的接入点和 `classwindow.Parse` 收紧逻辑，不影响数据库结构。
