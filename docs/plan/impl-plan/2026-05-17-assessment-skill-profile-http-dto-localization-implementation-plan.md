# assessment skill profile HTTP DTO 模块内化实现方案

## Objective

把 assessment 模块的 skill profile 查询输出与 HTTP response DTO 从全局 `internal/dto` 收回模块边界，保持现有外部 JSON 契约不变。

## Non-goals

- 不改 report、teacher AWD review 的 DTO 归属和输出结构
- 不改 skill profile 路由、权限、JSON 字段或时间格式
- 不改 `README.md`
- 不改共享架构文档或 API 设计文档
- 只有在 `code/backend/internal/dto/skill_profile.go` 实现后确认无引用时才删除它

## Inputs

- `code/backend/internal/module/assessment/api/http/handler.go`
- `code/backend/internal/module/assessment/api/http/response_types.go`
- `code/backend/internal/module/assessment/api/http/response_mapper.go`
- `code/backend/internal/module/assessment/application/queries/profile_service.go`
- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/assessment/contracts/contracts.go`
- `code/backend/internal/module/assessment/runtime/module.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## Current Boundary Summary

- `application/queries/profile_service.go` 直接返回全局 `dto.SkillProfileResp`
- `api/http/handler.go` 也直接依赖全局 `dto.SkillProfileResp`
- report 读取链通过 `ports.AssessmentProfileReader` 读取同一个全局 DTO
- recommendation 已经采用“`contracts` 输出 + `api/http` 本地 response DTO + mapper”的模块内化模式

## Task Slices

1. 建立 skill profile 模块内 query contract
   - 在 `assessment/contracts` 增加 `SkillProfile` 与 `SkillDimension`
   - `application/queries/profile_service.go` 改为返回 contract 类型
   - 验证：query 包不再 import `internal/dto`

2. 把 HTTP response DTO 收回 `assessment/api/http`
   - 在 `response_types.go` 增加本地 `SkillProfileResp` 与 `SkillDimension`
   - 在 `response_mapper.go` / `response_mapper_gen.go` 增加 contract -> HTTP DTO 映射
   - `handler.go` 改为返回本地 HTTP DTO
   - 验证：skill profile 外部 JSON 字段与原来一致

3. 调整装配与最小消费侧验证
   - `runtime/module.go` 让 report 继续使用现有 reader，不把本刀扩到 report 重构
   - `full_router_state_matrix_integration_test.go` 改为用模块内 HTTP DTO 解码 skill profile
   - 必要时补 query 测试，覆盖空画像或取消上下文等既有行为

4. 评估全局 DTO 删除条件
   - 搜索 `dto.SkillProfileResp`、`dto.SkillDimension`
   - 若仍有 report / command 路径引用，则保留 `internal/dto/skill_profile.go` 并在交付中明确原因

## Expected Changes

- `.harness/reuse-decisions/assessment-skill-profile-http-dto-localization.md`
- `docs/plan/impl-plan/2026-05-17-assessment-skill-profile-http-dto-localization-implementation-plan.md`
- `code/backend/internal/module/assessment/contracts/contracts.go`
- `code/backend/internal/module/assessment/api/http/handler.go`
- `code/backend/internal/module/assessment/api/http/response_types.go`
- `code/backend/internal/module/assessment/api/http/response_mapper.go`
- `code/backend/internal/module/assessment/api/http/response_mapper_gen.go`
- `code/backend/internal/module/assessment/application/queries/profile_service.go`
- `code/backend/internal/module/assessment/runtime/module.go`
- `code/backend/internal/module/assessment/application/queries/profile_service_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## Compatibility

- `GET /api/v1/users/me/skill-profile` 与教师查看学生 skill profile 的 JSON 结构保持不变
- `user_id`、`dimensions`、`updated_at` 字段名及其嵌套字段保持不变
- report 读取行为保持现状，不在本刀改变

## Validation

- `go test ./internal/module/assessment/...`
- `go test ./internal/app/... -run TestFullRouter_TeacherAccessAndRecommendationStateMatrix`

## Review Focus

- skill profile query / HTTP 是否已经不再依赖全局 `internal/dto` 类型
- HTTP 本地 DTO 是否完全对齐原有 JSON 契约
- 这次是否只切 skill profile 查询 / HTTP 边界，没有把 report 扩成额外重构
- `internal/dto/skill_profile.go` 是否真的还有剩余引用

## Rollback

- 如果 contract 与 HTTP DTO 分离导致装配或测试回归，先回退 `handler.go`、`profile_service.go` 与 mapper 改动，再按更小切片重新收口
