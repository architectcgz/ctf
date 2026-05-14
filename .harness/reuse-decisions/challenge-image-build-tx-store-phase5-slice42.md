# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/challenge/application/commands/image_build_service.go`
- `code/backend/internal/module/challenge/application/commands/image_build_service_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- `code/backend/internal/module/challenge/infrastructure/image_command_repository.go`
- `code/backend/internal/module/challenge/infrastructure/image_repository.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/architecture_allowlist_test.go`

## Similar implementations found

- `code/backend/internal/module/challenge/infrastructure/image_command_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_http_runtime_adapter.go`
- `code/backend/internal/module/practice/infrastructure/manual_review_repository.go`

## Decision

refactor_existing

## Reason

这次 debt 的核心不是 image build 缺少 repository，而是 `image_build_service.go` 把 transaction-scoped image lookup / create / update / build-job create 直接写成了 GORM 细节。

当前 raw image repository 已经承担了非 tx 路径的大部分 image command owner，且 `challenge/infrastructure/image_command_repository.go` 已经把 `gorm.ErrRecordNotFound` 收口成 `ports.ErrChallengeImageNotFound`。因此不需要再新造一套全局 image repo，也不应该把 GORM 依赖简单平移到另一个 application 文件。

最合理的中间收口是：

- `image_build_service.go` 自己只依赖一个 transaction-scoped 窄接口，表达 image build 在事务里真正需要的能力
- 具体 `*gorm.DB` 包装留在已经允许 transaction concrete 的 import service 文件中，作为局部 tx wrapper 复用
- 非 tx 路径统一消费 `ports.ErrChallengeImageNotFound`

这样能先删掉 `image_build_service.go -> gorm.io/gorm`，并给后续 challenge import / package revision 的 tx store 迁移提供同一模式：application service 只写语义，ORM sentinel 和 `Unscoped/Updates/Create` 细节继续压在更外层。

## Files to modify

- `code/backend/internal/module/challenge/application/commands/image_build_service.go`
- `code/backend/internal/module/challenge/application/commands/image_build_service_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/plan/impl-plan/2026-05-14-challenge-image-build-tx-store-phase5-slice42-implementation-plan.md`

## After implementation

- `challenge/application/commands/image_build_service.go -> gorm.io/gorm` allowlist 应可删除
- challenge import / AWD import 还会保留 transaction owner，但 image build 相关 concrete 已经从 application service 主体里抽离
- 后续若继续迁 `challenge_import_service.go`、`awd_challenge_import_service.go`、`challenge_package_revision_service.go`，优先复用这次“transaction-scoped 窄接口 + 局部 tx wrapper + ports sentinel”的模式
