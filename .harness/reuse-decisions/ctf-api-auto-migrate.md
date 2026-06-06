# Reuse Decision

## Change type
schema / migration

## Existing code searched
- `code/backend/Dockerfile`
- `code/backend/scripts/dev-run.sh`
- `code/backend/migrations/000012_create_runtime_nodes.up.sql`
- `docker/ctf/docker-compose.dev.yml`
- `README.md`

## Similar implementations found
- `code/backend/scripts/dev-run.sh`
  - 已有 `golang-migrate` CLI 版本和 `run_migrations()` 入口，可直接复用为镜像内 migration owner。
- `code/backend/migrations/000012_create_runtime_nodes.up.sql`
  - 已有正式 SQL migration，应该继续作为 schema owner，而不是把 `instances.node_id` 改回应用启动时的 `AutoMigrate`。

## Decision
extend_existing

## Reason
当前问题不是缺少 migration 文件，而是全容器启动链路没有执行正式 migration，导致数据库停在 `schema_migrations=11`，同时又被运行时 `AutoMigrate` 留下了半迁移的 `runtime_nodes` 表。最小正确修复是复用现有 `golang-migrate` 能力，把它接到镜像启动入口，并把 `000012` 收口成可重入的修复型 migration，避免继续扩散应用内 schema owner。

## Files to modify
- `code/backend/Dockerfile`
- `code/backend/scripts/docker-entrypoint.sh`
- `code/backend/migrations/000012_create_runtime_nodes.up.sql`
- `docker/ctf/docker-compose.dev.yml`
- `README.md`
