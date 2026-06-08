`ctf-redis` 的 Compose service 定义在 [docker-compose.dev.yml](../../docker-compose.dev.yml)。

这一层保留为 Redis 相关 infra 的统一落点。当前本地开发仍使用 named volume `ctf-redis-data` 持久化数据，避免这次目录收口顺带触发 Redis 数据卷迁移。
