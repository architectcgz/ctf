# Reuse Decision

## Change type

config / runtime / api / composition / docs

## Existing code searched

- `code/backend/internal/config/config.go`
- `code/backend/internal/model/topology.go`
- `code/backend/internal/module/practice/application/commands/`
- `code/backend/internal/module/runtime/api/http/handler.go`
- `code/backend/internal/module/instance/application/queries/instance_service.go`
- `docker/ctf/docker-compose.dev.yml`
- `.harness/reuse-decisions/registry-access-endpoint.md`

## Similar implementations found

- `container.registry.access_server` 已经把“canonical 对外地址”和“容器内直连地址”拆成两类配置。
- `model.ResolveRuntimeAliasAccessURL(...)` 已经承担运行时内部地址重写。
- `runtime/api/http` 与 `instance/application/queries` 已经是实例访问地址的集中出口。

## Decision

extend_existing

## Reason

当前问题不是新增一套运行时协议，而是同一个 `access_url` 同时被后端内部探测 / HTTP 代理和学生侧直连复用。最小正确方案不是改表结构，而是在现有 runtime URL 生成与展示链路上补一个 `container.access_host`，把“内部可达地址”和“对外展示地址”分开处理。

## Files to modify

- `code/backend/internal/config/config.go`
- `code/backend/internal/config/config_test.go`
- `code/backend/internal/model/topology.go`
- `code/backend/internal/model/topology_runtime_test.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- `code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- `code/backend/internal/module/practice/domain/mappers.go`
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/runtime/module.go`
- `code/backend/internal/module/runtime/runtime/adapters.go`
- `code/backend/internal/module/runtime/runtime/adapters_test.go`
- `code/backend/internal/module/runtime/api/http/handler.go`
- `code/backend/internal/module/instance/application/commands/instance_service.go`
- `code/backend/internal/module/instance/application/queries/instance_service.go`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
- `code/backend/internal/module/practice/application/commands/service_test.go`
- `docker/ctf/docker-compose.dev.yml`
- `docker/ctf/infra/README.md`
- `docs/architecture/backend/03-container-architecture.md`

## After implementation

- 本地 Docker 部署后端时，`public_host` 保持学生侧地址，`access_host` 专供后端容器内部链路使用。
- 后续如果迁到后端与宿主同网络视角的部署形态，可以清空 `access_host` 回退到单地址模式。
