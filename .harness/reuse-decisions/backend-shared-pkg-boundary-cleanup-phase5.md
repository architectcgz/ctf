# Reuse Decision

## Change type
shared kernel / contract / http transport / module boundary

## Existing code searched

- `code/backend/pkg/errcode/`
- `code/backend/internal/httpresponse/`
- `code/backend/internal/module/*/{contracts,application,api}/`
- `code/backend/internal/app/`

## Similar implementations found

- `code/backend/internal/httpresponse/response.go`
- `code/backend/internal/module/challenge/contracts/`
- `code/backend/internal/module/contest/contracts/`
- `code/backend/internal/module/instance/contracts/`
- `code/backend/internal/module/auth/contracts/`

## Decision
refactor_existing

## Reason

`pkg/errcode` 里混了三类 owner：

- 共享公共错误类型与平台级错误（例如 `invalid params`、`forbidden`、`internal`）
- 模块自有的公开错误契约（例如 challenge image / contest team / instance proxy ticket）
- HTTP transport 才真正消费的状态码映射

如果只把整个包平移到 `internal/*`，只是把共享桶换了个路径，owner 仍然不清。更合理的落点是：

- 共享错误类型与平台级错误收回 `internal/apperror`
- challenge / contest / instance / ops / auth 各自拥有的公开错误，收回各模块 `contracts`
- `HTTPStatus` 不再暴露给 application / domain 调用点，由 `httpresponse` 统一消费

## Files to modify

- `code/backend/pkg/errcode/*.go`
- `code/backend/internal/httpresponse/*.go`
- `code/backend/internal/app/**/*.go`
- `code/backend/internal/middleware/*.go`
- `code/backend/internal/module/**/{api,application,runtime}/**/*.go`
- `code/backend/internal/module/{challenge,contest,instance,auth,ops}/contracts/**/*.go`
- `code/backend/internal/module/auth/infrastructure/token_service_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_image_service_support.go`
- `code/backend/internal/module/contest/application/queries/awd_service_list_query.go`
- `code/backend/internal/module/contest/application/queries/contest_awd_service_query.go`
- `code/backend/internal/module/identity/application/commands/profile_service.go`

## After implementation

- 若这套“共享错误内核 + 模块公开错误契约 + transport 映射”的模式稳定，可后续沉淀进 `harness/reuse/index.yaml`。
