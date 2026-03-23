# Progress

## 2026-03-23

- 创建大业务模块内部分层 Phase 1 计划
- 已确认优先目标不是“大规模重命名”，而是先去掉 composition 对 concrete 依赖的继续扩散
- 完成首个收口子切片：
  - `ChallengeModule` 不再暴露 `ImageRepository / ImageService / Repository / FlagService` concrete 字段，改为 `Catalog / FlagValidator / ImageStore` contract 与生命周期接口
  - `ContestModule` 不再对外暴露 `Repository`
  - `AssessmentModule` 新增 `assessment/contracts`，对外改为暴露 `ProfileService / Recommendations` contract 与生命周期接口
  - `PracticeModule` 不再对外暴露 concrete `Service`
  - `TeachingReadmodelModule` 已改为依赖 `AssessmentModule.Recommendations`
- 限核定向验证通过：
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/challenge ./internal/module/contest ./internal/module/assessment ./internal/module/practice ./internal/module/teaching_readmodel/... -count=1`
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestCompositionBuildersUseRuntimeModuleForRuntimeDependencies|TestNewRouterRegistersStudentChallengeRoutes|TestRouterBuildUsesCompositionModules|TestArchitectureRulesRejectConcreteCrossModuleImports|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
- 完成 `assessment` 物理分层：
  - HTTP handler 已迁到 `internal/module/assessment/api/http`
  - 能力画像 / 推荐 / 报告 / cleaner 已迁到 `internal/module/assessment/application`
  - repository 已迁到 `internal/module/assessment/infrastructure`
  - `AssessmentModule` 已直接装配新三层目录，不再依赖 `internal/module/assessment` 根包
  - assessment 路由 handler 来源已切到 `internal/module/assessment/api/http`
- 新一轮限核定向验证通过：
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/assessment/... -count=1`
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts|TestNewRouterRegistersStudentChallengeRoutes|TestRouterBuildUsesCompositionModules|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge|TestArchitectureRulesRejectConcreteCrossModuleImports' -count=1`
- 完成 `challenge` 物理分层：
  - HTTP handler 已迁到 `internal/module/challenge/api/http`
  - challenge / flag / image / topology / writeup / tag service 已迁到 `internal/module/challenge/application`
  - challenge / image / writeup / tag repository 已迁到 `internal/module/challenge/infrastructure`
  - 新增 `internal/module/challenge/testsupport` 收敛测试建表与 runtime stub
  - `ChallengeModule` 与 `RuntimeModule` 已切到新目录装配，`practice/contest/app` 受影响测试已改为新依赖
  - `challenge` 根包旧 `handler/service/repository` 与对应测试已物理删除，不保留兼容层
- 本轮限核定向验证通过：
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/challenge/... -count=1`
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest ./internal/module/practice -count=1`
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts|TestNewRouterRegistersStudentChallengeRoutes|TestRouterBuildUsesCompositionModules|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge|TestArchitectureRulesRejectConcreteCrossModuleImports' -count=1`
