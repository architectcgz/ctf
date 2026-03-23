# Progress

## 2026-03-23

- 创建 readmodel 根壳清理计划
- 已记录两个 readmodel 当前对外暴露方式不一致的问题
- 已删除 `practice_readmodel/module.go`，`PracticeReadmodelModule` 直接注入 `*application.QueryService`
- 已删除 `teaching_readmodel/module.go`，`TeachingQuery` 已扩成 handler / composition 实际需要的窄接口集合
- `teaching_readmodel/api/http` 已改为依赖 `TeachingQuery` 接口，不再依赖具体 `QueryService`
- `internal/app/router_test.go` 与 `internal/app/practice_flow_integration_test.go` 已同步清理旧根壳引用
- 已通过定向验证：
  - `GOMAXPROCS=2 go test -p 1 -parallel 1 ./internal/module/practice_readmodel/... ./internal/module/teaching_readmodel/... -count=1`
  - `GOMAXPROCS=2 go test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts|TestNewRouter|TestTeacherRoutesAreServedByTeachingReadModel|TestTeachingReadmodelModuleContractsCompile|TestPracticeReadmodelModuleContractsCompile|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
