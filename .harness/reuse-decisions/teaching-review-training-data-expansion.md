# Reuse Decision

## Change type
job

## Existing code searched
- `code/backend/cmd/seed-teaching-review-data/main.go`
- `code/backend/cmd/seed-demo-challenges/main.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/assessment/application/queries/teacher_awd_review_service.go`
- `code/backend/internal/module/assessment/infrastructure/awd_review_repository.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/student_review_service.go`
- `code/backend/internal/module/contest/testsupport/fixtures.go`

## Similar implementations found
- `code/backend/cmd/seed-teaching-review-data/main.go`
  - 已经具备教师复盘样本写入、缓存清理、推荐读取和归档读取的完整链路
- `code/backend/internal/module/assessment/infrastructure/awd_review_repository.go`
  - 明确教师 AWD 复盘实际读取 `teams`、`team_members`、`awd_team_services`、`awd_attack_logs`、`awd_traffic_events`
- `code/backend/internal/module/contest/testsupport/fixtures.go`
  - 提供了 AWD 相关 service / team / round fixture 的字段约定，可作为 seed 字段语义参考
- `code/backend/cmd/seed-demo-challenges/main.go`
  - 说明仓库内已有“开发库种子命令”模式，但它只负责题目，不负责训练行为样本

## Decision
extend_existing

## Reason
当前任务不是新增一种独立数据能力，而是把已有 `seed-teaching-review-data` 从“小样本演示”扩到“适配大题库且能支撑教师 AWD 复盘”的训练样本生成。现有命令已经掌握正确的数据落库边界、清理逻辑、推荐缓存清理和命令行摘要输出。继续扩展它，比再起一个并行命令更符合最小改动，也能避免两套 seed 行为长期分叉。

## Files to modify
- `code/backend/cmd/seed-teaching-review-data/main.go`
- `code/backend/cmd/seed-teaching-review-data/main_test.go`
- `docs/plan/impl-plan/2026-05-13-teaching-review-training-data-expansion-implementation-plan.md`
- `.harness/reuse-decisions/teaching-review-training-data-expansion.md`
