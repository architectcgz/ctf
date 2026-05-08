`ctf-postgres` 的 Compose service 定义在 [docker-compose.dev.yml](../../docker-compose.dev.yml)。

这一层保留为 PostgreSQL 相关 infra 的统一落点。当前本地开发仍使用 named volume `ctf-postgres-data` 持久化数据，避免这次目录收口顺带触发数据库卷迁移。
