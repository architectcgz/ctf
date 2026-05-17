`docker/ctf/infra/` 是 CTF 项目自带 infra 的统一入口。

- `postgres/`: `ctf-postgres` 的目录占位与说明。当前持久化仍使用 Compose named volume `ctf-postgres-data`。
- `redis/`: `ctf-redis` 的目录占位与说明。当前持久化仍使用 Compose named volume `ctf-redis-data`。
- `registry/`: `ctf-registry` 的运行态目录与平台后端 registry 环境变量事实源。

当前默认约定：

- `ctf-api` 只从 `docker/ctf/infra/registry/ctf-platform-registry.env` 读取 registry 配置。
- 该 env 文件至少包含 `CTF_CONTAINER_REGISTRY_SERVER`；本地 compose dev 如果 canonical server 仍是 `127.0.0.1:*`，还会额外包含 `CTF_CONTAINER_REGISTRY_ACCESS_SERVER=ctf-registry:5000`，仅给 `ctf-api` 进程直连 registry API 使用。
- 当 `ctf-api` 自己也运行在 Docker 容器里时，本地 compose dev 需要同时区分两类 host：
  - `CTF_CONTAINER_PUBLIC_HOST=127.0.0.1`：继续给学生 / 浏览器 / TCP 直连返回宿主机本地地址。
  - `CTF_CONTAINER_ACCESS_HOST=host-gateway.internal`：只给 `ctf-api` 容器内部的实例就绪探测和 HTTP 代理使用。
  - 对应地，compose 需要给 `ctf-api` 增加 `host-gateway.internal:host-gateway`，保证后端容器能回到宿主机发布端口。
- `ctf-registry` 的默认认证与数据目录位于 `docker/ctf/infra/registry/runtime/`。
- 旧的 `$HOME/ctf-registry` 和 `docker/ctf/.env.registry` 只作为迁移来源，不再作为默认配置入口。
