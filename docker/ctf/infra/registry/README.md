`ctf-registry` 的相关事实源统一放在这里。

- `ctf-platform-registry.env`: 平台后端与 `ctf-api` 使用的唯一 registry 配置文件。
  - `CTF_CONTAINER_REGISTRY_SERVER`: 镜像 ref 与 Docker daemon 使用的 canonical registry server。
  - `CTF_CONTAINER_REGISTRY_ACCESS_SERVER`: 可选；仅给 `ctf-api` 进程直连 registry API 使用，本地 compose dev 默认是 `ctf-registry:5000`。
  - 当 `ctf-api` 本身跑在容器里，且 `CTF_CONTAINER_REGISTRY_SERVER` 仍是 `127.0.0.1` / `localhost` 时，必须同时提供 `CTF_CONTAINER_REGISTRY_ACCESS_SERVER`，否则后端启动会直接报配置错误。
- `ctf-platform-registry.env.example`: 示例模板。
- `runtime/auth/`: Basic Auth 文件与其他 registry 认证运行态文件。
- `runtime/data/`: registry 镜像数据目录。

默认通过 `scripts/registry/deploy-private-registry.sh` 维护，不建议手工改写生成文件。
