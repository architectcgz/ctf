# Reuse Decision

## Change type

service / port / infrastructure / composition

## Existing code searched

- `code/backend/internal/module/contest/application/commands/awd_preview_runtime_support.go`
- `code/backend/internal/module/contest/application/commands/awd_service.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- `code/backend/internal/module/contest/infrastructure/submission_registration_repository.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/ports/awd.go`
- `code/backend/internal/module/challenge/infrastructure/awd_challenge_repository.go`
- `code/backend/internal/module/challenge/infrastructure/image_query_repository.go`

## Similar implementations found

- `code/backend/internal/module/contest/infrastructure/submission_registration_repository.go`
- `code/backend/internal/module/challenge/infrastructure/awd_challenge_repository.go`
- `code/backend/internal/module/challenge/infrastructure/image_query_repository.go`

## Decision

refactor_existing

## Reason

`awd_preview_runtime_support.go` 当前同时从两个下游读取 runtime lookup：

- `AWDChallengeQueryRepository.FindAWDChallengeByID`
- `ImageStore.FindByID`

application 只需要知道“preview challenge 不存在”或“preview image 不存在”，不应该继续知道底层是不是 `gorm.ErrRecordNotFound`。同时这条链路有一个必须保留的差异化行为：

- 显式 URL 下，preview challenge/runtime 定义缺失可以继续降级放行
- 显式 URL 下，preview image 缺失或不可用仍要阻断

因此最小正确方案是：

- 在 `contest/ports/awd.go` 新增 contest 专用 sentinel
- 在 `contest/infrastructure` 新增窄 preview runtime lookup adapter，把两个下游的 `gorm.ErrRecordNotFound` 分别映射成 contest sentinel
- `awd_preview_runtime_support.go` 只消费 contest sentinel，再按既有逻辑决定 `errcode` 或降级
- `contest/runtime/module.go` 负责把 adapter 注入 `AWDService`

## Files to modify

- `code/backend/internal/module/contest/ports/awd.go`
- `code/backend/internal/module/contest/application/commands/awd_preview_runtime_support.go`
- `code/backend/internal/module/contest/application/commands/awd_preview_runtime_support_contract_test.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository.go`
- `code/backend/internal/module/contest/infrastructure/awd_preview_runtime_lookup_repository_test.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `docs/plan/impl-plan/2026-05-13-contest-awd-preview-runtime-lookup-not-found-phase5-slice32-implementation-plan.md`

## After implementation

- `contest/application/commands/awd_preview_runtime_support.go` 应不再直接依赖 `gorm.io/gorm`
- 显式 URL 下的 preview challenge not-found 降级应保持；preview image not-found 仍应保留阻断
- shared allowlist / shared facts docs 继续留给后续更大边界统一处理；本 slice 只交付代码侧 contract 收口证据
