# Reuse Decision

## Change type

command / report / contract

## Existing code searched

- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/contracts/contracts.go`
- `code/backend/internal/module/assessment/domain/profile.go`
- `code/backend/internal/dto/skill_profile.go`

## Similar implementations found

- `assessment` recommendation 已经使用 `contracts -> api/http` 的 owner 结构
- skill profile 查询 / HTTP 已经先完成了一半模块内化

## Decision

extend_existing

## Reason

这次不是新增新的 skill profile 结构，而是把 HTTP 之外剩余的 command / report 读链继续补齐到既有 `assessment/contracts`。最小正确方案是复用现有 contract，不再保留第二份全局 DTO。

## Files to modify

- `.harness/reuse-decisions/assessment-skill-profile-command-report-dto-localization.md`
- `docs/plan/impl-plan/2026-05-17-assessment-skill-profile-command-report-dto-localization-implementation-plan.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/reviews/backend/2026-05-17-assessment-skill-profile-command-report-dto-localization-review.md`
- `code/backend/internal/module/assessment/contracts/contracts.go`
- `code/backend/internal/module/assessment/domain/profile.go`
- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/dto/skill_profile.go`
