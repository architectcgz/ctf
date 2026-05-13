# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/practice/application/commands/submission_service.go`
- `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- `code/backend/internal/module/practice/application/commands/service_audit_test.go`
- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/application/commands/service_test.go`
- `code/backend/internal/module/practice/infrastructure/runtime_subject_repository.go`
- `code/backend/internal/module/practice/infrastructure/manual_review_repository.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/module/practice/infrastructure/runtime_subject_repository.go`
- `code/backend/internal/module/practice/infrastructure/manual_review_repository.go`
- `code/backend/internal/module/practice/infrastructure/score_query_repository.go`

## Decision

refactor_existing

## Reason

`submission_service.go` 剩下的 concrete 分两类：

- challenge not-found：已经有 `runtime_subject_repository` 可以复用，不需要再造第二个 challenge adapter
- solved submission not-found：还缺一个只负责 `FindCorrectSubmission` 的窄 adapter，把 raw repository 的 `gorm.ErrRecordNotFound` 映射成已有的 `ports.ErrPracticeSolvedSubmissionNotFound`

`instanceRepo.FindByUserAndChallenge` 这条不需要新 adapter。当前生产实现 `practice/infrastructure.Repository` 已经把 not-found 收口成 `nil, nil`，application 只要停止直接判断 `gorm.ErrRecordNotFound` 即可。

最小正确方案因此是：

- 新增一个 solved-submission adapter
- `SubmitFlag` 改成复用 `runtimeSubject` + 新 solved-submission adapter
- 动态 flag / solve grace 路径直接按 `instance == nil` 处理，不再让 application 了解 GORM sentinel

## Files to modify

- `code/backend/internal/module/practice/application/commands/submission_service.go`
- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/application/commands/service_test.go`
- `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- `code/backend/internal/module/practice/application/commands/service_audit_test.go`
- `code/backend/internal/module/practice/infrastructure/solved_submission_repository.go`
- `code/backend/internal/module/practice/infrastructure/solved_submission_repository_test.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-practice-submission-not-found-contract-phase5-slice21-implementation-plan.md`

## After implementation

- `practice` application surface 里的剩余 GORM concrete allowlist 会再少一条，`practice/application/commands/submission_service.go -> gorm.io/gorm` 应可删除
- 如果后续还要继续清 challenge / contest 里的 GORM concrete，可以继续复用这次“已有 sentinel + 局部 adapter + 复用 runtime subject”的模式，不需要再讨论错误分层约定
