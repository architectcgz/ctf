# assessment skill profile command/report DTO 模块内化实现方案

## Objective

把 `assessment` 剩余 command / report 链上的 `dto.SkillProfileResp`、`dto.SkillDimension` 收回 `assessment/contracts`，完成 `internal/dto/skill_profile.go` 的退出。

## Non-goals

- 不改 skill profile HTTP 路由、权限、JSON 字段
- 不改 recommendation、teacher AWD review、report export 外部接口
- 不顺手重构整个 report 模块

## Inputs

- `docs/plan/impl-plan/2026-05-17-assessment-skill-profile-http-dto-localization-implementation-plan.md`
- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/domain/profile.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/dto/skill_profile.go`

## Task Slices

1. 用 `assessment/contracts` 承接剩余 skill profile 形状
   - 为 `SkillDimension` / `SkillProfile` 补齐 report/export 需要的 tag
   - 统一 domain builder 只输出 contract

2. 收口 profile command service 与 profile reader port
   - command service 改为返回 `assessmentcontracts.SkillProfile`
   - `AssessmentProfileReader` 改为暴露 contract

3. 收口 report service 与相关测试
   - report 内部 data 结构改为依赖 `assessmentcontracts.SkillDimension` / `SkillProfile`
   - 保持导出 JSON / PDF 文本口径不变

4. 删除废弃全局 DTO
   - 若 `internal/dto/skill_profile.go` 无剩余引用则删除

## Expected Changes

- `code/backend/internal/module/assessment/contracts/contracts.go`
- `code/backend/internal/module/assessment/domain/profile.go`
- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/dto/skill_profile.go`

## Validation

- `go test ./internal/module/assessment/... -count=1`
- `go test ./internal/app -run TestFullRouter_TeacherAccessAndRecommendationStateMatrix -count=1`

## Review Focus

- skill profile 是否已经只在 `assessment/contracts` 与 `assessment/api/http` 之间流动
- report / export 是否没有因内部类型切换而改坏导出内容
- `internal/dto/skill_profile.go` 是否已经完全退出

