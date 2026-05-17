# Reuse Decision

## Change type
mapper / domain service / contracts

## Existing code searched
- code/backend/internal/module/challenge/domain/response_mapper_goverter.go
- code/backend/internal/module/challenge/domain/topology_codec.go
- code/backend/internal/module/challenge/contracts/topology.go
- code/backend/internal/module/challenge/contracts/challenge_import.go

## Similar implementations found
- code/backend/internal/module/challenge/application/queries/response_mapper_goverter.go
- code/backend/internal/module/challenge/application/commands/response_mapper_goverter.go

## Decision
refactor_existing

## Reason
domain 层拓扑与题包响应此前仍经过 `dto -> contracts` 二次转换，属于迁移过渡代码。当前 contracts 结构已稳定，直接让 domain mapper 产出 contracts 可以减少重复映射和类型漂移风险。

## Files to modify
- code/backend/internal/module/challenge/domain/response_mapper_goverter.go
- code/backend/internal/module/challenge/domain/response_mapper_goverter_gen.go
- code/backend/internal/module/challenge/domain/topology_codec.go
