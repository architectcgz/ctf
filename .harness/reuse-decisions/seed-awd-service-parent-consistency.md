# Reuse Decision

## Change type
job / service / data seed

## Existing code searched

- `code/backend/cmd/seed-teaching-review-data/main.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_support.go`
- `code/backend/internal/module/contest/entity/contest_awd_service.go`
- `code/backend/internal/module/contest/entity/contest_awd_service_snapshot.go`
- `code/backend/internal/module/contest/testsupport/fixtures.go`

## Similar implementations found

- `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- `code/backend/internal/module/contest/testsupport/fixtures.go`

## Decision
extend_existing

## Reason

当前 `seed-teaching-review-data` 的 AWD 样本已经具备 contest / round / team / attack / traffic 的完整种子结构，问题只在于把 `service_id` 当成逻辑常量写进了子表，却没有同步创建 `contest_awd_services` 父记录。

这次不新建独立 seed 子模块，也不跨层直接复用 command service。最小修复是：

- 扩展现有 `awdChallengeCatalog`，让 seed 端拿到构建父记录所需的 challenge 元数据。
- 在 seed 命令内部按现有 `contest_awd_services` 字段语义补一份最小 builder。
- 先创建真实 `contest_awd_services`，再把数据库生成的 `service_id` 传给 `AWDTeamService / AWDAttackLog / AWDTrafficEvent`。
- 顺手补强 `resetSeededAWDData`，显式清理 `contest_awd_services`，避免旧脏数据依赖 FK 级联才被回收。

## Files to modify

- `code/backend/cmd/seed-teaching-review-data/main.go`
- `docs/plan/impl-plan/2026-05-28-seed-awd-service-parent-consistency-implementation-plan.md`

## After implementation

- 若后续其他 seed 命令也要构造 AWD contest service，可再评估是否沉淀共享 helper；本次先不抽象。
