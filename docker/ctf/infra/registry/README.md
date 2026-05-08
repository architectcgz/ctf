`ctf-registry` 的相关事实源统一放在这里。

- `ctf-platform-registry.env`: 平台后端与 `ctf-api` 使用的唯一 registry 配置文件。
- `ctf-platform-registry.env.example`: 示例模板。
- `runtime/auth/`: Basic Auth 文件与其他 registry 认证运行态文件。
- `runtime/data/`: registry 镜像数据目录。

默认通过 `scripts/registry/deploy-private-registry.sh` 维护，不建议手工改写生成文件。
