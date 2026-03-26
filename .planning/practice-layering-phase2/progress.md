# Progress

- 2026-03-26: 启动 `practice-layering-phase2`，首刀聚焦删除宽 `PracticeRepository`，暂不扩散到 `runtime` adapter 清理。
- 2026-03-26: 已删除 `practice/ports` 中 legacy 宽 `PracticeRepository`，改为 `PracticeCommandRepository / PracticeCommandTxRepository / PracticeScoreRepository / PracticeRankingRepository`。
- 2026-03-26: `practice` application 构造依赖已切换到窄端口，命令服务、分数写侧、排行榜读侧不再共用单个宽仓储接口。
- 2026-03-26: [`practice_module.go`](/home/azhi/workspace/projects/ctf/code/backend/internal/app/composition/practice_module.go) 已切到 typed deps，避免继续直接把 concrete `practiceinfra.Repository` 塞给多个服务。
- 2026-03-26: 已验证
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/practice/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestCompositionBuildersUseRuntimeModuleForRuntimeDependencies|TestPracticeModuleUsesTypedPortsDeps|TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes' -count=1`
