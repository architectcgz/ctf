# Reuse Decision

## Change type

contract / query / cache / practice / score

## Existing code searched

- `code/backend/internal/dto/score.go`
- `code/backend/internal/module/practice/**/*score*.go`
- `code/backend/internal/module/practice/ports/*.go`
- `code/backend/internal/module/practice/contracts/*.go`

## Similar implementations found

- `practice/contracts/manual_review.go` 已承接 practice 模块跨 handler / service 共享契约
- `practice/application/commands/submission_output.go` 承接 command 专属输出
- `practice/api/http/progress_dto.go` 承接 HTTP 专属响应

## Decision

refactor_existing

## Reason

`UserScoreInfo` 和 `RankingItem` 虽然来自 practice 排行榜 / 计分，但消费面横跨 query service、command service、ports、redis state store 和 handler interface。它们不再适合留在全局 `dto`，也不适合只挂在单一 application 层。最小正确方案是把这两个类型收回 `practice/contracts`，让 practice 自己成为唯一 owner。

## Files to modify

- `.harness/reuse-decisions/practice-score-dto-localization.md`
- `docs/plan/archive/impl-plan/2026-05-18-practice-score-dto-localization-implementation-plan.md`
- `code/backend/internal/dto/score.go`
- `code/backend/internal/module/practice/contracts/score.go`
- `code/backend/internal/module/practice/api/http/handler.go`
- `code/backend/internal/module/practice/application/commands/score_service.go`
- `code/backend/internal/module/practice/application/queries/score_service.go`
- `code/backend/internal/module/practice/application/queries/response_mapper_goverter.go`
- `code/backend/internal/module/practice/application/queries/response_mapper_goverter_gen.go`
- `code/backend/internal/module/practice/infrastructure/score_state_store.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/ports/score_state_context_contract_test.go`

## After implementation

- `practice` 计分相关代码不再引用 `dto.UserScoreInfo` / `dto.RankingItem`
- `internal/dto/score.go` 删除
