# Reuse Decision

## Change type

command / test / rule-fix

## Existing code searched

- `code/backend/cmd/seed-teaching-review-data/main.go`
- `code/backend/cmd/seed-teaching-review-data/main_test.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `code/backend/internal/module/assessment/domain/report.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository.go`
- `code/backend/internal/module/teaching_query/application/queries/student_review_service.go`
- `code/backend/internal/module/teaching_query/ports/query.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
- `code/backend/internal/teaching/advice/advice.go`
- `code/backend/internal/teaching/evidence/evidence.go`

## Similar implementations found

- `code/backend/cmd/seed-teaching-review-data/main.go`
  - 已经是教学复盘样本写入、归档读取和命令行核对输出的唯一入口
- `code/backend/internal/module/assessment/application/commands/report_service.go`
  - 已经是个人复盘归档事实快照和观察项装配的唯一 owner
- `code/backend/internal/module/assessment/infrastructure/report_repository.go`
  - 已经负责复盘归档证据事件的数据库读取与装配
- `code/backend/internal/teaching/evidence/evidence.go`
  - 已经是训练证据事件类型和元数据的唯一构造入口

## Decision

extend_existing

## Reason

这次不是新增新的复盘能力，而是把现有复盘链路里的 AWD 口径收紧到“学生本人提交的 AWD 攻击记录”，并在现有 seed 命令里补足命中、重复失败、单次试探这些样本边界。继续复用现有 seed 命令、归档快照构建和建议规则层，可以避免再起并行的样本脚本或第二套 AWD 判断逻辑。

## Files to modify

- `code/backend/cmd/seed-teaching-review-data/main.go`
- `code/backend/cmd/seed-teaching-review-data/main_test.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `code/backend/internal/module/assessment/domain/report.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository.go`
- `code/backend/internal/module/teaching_query/application/queries/student_review_service.go`
- `code/backend/internal/module/teaching_query/ports/query.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
- `code/backend/internal/teaching/advice/advice.go`
- `code/backend/internal/teaching/evidence/evidence.go`
