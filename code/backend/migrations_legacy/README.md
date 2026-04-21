# Legacy Migrations Archive

2026-04-21 将历史 `000001` 到 `000064` 迁移链归档到这里，并以最新 schema 重建为新的 `backend/migrations/000001_init_schema.*` baseline。

约定如下：

- `backend/migrations/` 只保留当前生效的 baseline 与后续新增迁移。
- `backend/migrations_legacy/` 仅用于追溯历史演进，不再参与 `golang-migrate` 执行。
- 如需验证旧链，可在独立临时目录中手动引用本归档，不要再将这些文件移回 `backend/migrations/`。
