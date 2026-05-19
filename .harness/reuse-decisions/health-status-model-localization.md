# Reuse Decision

## Change type

model localization / service contract cleanup

## Existing code searched

- `code/backend/internal/model/health.go`
- `code/backend/internal/service/health/service.go`
- `code/backend/internal/module/**/contracts/*.go`

## Similar implementations found

- `code/backend/internal/module/*/contracts` 负责跨模块契约
- `code/backend/internal/service/health` 已是健康检查链路的唯一 owner

## Decision

refactor_existing

## Reason

`HealthStatus` 仅被 `internal/service/health/service.go` 使用，不属于跨模块共享领域模型。继续放在 `internal/model` 会制造无意义共享层依赖。最小改动是把结构体收口到 `service/health` 本地并删除全局模型文件。

## Files to modify

- `.harness/reuse-decisions/health-status-model-localization.md`
- `code/backend/internal/model/health.go`
- `code/backend/internal/service/health/service.go`

## After implementation

- `HealthStatus` 由 `internal/service/health` 独立维护
- `internal/model/health.go` 删除
