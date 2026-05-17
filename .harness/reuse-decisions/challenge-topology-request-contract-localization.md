# Reuse Decision

## Change type
contracts / mapper / domain service / tests

## Existing code searched
- code/backend/internal/module/challenge/contracts/topology.go
- code/backend/internal/module/challenge/application/commands/topology_command_input.go
- code/backend/internal/module/challenge/domain/topology_codec.go
- code/backend/internal/module/challenge/domain/package_topology_parser.go
- code/backend/internal/module/challenge/api/http/request_mapper_gen.go

## Similar implementations found
- code/backend/internal/module/challenge/contracts/challenge_core.go
- code/backend/internal/module/challenge/contracts/challenge_import.go
- code/backend/internal/module/challenge/application/commands/writeup_command_input.go

## Decision
extend_existing

## Reason
模块已采用 `challenge/contracts` 作为 handler 与业务边界，这一刀继续复用同一策略，把 topology 请求参数从 `internal/dto` 迁移到模块 contracts，避免 `topology` 链路继续依赖全局 dto，并保持 request mapper 生成式转换。

## Files to modify
- code/backend/internal/module/challenge/contracts/topology.go
- code/backend/internal/module/challenge/application/commands/topology_command_input.go
- code/backend/internal/module/challenge/domain/topology_codec.go
- code/backend/internal/module/challenge/domain/package_topology_parser.go
- code/backend/internal/module/challenge/api/http/request_mapper_gen.go
- code/backend/internal/module/challenge/application/commands/topology_service_context_test.go
- code/backend/internal/module/challenge/application/commands/writeup_topology_service_test.go
