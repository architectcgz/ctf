# Reuse Decision

## Change type

contract / command / query / handler / mapper dto localization

## Existing code searched

- `code/backend/internal/module/challenge/api/http/topology_handler.go`
- `code/backend/internal/module/challenge/application/{commands,queries}/topology_service.go`
- `code/backend/internal/module/challenge/domain/{topology_codec.go,response_mapper_goverter*.go}`
- `code/backend/internal/module/challenge/api/http/response_mapper*.go`

## Similar implementations found

- `image`、`tag/flag`、`writeup` 链路已收口到 `challenge/contracts`
- `api/http` 输出沿用统一 `to...` helper 映射模式

## Decision

refactor_existing

## Reason

`topology` handler 与 service 边界仍暴露全局 `dto.ChallengeTopologyResp` / `dto.EnvironmentTemplateResp`。本次把 topology 响应类型迁移到 `challenge/contracts`，并保持 challenge import 的旧 dto 输出链路不变，避免扩散到非本刀范围。

## Files to modify

- `.harness/reuse-decisions/challenge-topology-response-contract-localization.md`
- `code/backend/internal/module/challenge/contracts/topology.go`
- `code/backend/internal/module/challenge/domain/topology_codec.go`
- `code/backend/internal/module/challenge/domain/response_mapper_goverter.go`
- `code/backend/internal/module/challenge/application/commands/topology_service.go`
- `code/backend/internal/module/challenge/application/queries/topology_service.go`
- `code/backend/internal/module/challenge/api/http/topology_handler.go`
- `code/backend/internal/module/challenge/api/http/response_mapper.go`
