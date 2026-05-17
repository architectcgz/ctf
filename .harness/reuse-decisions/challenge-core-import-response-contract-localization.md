# Reuse Decision

## Change type
handler / service / mapper / contracts / tests

## Existing code searched
- code/backend/internal/module/challenge/contracts/
- code/backend/internal/module/challenge/api/http/response_mapper.go
- code/backend/internal/module/challenge/application/commands/response_mapper_goverter.go
- code/backend/internal/module/challenge/application/queries/response_mapper_goverter.go
- code/backend/internal/module/challenge/domain/topology_codec.go

## Similar implementations found
- code/backend/internal/module/challenge/contracts/tag_flag.go
- code/backend/internal/module/challenge/contracts/topology.go
- code/backend/internal/module/challenge/api/http/response_mapper.go
- code/backend/internal/module/challenge/application/commands/awd_challenge_command_facade.go

## Decision
extend_existing

## Reason
challenge 模块已经建立 contracts 作为 handler 与业务层之间的响应边界，这一刀继续复用并扩展同一模式，把 core 与 import 主链路的返回类型从 dto 收口到 module contracts；查询入参保持 repository 层 dto 契约不动，通过 application 层单点转换适配，避免跨层重复改动。

## Files to modify
- code/backend/internal/module/challenge/contracts/challenge_core.go
- code/backend/internal/module/challenge/contracts/challenge_import.go
- code/backend/internal/module/challenge/api/http/handler.go
- code/backend/internal/module/challenge/api/http/request_mapper.go
- code/backend/internal/module/challenge/api/http/request_mapper_gen.go
- code/backend/internal/module/challenge/api/http/response_mapper.go
- code/backend/internal/module/challenge/api/http/response_mapper_gen.go
- code/backend/internal/module/challenge/api/http/challenge_import_handler_test.go
- code/backend/internal/module/challenge/application/commands/challenge_service.go
- code/backend/internal/module/challenge/application/commands/challenge_import_service.go
- code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go
- code/backend/internal/module/challenge/application/commands/challenge_package_revision_service.go
- code/backend/internal/module/challenge/application/commands/response_mapper_goverter.go
- code/backend/internal/module/challenge/application/commands/response_mapper_goverter_gen.go
- code/backend/internal/module/challenge/application/queries/challenge_service.go
- code/backend/internal/module/challenge/application/queries/challenge_service_test.go
- code/backend/internal/module/challenge/application/queries/response_mapper_goverter.go
- code/backend/internal/module/challenge/application/queries/response_mapper_goverter_gen.go
- code/backend/internal/module/challenge/domain/mappers.go
- code/backend/internal/module/challenge/domain/response_mapper_goverter.go
- code/backend/internal/module/challenge/domain/response_mapper_goverter_gen.go
- code/backend/internal/module/challenge/domain/topology_codec.go
