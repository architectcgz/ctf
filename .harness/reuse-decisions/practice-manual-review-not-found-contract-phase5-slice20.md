# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/practice/application/commands/manual_review_service.go`
- `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/application/commands/service_test.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/runtime_subject_repository.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/module/practice/infrastructure/runtime_subject_repository.go`
- `code/backend/internal/module/practice/infrastructure/score_query_repository.go`
- `code/backend/internal/module/practice/infrastructure/contest_scope_repository.go`

## Decision

refactor_existing

## Reason

这次 debt 不是 `practice` 缺少人工评阅功能，而是 `manual_review_service.go` 直接知道多种 `gorm.ErrRecordNotFound` 语义：人工评阅记录、已通过提交、教师用户和 challenge lookup。最小正确方案仍然沿用 phase5 已验证的“sentinel 在 ports、映射在 infrastructure adapter、application 只看模块内语义”的模式，但把 owner 保持在两个窄点：

- `practice` 自己新增 manual review adapter，负责把 raw practice repository 的人工评阅相关 not-found 映射成模块内 sentinel
- challenge not-found 不再在 manual review service 里直接看 GORM，而是复用现有 `runtime_subject_repository` adapter 提供的 `ErrPracticeChallengeNotFound`

这样可以把本 slice 收在 `manual_review_service.go` 和其专属 adapter，不去改 raw repository 的全局错误语义，也不提前重写 `submission_service.go`。

## Files to modify

- `code/backend/internal/module/practice/application/commands/manual_review_service.go`
- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/application/commands/service_test.go`
- `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/infrastructure/manual_review_repository.go`
- `code/backend/internal/module/practice/infrastructure/manual_review_repository_test.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-practice-manual-review-not-found-contract-phase5-slice20-implementation-plan.md`

## After implementation

- 后续收 `practice/application/commands/submission_service.go` 时，可继续复用这次新增的 `ErrPracticeSolvedSubmissionNotFound` 与 challenge not-found adapter，不必重新从 raw repo 直接判断 GORM sentinel
- 如果后续发现 `practice` 里还有更多 user lookup not-found 需要统一收口，再考虑抽成更通用的 user lookup adapter；这次不提前做全局化
