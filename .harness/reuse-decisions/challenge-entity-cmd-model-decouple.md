# Reuse Decision

## Change type

entity extraction / command import refactor / constants dependency decouple

## Existing code searched

- `code/backend/internal/model/challenge.go`
- `code/backend/internal/module/challenge/entity/image.go`
- `code/backend/internal/module/challenge/domain/package_parser.go`
- `code/backend/internal/teaching/advice/advice.go`
- `code/backend/cmd/import-challenge-packs/main.go`
- `code/backend/cmd/seed-demo-challenges/main.go`
- `code/backend/internal/module/challenge`

## Similar implementations found

- `challenge/entity/image.go` 已承载持久化实体与状态常量，适合作为 `Challenge` 实体归属位置。
- `challenge/contracts/persistence.go` 已暴露 challenge 领域常量，不需要在 `teaching` 侧继续直连 `internal/model`。
- `cmd/import-challenge-packs` 与 `cmd/seed-demo-challenges` 已按 challenge 领域概念组织，迁移到模块内 entity 不改变行为路径。

## Decision

refactor_existing

## Reason

目标是继续清理 `internal/model` 的残留依赖，优先处理可独立收口且风险较低的一组：challenge 题目实体定义归位到模块 `entity`，并让命令工具与 teaching 建议逻辑改用模块内 `entity/contracts`。这样不引入新抽象层，也不改变外部行为，只收口依赖归属。

## Files to modify

- `.harness/reuse-decisions/challenge-entity-cmd-model-decouple.md`
- `code/backend/cmd/import-challenge-packs/main.go`
- `code/backend/cmd/import-challenge-packs/main_test.go`
- `code/backend/cmd/seed-demo-challenges/main.go`
- `code/backend/internal/module/challenge/domain/package_parser.go`
- `code/backend/internal/module/challenge/entity/challenge.go`
- `code/backend/internal/teaching/advice/advice.go`
