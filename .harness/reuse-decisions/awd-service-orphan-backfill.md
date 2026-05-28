# Reuse Decision

## Change type
data-fix

## Existing code searched

- `code/backend/cmd/seed-teaching-review-data/main.go`
- `code/backend/internal/module/contest/entity/contest_awd_service.go`
- `code/backend/internal/module/challenge/entity/awd_challenge.go`
- `code/backend/migrations/000010_add_low_risk_foreign_keys.up.sql`
- `code/backend/migrations/000011_add_additional_foreign_keys.up.sql`

## Similar implementations found

- `code/backend/cmd/seed-teaching-review-data/main.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_support.go`

## Decision
extend_existing

## Reason

这次不是新增新的数据模型或迁移模式，而是回填历史 seed 残留的 `awd_* .service_id` 孤儿引用。P1 已经修复了 seed 脚本，不再继续制造新孤儿；当前需要做的是参照现有 `contest_awd_services` 结构，为受影响赛事补父记录，再把子表外键定向回填到真实父记录。

## Files to modify

- `.harness/reuse-decisions/awd-service-orphan-backfill.md`
- `docs/plan/impl-plan/2026-05-28-awd-service-orphan-backfill-implementation-plan.md`
- `docs/reviews/backend/2026-05-28-awd-service-orphan-backfill-review.md`
- `docs/operations/2026-05-28-awd-service-orphan-backfill.md`

## After implementation

- `awd_* .service_id` 这批历史孤儿会被收口；`challenges.image_id = 0` 作为历史 no-image 哨兵值的问题继续保留，后续按独立 schema / contract 清理线处理。
