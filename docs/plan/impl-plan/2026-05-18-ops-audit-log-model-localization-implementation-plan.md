# ops audit log model 模块内化实现方案

## Objective

把 `internal/model/audit_log.go` 中的 `AuditLog` 持久化实体收回 `internal/module/ops/entity`，并把审计动作常量收口到共享 `internal/auditlog` 包，让审计日志的 owner、写入仓储和测试 schema 统一依赖正确的模块边界。

## Non-goals

- 不处理 `User`
- 不处理 `Submission`
- 不处理 `Contest` 与 `AWD*`
- 不改变审计表结构、动作值、查询语义或中间件行为

## Inputs

- `internal/model/audit_log.go`
- `internal/auditlog/*.go`
- `internal/module/ops/...`
- `internal/app/router_routes.go`
- `internal/module/auth/api/http/...`
- `.harness/reuse-decisions/ops-audit-log-model-localization.md`

## Ownership Evaluation

- owner 明确：`AuditLog` 的持久化 owner 是 `ops`
- landing zone 明确：`internal/module/ops/entity/audit_log.go`
- 共享语义可保留：其他模块继续消费动作常量，但来源改为 `internal/auditlog`
- 结构收敛目标明确：删除旧全局模型文件

## Task slices

1. 新增 `ops/entity/audit_log.go` 与 `auditlog/actions.go`
2. 更新 ops ports / repository / command service / risk repository
3. 更新 app、auth、runtime、middleware 和查询仓储中的审计动作常量引用
4. 更新 app / module integration tests 的 schema、seed 和断言
5. 删除 `internal/model/audit_log.go`

## Validation

- `go test ./internal/module/ops/... -count=1`
- `go test ./internal/module/auth/api/http -run 'TestHTTP_FailedLoginIsRecordedInAuditLog|TestHTTP_LogoutRevokesSessionAndAdminCanQueryAuditLogs' -count=1`
- `go test ./internal/app -run 'TestFullRouter_AdminOpsAndNotificationStateMatrix|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge|TestFullRouter_ChallengeWriteupSubmissionVisibilityAndAuditStateMatrix' -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `AuditLog` 实体和动作常量是否完全收敛到 `ops/entity`
- app / middleware / runtime / auth 是否没有残留旧全局引用
- app / module tests 的 migrate、seed 和断言是否同步切换

## Rollback

本刀无 schema 变更，如有回归可直接恢复到 `internal/model/audit_log.go`。
