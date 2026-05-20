# Reuse Decision

## Change type
module public error contract cleanup

## Existing code searched

- `code/backend/internal/module/challenge/contracts/errors.go`
- `code/backend/internal/module/challenge/**/*.go`
- `code/backend/internal/module/practice/**/*.go`

## Similar implementations found

- `code/backend/internal/module/challenge/contracts/errors.go`
- `code/backend/internal/module/practice/application/commands/submission_service.go`

## Decision
refactor_existing

## Reason

`challengecontracts.ErrFlagIncorrect` 只在定义处存在，没有任何运行时调用点。当前答题链路里，“Flag 错误”不是异常，而是正常业务结果：

- `SubmitFlag(...)` 对错误答案返回 `SubmissionStatusIncorrect`
- 只有题目不存在、未发布、已解出、提交过频等才作为错误 contract 暴露

继续保留 `ErrFlagIncorrect` 会让 challenge 模块公开错误契约比真实行为更宽，因此应删除。

## Files to modify

- `code/backend/internal/module/challenge/contracts/errors.go`

## After implementation

- 若未来某条独立 API 需要把“错误答案”表达为错误契约，再由 challenge owner 在真实链路落地时重新引入。
