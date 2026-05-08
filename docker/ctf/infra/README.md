`docker/ctf/infra/` 是 CTF 项目自带 infra 的统一入口。

- `postgres/`: `ctf-postgres` 的目录占位与说明。当前持久化仍使用 Compose named volume `ctf-postgres-data`。
- `redis/`: `ctf-redis` 的目录占位与说明。当前持久化仍使用 Compose named volume `ctf-redis-data`。
- `registry/`: `ctf-registry` 的运行态目录与平台后端 registry 环境变量事实源。

当前默认约定：

- `ctf-api` 只从 `docker/ctf/infra/registry/ctf-platform-registry.env` 读取 registry 配置。
- `ctf-registry` 的默认认证与数据目录位于 `docker/ctf/infra/registry/runtime/`。
- 旧的 `$HOME/ctf-registry` 和 `docker/ctf/.env.registry` 只作为迁移来源，不再作为默认配置入口。
