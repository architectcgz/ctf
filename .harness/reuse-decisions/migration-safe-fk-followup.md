# Reuse Decision

## Change type
migration

## Existing code searched

- `code/backend/migrations/000001_init_schema.up.sql`
- `code/backend/migrations/000009_create_network_allocations.up.sql`
- `code/backend/migrations/000009_create_network_allocations.down.sql`
- `code/backend/migrations/000010_add_low_risk_foreign_keys.up.sql`
- `code/backend/migrations/000010_add_low_risk_foreign_keys.down.sql`
- `code/backend/migrations/000011_add_additional_foreign_keys.up.sql`
- `code/backend/migrations/000011_add_additional_foreign_keys.down.sql`
- `code/backend/internal/module/instance/entity/instance.go`
- `code/backend/internal/module/contest/entity/awd.go`
- `code/backend/internal/module/challenge/entity/awd_challenge.go`
- `code/backend/internal/module/challenge/entity/challenge.go`
- `code/backend/internal/module/ops/entity/audit_log.go`

## Similar implementations found

- `code/backend/migrations/000010_add_low_risk_foreign_keys.up.sql`
- `code/backend/migrations/000011_add_additional_foreign_keys.up.sql`

## Decision
extend_existing

## Reason

这次仍然是补历史遗漏的物理 FK，不需要新建新的 migration 模式。当前本地数据库 `schema_migrations` 版本还停在 `10`，说明 `000011` 尚未执行，因此直接扩展未落库的 `000011_add_additional_foreign_keys` 是最小改动。

已通过开发库 orphan 检查确认：`audit_logs.user_id`、`awd_*` 的 `created_by/last_verified_by/submitted_by_user_id`、`instances` 主引用列、`submissions` 主引用列当前都可以直接补约束；`awd_* .service_id` 和 `challenges.image_id` 仍有历史脏数据，暂不并入本次 migration。

full-chain smoke test 暴露出既有 `000009_create_network_allocations` 依赖连接默认 `search_path`，在 `000001` 已显式清空 `search_path` 的链路上会阻断 `000010/000011` 验证，因此把 `000009` 的 schema-qualified 修正纳入同一 migration 收口切片。

## Files to modify

- `code/backend/migrations/000009_create_network_allocations.up.sql`
- `code/backend/migrations/000009_create_network_allocations.down.sql`
- `code/backend/migrations/000010_add_low_risk_foreign_keys.up.sql`
- `code/backend/migrations/000010_add_low_risk_foreign_keys.down.sql`
- `code/backend/migrations/000011_add_additional_foreign_keys.up.sql`
- `code/backend/migrations/000011_add_additional_foreign_keys.down.sql`
- `docs/plan/impl-plan/2026-05-28-migration-safe-fk-followup-implementation-plan.md`

## After implementation

- 剩余脏数据列继续保留在迁移技术债范围内，等历史数据清洗完成后再单独补 FK。
