# practice progress/timeline HTTP DTO 模块内化实现方案

## Objective

把 practice 模块的 progress/timeline HTTP DTO 从全局 `internal/dto` 收回 `practice` 模块边界，同时让 progress cache 和 query service contract 不再依赖全局 progress DTO。

## Non-goals

- 不迁移 `teaching_query` 的教师端 timeline DTO owner
- 不修改 practice 之外的 handler、service、repository 逻辑
- 不调整外部 HTTP JSON 字段、路由、分页参数和状态码
- 不扩展到 `teaching_query` 的 contract 重构

## Inputs

- `code/backend/internal/module/practice/api/http/handler.go`
- `code/backend/internal/module/practice/api/http/handler_progress_test.go`
- `code/backend/internal/module/practice/application/queries/progress_timeline_*.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/progress_cache.go`
- `code/backend/internal/dto/progress.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## Ownership Boundary

- `practice/api/http`
  - 负责：student practice progress/timeline 的 HTTP DTO 和 JSON 输出
  - 不负责：缓存持久化形状、repository 查询记录
- `practice/application/queries`
  - 负责：组装模块内 progress/timeline 查询结果，协调 cache 读写
  - 不负责：暴露全局 shared DTO 给外部模块复用
- `practice/ports`
  - 负责：定义 progress cache / query 的模块内 contract 和快照形状
  - 不负责：承担 HTTP JSON owner
- `internal/dto`
  - 负责：仍被其他模块共享消费的 DTO
  - 不负责：继续承载 practice student progress 的模块内 HTTP 响应

## Task Slices

1. 先写失败测试
   - 把 practice handler / query / ports 相关测试切到模块内 DTO 或模块内快照类型
   - 验证当前代码因类型仍在 `internal/dto` 而失败

2. 实现 practice 内部收口
   - 在 `practice/api/http` 新增 progress/timeline HTTP DTO
   - 在 `practice/ports` 新增 progress 快照与 timeline 响应 contract
   - 调整 query service、handler、progress cache 使用新 contract

3. 清理全局 DTO 文件
   - 搜索 `ProgressResp/TimelineResp` 全仓引用
   - 若 `progress.go` 仅承载 teaching/shared timeline 类型，则迁移 timeline 类型到新的 `internal/dto/timeline.go` 并删除 `internal/dto/progress.go`

4. 运行最小验证
   - `go test ./internal/module/practice/...`
   - 视影响面补跑直接相关的 `internal/app` 集成测试

## Expected Changes

- `.harness/reuse-decisions/practice-progress-timeline-http-dto-localization.md`
- `docs/plan/impl-plan/2026-05-17-practice-progress-timeline-http-dto-localization-implementation-plan.md`
- `code/backend/internal/module/practice/api/http/`
- `code/backend/internal/module/practice/application/queries/`
- `code/backend/internal/module/practice/ports/`
- `code/backend/internal/module/practice/infrastructure/progress_cache.go`
- `code/backend/internal/dto/progress.go`
- `code/backend/internal/dto/timeline.go`
- 必要时：`code/backend/internal/app/*practice*test.go`

## Compatibility

- `GET /api/v1/users/me/progress` 响应 JSON 保持不变
- `GET /api/v1/users/me/timeline` 响应 JSON 保持不变
- 教师端 timeline 的全局 DTO 继续保留在 `internal/dto`，但文件归属可调整

## Validation

- `cd code/backend && go test ./internal/module/practice/... -count=1 -timeout 3m`
- `cd code/backend && go test ./internal/app -run TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge -count=1 -timeout 5m`
- 如涉及教师端共享 timeline 文件移动，再补：
  - `cd code/backend && go test ./internal/app -run TestFullRouter_TeacherAccessAndRecommendationStateMatrix -count=1 -timeout 5m`

## Review Focus

- practice progress/timeline HTTP DTO 是否已经回到 `practice/api/http`
- progress cache / query contract 是否不再依赖 `internal/dto/progress.go`
- 外部 JSON 契约是否保持不变
- `internal/dto/progress.go` 是否真的已经可删，而不是只换了文件名但漏掉引用

## Rollback

- 如果 contract 调整导致 `practice` query/cache 接口过宽，可先保留模块内快照类型与 HTTP DTO 分离，再缩小映射范围
- 如果删除 `internal/dto/progress.go` 触发 teaching/shared timeline 回归，可先只迁移 practice 依赖并保留全局 timeline 文件，再做更小一步清理
