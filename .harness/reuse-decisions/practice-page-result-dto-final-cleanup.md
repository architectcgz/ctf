# Reuse Decision

## Change type

contract / handler / app-test / dto cleanup

## Existing code searched

- `code/backend/internal/dto/{common.go,common_test.go,challenge_test.go}`
- `code/backend/internal/module/practice/{contracts,api/http,application/commands}`
- `code/backend/internal/module/challenge/api/http/challenge_request_types.go`
- `code/backend/internal/app/practice_flow_integration_test.go`

## Similar implementations found

- `challenge/contracts.PageResult` 已作为模块内 application -> http 分页壳 owner
- `practice/contracts` 已持有 manual review query / response 契约
- `challenge/api/http` 已持有 `ConfigureFlagReq`

## Decision

refactor_existing

## Reason

`internal/dto/common.go` 现在只剩 `practice` manual review 分页链路在用，已经不是全局 owner；`challenge_test.go` 也只是验证已经迁到 `challenge/api/http` 的 `ConfigureFlagReq`。最小正确方案是把分页壳收回 `practice/contracts`，把请求绑定测试迁回 `challenge/api/http`，然后删除 `internal/dto` 的最后三个文件。

## Files to modify

- `.harness/reuse-decisions/practice-page-result-dto-final-cleanup.md`
- `docs/plan/impl-plan/2026-05-18-practice-page-result-dto-final-cleanup-implementation-plan.md`
- `code/backend/internal/module/practice/contracts/page_result.go`
- `code/backend/internal/module/practice/contracts/page_result_test.go`
- `code/backend/internal/module/practice/api/http/handler.go`
- `code/backend/internal/module/practice/application/commands/manual_review_service.go`
- `code/backend/internal/module/challenge/api/http/configure_flag_req_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/dto/common.go`
- `code/backend/internal/dto/common_test.go`
- `code/backend/internal/dto/challenge_test.go`

## After implementation

- `practice` manual review 分页链路不再依赖 `internal/dto`
- `ConfigureFlagReq` 测试回到 `challenge/api/http`
- `internal/dto` 目录清空
