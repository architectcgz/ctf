# 2026-05-28 migration safe fk followup implementation plan

> 状态：Current
> 类型：实施过程

## 目标与非目标

- 目标：在现有未执行的 `000011_add_additional_foreign_keys` 中补上当前开发库已确认无 orphan 的 FK。
- 非目标：不处理仍有历史脏数据的 `awd_* .service_id` 与 `challenges.image_id`，不重排 migration 编号。

## 输入事实

- `code/backend/migrations/000001_init_schema.up.sql`
- `code/backend/migrations/000009_create_network_allocations.up.sql`
- `code/backend/migrations/000009_create_network_allocations.down.sql`
- `code/backend/migrations/000010_add_low_risk_foreign_keys.up.sql`
- `code/backend/migrations/000011_add_additional_foreign_keys.up.sql`
- 开发库 `ctf` 的 orphan 检查结果

## 任务切片

1. 扩展 `000011` 的 up/down SQL。
2. 修正 `000009` 对默认 `search_path` 的隐式依赖，保证 full-chain migration smoke test 可执行。
3. 在开发库临时数据库顺序执行 `000001` 到 `000011`，确认 migration 可落库。
4. 复查剩余未补 FK，明确本次保留项。

## 预期改动面

- `code/backend/migrations/000009_create_network_allocations.up.sql`
- `code/backend/migrations/000009_create_network_allocations.down.sql`
- `code/backend/migrations/000011_add_additional_foreign_keys.up.sql`
- `code/backend/migrations/000011_add_additional_foreign_keys.down.sql`

## 兼容性与风险

- `users`、`challenges`、`contests`、`teams`、`contest_awd_services` 的软删除不触发物理 FK 删除动作，因此本次主要约束历史数据完整性，而不是改变日常删除路径。
- `instances` 属于运行态表，关联列删除策略优先保持“owner 被物理删除时实例记录一并收口”。
- `000001_init_schema.up.sql` 会把连接级 `search_path` 置空，因此后续 migration 不能再依赖未限定 schema 的 DDL；否则 clean-db 验证会在旧 migration 阶段提前失败。

## 验证

- `bash scripts/check-task-intake.sh --reuse-decision migration-safe-fk-followup`
- 在 `ctf-postgres` 临时数据库顺序执行 `000001` 到 `000011`
- 必要时对开发库再次执行 orphan 检查，确认保留项没有误并入
