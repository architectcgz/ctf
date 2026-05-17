# Reuse Decision

## Change type

config / infrastructure / runtime / docs

## Existing code searched

- `code/backend/internal/config/config.go`
- `code/backend/internal/config/config_test.go`
- `code/backend/internal/module/challenge/infrastructure/registry_client.go`
- `code/backend/internal/module/challenge/infrastructure/registry_client_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `scripts/registry/deploy-private-registry.sh`
- `docker/ctf/infra/registry/ctf-platform-registry.env.example`
- `docs/architecture/features/题包Registry交付架构.md`

## Similar implementations found

- `container.registry.server` 已经同时服务于 image ref 生成、Docker auth 和 verifier wiring
- `challenge/infrastructure/registry_client.go` 已经是单一 registry verifier adapter，适合扩一个可选访问地址
- `scripts/registry/deploy-private-registry.sh` 已经是本地 private registry env 的唯一生成入口

## Decision

extend_existing

## Reason

当前问题不是缺少新的 registry 能力，而是已有 `container.registry.server` 被三段不同执行面混用。最小正确方案不是新建第二套 registry service，也不是把 DB 里的 image ref 改成 compose service 名，而是在现有 `container.registry` 上补一个仅供 verifier 直连的可选访问地址，并让本地 deploy 脚本自动生成它。

## Files to modify

- `code/backend/internal/config/config.go`
- `code/backend/internal/config/config_test.go`
- `code/backend/internal/module/challenge/infrastructure/registry_client.go`
- `code/backend/internal/module/challenge/infrastructure/registry_client_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/configs/config.yaml`
- `docker/ctf/infra/registry/ctf-platform-registry.env.example`
- `docker/ctf/infra/registry/ctf-platform-registry.env`
- `scripts/registry/deploy-private-registry.sh`
- `scripts/registry/deploy-private-registry.conf.example`
- `docs/architecture/features/题包Registry交付架构.md`
- `docs/architecture/backend/03-container-architecture.md`
- `docker/ctf/infra/README.md`
- `docker/ctf/infra/registry/README.md`

## After implementation

- 本地 compose dev 若继续使用 `127.0.0.1:*` 作为 canonical registry server，应默认配套写出 `access_server`
- 后续如果迁到 TLS 或可被 daemon/container 共同访问的真实域名，可直接清空 `access_server`，回退到单地址模式
