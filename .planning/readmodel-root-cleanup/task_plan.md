# Task Plan

## Goal

删除 `practice_readmodel` 与 `teaching_readmodel` 的根包兼容壳，让 readmodel 模块只通过 `api / application / infrastructure / contracts` 暴露能力。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点两个 readmodel 的根壳职责 | completed | 已确认 `practice` 是纯转发壳，`teaching` 是不完整兼容壳 |
| 2. 为两个 readmodel 固化 narrow contracts | completed | `PracticeQuery` / `TeachingQuery` 已成为稳定对外面 |
| 3. 更新 composition 装配 | completed | 已直接注入 query service / contract |
| 4. 删除 `module.go` 根壳与多余转发 | completed | 两个 `module.go` 已删除 |
| 5. 定向验证 | completed | readmodel + app focused tests 已通过 |

## Key Files

- `code/backend/internal/app/composition/practice_readmodel_module.go`
- `code/backend/internal/app/composition/teaching_readmodel_module.go`
- `code/backend/internal/module/practice_readmodel/contracts.go`
- `code/backend/internal/module/practice_readmodel/module.go`
- `code/backend/internal/module/practice_readmodel/api/http/handler.go`
- `code/backend/internal/module/teaching_readmodel/contracts.go`
- `code/backend/internal/module/teaching_readmodel/module.go`
- `code/backend/internal/module/teaching_readmodel/api/http/handler.go`

## Acceptance Checks

- `practice_readmodel/module.go` 删除
- `teaching_readmodel/module.go` 删除
- composition 不再 import 这两个模块的根包壳
- 定向测试通过：
  - `GOMAXPROCS=2 go test -p 1 -parallel 1 ./internal/module/practice_readmodel/... ./internal/module/teaching_readmodel/... -count=1`
  - `GOMAXPROCS=2 go test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts|TestNewRouter|TestTeacherRoutesAreServedByTeachingReadModel' -count=1`

## Constraints

- 保持现有教师/学生 read API 路径不变
- 不把读模型重新塞回 `practice` 或其他写模块
- 优先沿用 `practice-readmodel-phase2` 已验证的模式
