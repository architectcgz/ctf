# teaching_query 复用 identity 用户查询实现计划

> 状态：Draft
> 事实源：`code/backend/internal/app/composition/identity_module.go`、`code/backend/internal/module/teaching_query/`
> 替代：无

## 1. 目标

- 让 `teaching_query` 的基础用户 lookup 复用 `identity.Users`
- 保留教师目录、班级洞察、复盘证据等聚合 SQL 在 `teaching_query`
- 不把当前查询改成跨模块二次分页或二次排序

## 2. 范围

### 包含

- `BuildTeachingQueryModule(...)` 注入 `identity.Users`
- `teaching_query` runtime 和 query service 拆分“用户 lookup”与“聚合 repo”依赖
- 删除 `teaching_query/infrastructure/repository.go` 中不再需要的 `FindUserByID`
- 同步 teaching_query 测试、app 层装配测试和当前架构事实文档

### 不包含

- 不新增 `identity` 的教师目录聚合 contract
- 不改 `teaching_query` 现有学生目录、班级总览、复盘聚合 SQL
- 不调整 `assessment` 推荐能力注入方式

## 3. 实施步骤

1. 调整 composition 和 runtime 依赖，把 `identity.Users` 传入 `teaching_query`
2. 修改 `service.go`、`overview_service.go`、`class_insight_service.go`、`student_review_service.go`，把用户 lookup 从本地 repo 拆出来
3. 删除 `teaching_query` 本地 repo 中冗余的 `FindUserByID` 实现，并收口相关 ports / tests
4. 更新 `01-system-architecture.md`、`07-modular-monolith-refactor.md` 中 `teaching_query` 的依赖说明

## 4. 验证

- `go test ./internal/module/teaching_query/... -count=1`
- `go test ./internal/app -run 'TestTeachingQueryModuleContractsCompile|TestCompositionModulesExposeContracts|TestTeachingQueryModuleUsesTypedDeps|TestTeacherRoutesAreServedByTeachingQuery' -count=1`
- `python3 scripts/check-docs-consistency.py`

## 5. 风险

- 如果直接复用 `identity` 的 not-found 语义，`teaching_query` 现有 unauthorized / forbidden 分支可能被误改成 internal error；需要显式兼容 `identitycontracts.ErrUserNotFound`
- 如果误把学生目录查询也挪到 `identity`，会破坏当前聚合排序和分页语义
