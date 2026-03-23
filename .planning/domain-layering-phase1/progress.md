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
- 尚未开始模块内部物理搬迁；`challenge / contest / assessment / practice` 仍保持根目录大平铺，下一轮继续按 owner-first 分层切片
