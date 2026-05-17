# Reuse Decision

## Change type
contracts / service / domain

## Existing code searched
- code/backend/internal/module/challenge/contracts/challenge_core.go
- code/backend/internal/module/challenge/application/commands/challenge_service.go
- code/backend/internal/module/challenge/domain/mappers.go

## Similar implementations found
- code/backend/internal/module/challenge/contracts/challenge_import.go
- code/backend/internal/module/challenge/contracts/writeup.go

## Decision
extend_existing

## Reason
核心链路已将 handler 与业务边界收口到 `challenge/contracts`，`ChallengeHintReq` 仍停留在全局 `dto`。本次直接扩展模块 contracts，移除 commands/domain 对全局 dto 的这段依赖，保持同一模块内契约闭环。

## Files to modify
- code/backend/internal/module/challenge/contracts/challenge_core.go
- code/backend/internal/module/challenge/domain/mappers.go
- code/backend/internal/module/challenge/application/commands/challenge_service.go
