# ops notification model 模块内化实现方案

## Objective

把 `internal/model` 中仅由 `ops` 模块持有的持久化实体 `Notification`、`NotificationBatch` 以及通知类型常量收回 `internal/module/ops/entity`，让 `ops` 的 command / query / repository / HTTP 测试 / app 集成测试直接依赖模块内实体。

## Non-goals

- 不处理 `audit_log`
- 不扩到 `user`、`contest`、`team`、`submission`、`instance` 等多 owner 核心实体
- 不改变通知表结构、表名、字段 JSON shape 或业务行为

## Inputs

- `internal/model/notification.go`
- `internal/model/notification_batch.go`
- `internal/module/ops/...`
- `.harness/reuse-decisions/ops-notification-model-localization.md`

## Ownership Evaluation

- owner 明确：`Notification` / `NotificationBatch` 只被 `ops` 模块和 app 测试消费
- landing zone 明确：GORM 持久化实体进入 `internal/module/ops/entity`
- 非目标明确：不借这刀引入新的共享层，不处理 `audit_log`
- 结构收敛目标明确：删除旧全局实体文件，避免留下“先兼容、后迁移”的双轨结构

## Task slices

1. 新增 `internal/module/ops/entity`
   - 放入 `notification.go`、`notification_batch.go`
   - 保持表名、字段和 GORM tag 不变

2. 更新 `ops` 模块引用
   - `ports`、`infrastructure`、`application/commands`、`application/queries`
   - `goverter` mapper 输入类型切换到 `ops/entity`

3. 更新测试与 app 集成测试
   - `ops` 模块单测 / HTTP integration test
   - `internal/app/full_router*_integration_test.go`

4. 清理旧全局实体
   - 删除 `internal/model/notification.go`
   - 删除 `internal/model/notification_batch.go`

## Expected files

- `code/backend/internal/module/ops/entity/*.go`
- `code/backend/internal/module/ops/ports/notification.go`
- `code/backend/internal/module/ops/infrastructure/notification_repository.go`
- `code/backend/internal/module/ops/application/commands/*`
- `code/backend/internal/module/ops/application/queries/*`
- `code/backend/internal/module/ops/api/http/notification_http_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## Compatibility impact

- 数据库兼容：无 schema 变更
- API 兼容：无响应或请求 shape 变更
- 代码边界：`notification` 持久化实体不再从全局 `internal/model` 暴露

## Validation

- `go generate ./internal/module/ops/application/commands ./internal/module/ops/application/queries`
- `go test ./internal/module/ops/... -count=1`
- `go test ./internal/app -run 'TestFullRouter_AdminOpsAndNotificationStateMatrix' -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- owner 是否彻底收敛到 `ops/entity`
- app test / module test 是否没有残留 `model.Notification*`
- mapper 生成结果是否只做类型切换，没有引入行为差异
- 仓储里是否只保留对真正共享实体 `model.User` 的依赖

## Rollback

如果迁移后出现编译或行为回归，可先恢复 `ops/entity` 到旧的 `internal/model` 引用，再按更小切片重新迁移；由于没有 schema 变更，回退只涉及代码层。
