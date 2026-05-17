# assessment recommendation 全局 DTO 删除实现方案

## Objective

删除已经无调用点的 `internal/dto/recommendation.go`，完成 recommendation 这条链从全局 DTO 桶退出的最后一步。

## Non-goals

- 不迁移其他 `assessment` DTO
- 不调整教师端 recommendation DTO 归属
- 不修改任何外部 JSON 字段、路由或查询参数

## Inputs

- `code/backend/internal/dto/recommendation.go`
- `code/backend/internal/module/assessment/api/http/response_types.go`
- `code/backend/internal/module/assessment/api/http/response_mapper.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `docs/architecture/backend/04-api-design.md`

## Task Slices

1. 确认无剩余调用点
   - 搜索 `dto.RecommendationResp`、`dto.ChallengeRecommendation`、`dto.RecommendationWeakDimension`
   - 验证：代码引用为零

2. 删除旧全局 DTO 文件
   - 删除 `code/backend/internal/dto/recommendation.go`
   - 验证：编译与测试不受影响

3. 更新事实说明
   - 在 API 设计文档补一条“全局文件已移除，仅 owner 收口”的说明
   - 验证：`bash scripts/check-workflow-complete.sh`

## Expected Changes

- `code/backend/internal/dto/recommendation.go`
- `docs/architecture/backend/04-api-design.md`
- `.harness/reuse-decisions/assessment-recommendation-global-dto-removal.md`
- `docs/plan/impl-plan/2026-05-17-assessment-recommendation-global-dto-removal-implementation-plan.md`

## Compatibility

- recommendation 对外 HTTP 契约不变
- 变化只在内部文件归属与死代码删除

## Validation

- `go test ./internal/module/assessment/...`
- `go test ./internal/app/... -run TestFullRouter_TeacherAccessAndRecommendationStateMatrix`
- `bash scripts/check-workflow-complete.sh`

## Review Focus

- 是否真的没有剩余全局 recommendation DTO 调用点
- 删除是否只移除死文件，不引入契约变化

## Rollback

- 如果删除后发现还有隐含调用点，可先恢复 `internal/dto/recommendation.go` 再补剩余迁移
