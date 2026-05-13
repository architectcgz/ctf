# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/contest/application/commands/submission_submit_validation.go`
- `code/backend/internal/module/contest/application/commands/submission_service.go`
- `code/backend/internal/module/contest/application/commands/submission_validation.go`
- `code/backend/internal/module/contest/application/commands/participation_error_contract_test.go`
- `code/backend/internal/module/contest/infrastructure/submission_lookup_repository.go`
- `code/backend/internal/module/contest/infrastructure/submission_registration_repository.go`
- `code/backend/internal/module/contest/infrastructure/submission_registration_repository_test.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/submission.go`
- `code/backend/internal/module/practice/infrastructure/contest_scope_repository.go`
- `code/backend/internal/module/practice/ports/ports.go`

## Similar implementations found

- `code/backend/internal/module/practice/infrastructure/contest_scope_repository.go`
- `code/backend/internal/module/challenge/infrastructure/awd_challenge_repository.go`
- `code/backend/internal/module/contest/infrastructure/submission_registration_repository.go`

## Decision

refactor_existing

## Reason

`contest` 这条 submission wiring 已经有一个窄 `SubmissionRegistrationRepository` adapter，当前只负责 registration not-found。`submission_submit_validation.go` 里残留的 concrete 依赖，实质上是同一个 adapter 还没有继续承接 submission challenge lookup：

- `FindContestChallenge` 的 not-found 需要映射成“题目不在竞赛中”
- `FindChallengeByID` 的 not-found 需要映射成“题目不存在”

最小正确收口方式不是再造第二个 submission adapter，而是沿用现有 `SubmissionRegistrationRepository`，在 `contest/ports/submission.go` 增加 submission 专用 sentinel，并把两个 challenge lookup 的 `gorm.ErrRecordNotFound` 一并收口。`SubmissionService` 只消费 sentinel，runtime 继续注入同一个 adapter。

这次不扩到：

- `architecture_allowlist_test.go`
- shared allowlist / shared facts docs
- AWD support / jobs / query
- challenge 模块代码

## Files to modify

- `code/backend/internal/module/contest/ports/submission.go`
- `code/backend/internal/module/contest/application/commands/submission_submit_validation.go`
- `code/backend/internal/module/contest/application/commands/participation_error_contract_test.go`
- `code/backend/internal/module/contest/infrastructure/submission_registration_repository.go`
- `code/backend/internal/module/contest/infrastructure/submission_registration_repository_test.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `docs/plan/impl-plan/2026-05-13-contest-submission-challenge-lookup-not-found-contract-phase5-slice30-implementation-plan.md`

## After implementation

- `contest/application/commands/submission_submit_validation.go` 应不再直接依赖 `gorm.io/gorm`
- phase5 shared allowlist / shared facts 仍由后续更大边界统一更新；本 slice 只交付代码侧 contract 收口证据
