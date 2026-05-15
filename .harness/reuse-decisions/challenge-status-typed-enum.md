# Reuse Decision

## Change type
service / mapper

## Existing code searched
- code/backend/internal/model/challenge.go
- code/backend/internal/model/awd_challenge.go
- code/backend/internal/module/challenge/application/queries/challenge_service.go
- code/backend/internal/module/challenge/domain/response_mapper_goverter.go
- code/backend/internal/module/challenge/domain/response_mapper_goverter_gen.go
- code/backend/internal/dto/challenge.go
- code/frontend/src/api/contracts.ts

## Similar implementations found
- code/backend/internal/model/awd_challenge.go
  - 已经把 AWD 题目状态收口为 `type AWDChallengeStatus string` + typed const，适合作为普通题目状态的现有模式。
- code/backend/internal/module/challenge/application/queries/challenge_service.go
  - 已经是学员访问题目时的状态判断 owner，适合继续在原服务里收口不可访问状态分支。
- code/backend/internal/module/challenge/domain/response_mapper_goverter_gen.go
  - 已经承担 `model.Challenge -> dto.ChallengeResp` 的字符串输出，不需要改 API contract，只需要保持这里做显式 `string(...)` 转换。

## Decision
refactor_existing

## Reason
这次不是新增题目状态能力，也不是引入新的 DTO / 数据库存储形态，而是把现有普通题目状态对齐到仓库里已经存在的 AWD typed status 模式。直接重构现有 `model.Challenge`、复用现有查询服务 owner 和现有 mapper 输出链路，改动最小，也能让 Go 编译期更早暴露状态误用，同时保持数据库和前端接口仍然使用原有字符串值。

## Files to modify
- .harness/reuse-decisions/challenge-status-typed-enum.md
- code/backend/internal/model/challenge.go
- code/backend/internal/module/challenge/application/queries/challenge_service.go
- code/backend/internal/module/challenge/domain/response_mapper_goverter_gen.go
- code/backend/internal/module/challenge/application/commands/challenge_service_test.go
- code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go
- code/backend/internal/module/challenge/infrastructure/repository_test.go
- code/backend/internal/app/full_router_state_matrix_integration_test.go
