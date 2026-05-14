# Reuse Decision

## Change type
service

## Existing code searched
- `code/backend/internal/teaching/advice/advice.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/challenge/infrastructure/repository.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service_test.go`
- `docs/architecture/features/教学复盘建议生成架构.md`

## Similar implementations found
- `code/backend/internal/teaching/advice/advice.go`
  - 已经是班级复盘建议、个人观察和推荐理由的共享规则 owner
- `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service.go`
  - 已经负责把班级 advice item 组装成教师侧 `TeacherClassReviewResp`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
  - 已经负责把 review archive 事实装配成共享规则层可消费的 `StudentFactSnapshot`
- `code/backend/internal/module/challenge/infrastructure/repository.go`
  - 已经是推荐候选题查询 owner，适合在原查询结果上补“实际命中维度”
- `docs/architecture/features/教学复盘建议生成架构.md`
  - 已经是教学复盘建议规则层的最终事实源

## Decision
extend_existing

## Reason
这次不是新增新的教学建议模块，而是继续在现有链路里收紧几类已经影响教师判断质量的规则：个人复盘弱项阈值过松、归档提交统计口径混杂、观察结论文案过于模板化，以及班级维度聚类过早放大。继续扩展 `internal/teaching/advice` 与 `report_service.go` 两个既有 owner，并同步更新现有设计 / 架构文档，符合最小改动，也能保持“建议层唯一 owner”不被打散。

## Files to modify
- `code/backend/internal/teaching/advice/advice.go`
- `code/backend/internal/teaching/advice/advice_test.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `docs/design/教学复盘建议优化方案.md`
- `docs/architecture/features/教学复盘建议生成架构.md`
- `docs/plan/impl-plan/2026-05-13-teaching-review-strategy-tuning-implementation-plan.md`
