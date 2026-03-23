# Task Plan

## Goal

启动 `challenge / contest / assessment / practice` 的内部物理分层 Phase 1，先去掉 composition 对 concrete repository/service 的直接暴露，再按职责切分大平铺模块。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 composition 当前暴露的 concrete 能力 | pending | `ChallengeModule` / `ContestModule` / `AssessmentModule` |
| 2. `challenge` Phase 1 | pending | 按 `catalog / flag / image / topology / writeup` 切片 |
| 3. `contest` Phase 1 | pending | 按 `core / participation / team / submission / awd` 切片 |
| 4. `assessment` Phase 1 | pending | 按 `profile / recommendation / report` 切片 |
| 5. `practice` Phase 1 | pending | 保持写流程 owner，避免读逻辑回流 |
| 6. 收紧 composition 暴露与架构测试 | pending | 只暴露 handler + contracts |
| 7. 定向验证 | pending | 模块级 focused tests |

## Key Files

- `code/backend/internal/app/composition/challenge_module.go`
- `code/backend/internal/app/composition/contest_module.go`
- `code/backend/internal/app/composition/assessment_module.go`
- `code/backend/internal/module/challenge/*`
- `code/backend/internal/module/contest/*`
- `code/backend/internal/module/assessment/*`
- `code/backend/internal/module/practice/*`

## Acceptance Checks

- `ChallengeModule` 不再暴露 `ImageRepository / ImageService / Repository` 这类 concrete 成员
- `ContestModule` 不再暴露宽泛的 `Repository`
- `AssessmentModule` 对外只暴露必要 contract，而不是继续泄漏 concrete service
- `practice` 保持事件驱动与写流程 owner，不重新吸收 read/query 逻辑
- 定向测试通过：
  - `GOMAXPROCS=2 go test -p 1 -parallel 1 ./internal/module/challenge ./internal/module/contest ./internal/module/assessment ./internal/module/practice -count=1`
  - `GOMAXPROCS=2 go test -p 1 -parallel 1 ./internal/app/composition ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestCompositionBuildersUseRuntimeModuleForRuntimeDependencies' -count=1`

## Constraints

- 按子能力小步迁移，每个子切片可独立提交
- 不一次性重命名所有包；优先先去掉 concrete 暴露
- 不改外部 API，不碰 frontend

