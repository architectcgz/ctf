# Task Plan

## Goal

启动 `challenge / contest / assessment / practice` 的内部物理分层 Phase 1，先去掉 composition 对 concrete repository/service 的直接暴露，再按职责切分大平铺模块。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 composition 当前暴露的 concrete 能力 | completed | 已确认 `challenge / contest / assessment / practice` 的泄漏点与依赖面 |
| 2. `challenge` Phase 1 | completed | 已完成 `api/http`、`application`、`infrastructure` 物理分层；app/runtime/受影响测试已切到新目录，根包旧实现已物理删除 |
| 3. `contest` Phase 1 | in_progress | `ContestModule` 已移除对外 `Repository` 暴露，内部 `core / participation / team / submission / awd` 仍待拆分 |
| 4. `assessment` Phase 1 | completed | 已完成 `api/http`、`application`、`infrastructure` 物理分层并切到新 handler 装配 |
| 5. `practice` Phase 1 | completed | 已完成 `api/http`、`application`、`infrastructure`、`testsupport` 物理分层；`PracticeModule`/测试已切到新目录，根包旧实现与重复测试已物理删除 |
| 6. 收紧 composition 暴露与架构测试 | completed | composition 已改为只暴露 handler + contracts/生命周期接口 |
| 7. 定向验证 | completed | contract 收口与 `assessment` 物理分层的 focused tests 已通过 |

## Key Files

- `code/backend/internal/app/composition/challenge_module.go`
- `code/backend/internal/app/composition/contest_module.go`
- `code/backend/internal/app/composition/assessment_module.go`
- `code/backend/internal/module/challenge/api/http/*`
- `code/backend/internal/module/challenge/application/*`
- `code/backend/internal/module/challenge/infrastructure/*`
- `code/backend/internal/module/contest/*`
- `code/backend/internal/module/assessment/*`
- `code/backend/internal/module/practice/api/http/*`
- `code/backend/internal/module/practice/application/*`
- `code/backend/internal/module/practice/infrastructure/*`
- `code/backend/internal/module/practice/testsupport/*`

## Acceptance Checks

- `ChallengeModule` 不再暴露 `ImageRepository / ImageService / Repository` 这类 concrete 成员
- `challenge` 根包不再保留 `New*Handler / New*Service / New*Repository` 兼容入口
- `ContestModule` 不再暴露宽泛的 `Repository`
- `AssessmentModule` 对外只暴露必要 contract，而不是继续泄漏 concrete service
- `practice` 保持事件驱动与写流程 owner，不重新吸收 read/query 逻辑
- `practice` 根包不再保留 `New*Handler / New*Service / New*Repository` 兼容入口
- 定向测试通过：
  - `GOMAXPROCS=2 go test -p 1 -parallel 1 ./internal/module/challenge/... ./internal/module/contest ./internal/module/practice/... -count=1`
  - `GOMAXPROCS=2 go test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts|TestNewRouterRegistersStudentChallengeRoutes|TestRouterBuildUsesCompositionModules|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge|TestArchitectureRulesRejectConcreteCrossModuleImports' -count=1`

## Constraints

- 按子能力小步迁移，每个子切片可独立提交
- 已完成子能力物理迁移的模块不保留兼容层
- 不改外部 API，不碰 frontend
