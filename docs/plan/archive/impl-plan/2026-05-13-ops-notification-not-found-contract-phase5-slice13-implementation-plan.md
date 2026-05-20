# Ops Notification Not Found Contract Phase 5 Slice 13 Implementation Plan

## Objective

继续 phase 5 收窄 `ops` application concrete allowlist，把 notification command 的 GORM not-found 判断下沉到模块 repo contract：

- 去掉 `ops/application/commands/notification_service.go -> gorm.io/gorm`
- 保持通知不存在时返回 `ErrNotificationNotFound` 的现有行为不变

## Non-goals

- 不改 notification 表结构、批量通知逻辑或 websocket 广播逻辑
- 不处理其他模块剩余的 GORM / HTTP concrete allowlist
- 不重写 notification query repository 或分页 contract

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/ops/application/commands/notification_service.go`
- `code/backend/internal/module/ops/application/commands/notification_service_test.go`
- `code/backend/internal/module/ops/ports/notification.go`
- `code/backend/internal/module/ops/infrastructure/notification_repository.go`

## Current Baseline

- `notification_service.go` 当前直接 import `gorm.io/gorm`
- application 只是为了把 `gorm.ErrRecordNotFound` 映射成 `errcode.ErrNotificationNotFound`
- allowlist 当前保留：
  - `ops/application/commands/notification_service.go -> gorm.io/gorm`

## Chosen Direction

把 notification not-found 语义收口成 `ops` 自己的 repository contract：

1. 在 `ops/ports` 暴露模块内 `ErrNotificationNotFound`
2. `ops/infrastructure/notification_repository.go` 负责把 `gorm.ErrRecordNotFound` 映射成该 sentinel
3. `notification_service.go` 只判断 `opsports.ErrNotificationNotFound`，不再知道 GORM sentinel
4. 保持 runtime wiring 不变，不引入新的 repository 或 service 抽象

## Ownership Boundary

- `ops/application/commands/notification_service.go`
  - 负责：通知业务编排和错误码映射
  - 不负责：知道 `gorm.ErrRecordNotFound` 或 ORM 具体错误类型
- `ops/infrastructure/notification_repository.go`
  - 负责：把 persistence 层 not-found 映射成 `ops` 自己的 repo 契约
  - 不负责：决定 HTTP 错误码或 websocket 响应行为

## Change Surface

- Add: `.harness/reuse-decisions/ops-notification-not-found-contract-phase5-slice13.md`
- Add: `docs/plan/impl-plan/2026-05-13-ops-notification-not-found-contract-phase5-slice13-implementation-plan.md`
- Modify: `code/backend/internal/module/ops/ports/notification.go`
- Modify: `code/backend/internal/module/ops/application/commands/notification_service.go`
- Modify: `code/backend/internal/module/ops/application/commands/notification_service_test.go`
- Modify: `code/backend/internal/module/ops/infrastructure/notification_repository.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

- [ ] Slice 1: 收口 notification repo not-found contract
  - 目标：notification command 不再 import `gorm.io/gorm`，not-found 行为保持一致
  - 验证：
    - `cd code/backend && go test ./internal/module/ops/application/commands -run 'NotificationService' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/ops/runtime -run '^$' -count=1 -timeout 5m`
  - Review focus：application 是否已经不再知道 GORM sentinel；not-found 仍然稳定映射到 `ErrNotificationNotFound`

- [ ] Slice 2: 删除 allowlist 并同步文档
  - 目标：删除 `ops/application/commands/notification_service.go` 的 GORM allowlist，并更新 phase 5 当前事实
  - 验证：
    - `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus：只删除这条实际收口的 allowlist；文档准确描述 `ops` 当前 concrete allowlist 状态

## Risks

- 如果 infrastructure 没有稳定映射 not-found，`MarkAsRead` 会把 404 变成 500
- 如果 application 仍保留 GORM import，allowlist 不会真正收口

## Verification Plan

1. `cd code/backend && go test ./internal/module/ops/application/commands -run 'NotificationService' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/ops/runtime -run '^$' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
4. `python3 scripts/check-docs-consistency.py`
5. `bash scripts/check-consistency.sh`
6. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：notification command 继续持有业务编排和错误码映射语义，persistence sentinel 映射落回 infrastructure
- reuse point 明确：沿用 phase5 已验证的“application 只看模块自己的契约，具体库错误留在 infrastructure”模式
- 这刀同时解决行为与结构：保持 not-found 用户行为不变，同时删掉 `ops` command surface 的 concrete GORM import
