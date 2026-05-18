# assessment report model 模块内化实现方案

## Objective

把 `internal/model/report.go` 中的 `Report` 持久化实体，以及与其强绑定的 report type / format / status 常量收回 `internal/module/assessment/entity`，让 `assessment` 报告导出链路和 app 集成测试直接依赖模块内实体。

## Non-goals

- 不处理 `skill_profile`
- 不迁移 `Dimension*` 常量
- 不改报告导出行为、表结构、API shape

## Inputs

- `internal/model/report.go`
- `internal/module/assessment/...`
- `internal/app/full_router*_integration_test.go`
- `.harness/reuse-decisions/assessment-report-model-localization.md`

## Ownership Evaluation

- owner 明确：`Report` 只由 `assessment` 报表链路和 app 测试消费
- landing zone 明确：`internal/module/assessment/entity/report.go`
- 非目标明确：不连带拆 `skill_profile` 和全局维度常量
- 结构收敛目标明确：删除旧 `internal/model/report.go`

## Task slices

1. 新增 `assessment/entity/report.go`
   - 保留 GORM 字段、表名和 report 常量

2. 更新 `assessment` 层引用
   - `ports`、`infrastructure`、`application/commands`、`domain`
   - `goverter` mapper 输入类型切换到 `assessment/entity`

3. 更新测试与 app 集成测试
   - `assessment` command / repository tests
   - `internal/app/full_router*_integration_test.go`

4. 删除旧全局实体
   - `internal/model/report.go`

## Expected files

- `code/backend/internal/module/assessment/entity/report.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository_test.go`
- `code/backend/internal/module/assessment/application/commands/*report*`
- `code/backend/internal/module/assessment/application/commands/response_mapper_goverter*.go`
- `code/backend/internal/module/assessment/domain/report.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## Compatibility impact

- 数据库兼容：无 schema 变更
- API 兼容：无请求/响应变更
- 代码边界：报告实体不再由全局 `internal/model` 暴露

## Validation

- `go generate ./internal/module/assessment/application/commands`
- `go test ./internal/module/assessment/... -count=1`
- `go test ./internal/app -run 'TestFullRouter_ReportPreviewAndDownloadStateMatrix' -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- `Report` owner 是否完全收敛到 `assessment/entity`
- report type / format / status 常量是否没有残留全局引用
- `goverter` 生成结果是否只改类型、不改行为
- app 测试与 assessment tests 是否没有漏回 `model.Report`

## Rollback

如迁移后出现回归，可恢复 `assessment/entity/report.go` 到 `internal/model/report.go`，因为本刀不涉及 schema 变更，回退只需要代码层调整。
