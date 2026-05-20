# Practice / Assessment 边界 Phase 3 Slice 1 Implementation Plan

## Objective

完成后端模块边界迁移里 `practice -> assessment` 这条剩余同步耦合的第一刀收口：

- 去掉 `practice` runtime / command service 对 `assessment.ProfileService` 的直接依赖
- 保持能力画像增量更新继续通过现有 `practice.flag_accepted` 事件消费者完成
- 不改变练习提交、人工评审通过、得分刷新和事件 payload 的外部行为

## Non-goals

- 不在这一轮改 `assessment` 的画像计算规则、分布式锁或 Redis key
- 不新增新的 practice 事件类型，也不扩展 `practicecontracts.FlagAcceptedEvent`
- 不处理 `practice_readmodel` 归并、`contest -> auth`、`runtime` 物理拆包或 application concrete allowlist 的其他条目
- 不补独立消息队列；仍然沿用现有进程内 `events.Bus`

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/app/composition/practice_module.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/practice/application/commands/{service.go,service_lifecycle.go,submission_service.go,manual_review_service.go}`
- `code/backend/internal/module/assessment/runtime/module.go`
- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/architecture_allowlist_test.go`

## Current Baseline

- `practice` 组合层仍显式注入 `assessment.ProfileService`，并在 `practice/runtime/module.go` 保留 `Assessment assessmentcontracts.ProfileService` 依赖。
- `practice/application/commands/service_lifecycle.go` 里仍有 `triggerAssessmentUpdate` 异步直调路径。
- 同时，`practice` 在正确提交和人工评审通过时已经发布 `practice.flag_accepted`，而 `assessment` runtime 也已经通过 `RegisterPracticeEventConsumers` 订阅这条事件。
- 当前形成了“双轨”实现：事件消费者和直调画像更新同时存在，但只有事件链路是目标边界的一部分。

## Chosen Direction

这次不引入新的 bridge，不保留“事件失败时再直调画像”的过渡兜底，而是直接把 `practice -> assessment` 依赖从 wiring 和 service 中删除：

1. `practice` 只负责发布 `practice.flag_accepted` 与异步得分更新
2. `assessment` 继续作为事件消费者处理画像增量更新和推荐缓存刷新
3. `app/composition` 与 `practice/runtime` 不再接收或传递 `assessment.ProfileService`
4. 删除 `practice/application/commands` 内与 assessment 直调相关的结构、方法和测试
5. 收紧 `practice` 架构守卫与模块依赖 allowlist，防止 `practice -> assessment` 回流

## Ownership Boundary

- `practice`
  - 负责：提交判定、人工评审通过后的事件发布、得分刷新调度
  - 不负责：直接调用能力画像写服务或自行定义画像更新时序
- `assessment`
  - 负责：消费 `practice.flag_accepted`，执行画像增量更新与推荐缓存失效
  - 不负责：由 `practice` 同步驱动写入，或要求 `practice` 持有 profile service 句柄
- `composition`
  - 负责：只装配 `practice` 真实需要的 challenge / instance / runtime 依赖
  - 不负责：继续为 `practice` 传递已被事件化的 `assessment` 能力

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-12-practice-assessment-boundary-phase3-slice1-implementation-plan.md`
- Add: `.harness/reuse-decisions/practice-assessment-boundary-phase3-slice1.md`
- Modify: `code/backend/internal/app/composition/practice_module.go`
- Modify: `code/backend/internal/module/practice/runtime/module.go`
- Modify: `code/backend/internal/module/practice/application/commands/service.go`
- Modify: `code/backend/internal/module/practice/application/commands/service_lifecycle.go`
- Modify: `code/backend/internal/module/practice/application/commands/service_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/service_lifecycle_test.go`
- Modify: `code/backend/internal/module/practice/architecture_test.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/architecture/backend/{01-system-architecture.md,07-modular-monolith-refactor.md}`
- Modify: `docs/design/backend-module-boundary-target.md`

## Task Slices

### Slice 1: 删除 practice 的 assessment wiring

目标：

- `BuildPracticeModule` 不再接收 `assessment`
- `practice/runtime.Module` 不再声明 `Assessment assessmentcontracts.ProfileService`
- `practice` command service 构造函数不再注入 assessment service

Validation:

- `cd code/backend && rg -n "Assessment\\s+assessmentcontracts\\.ProfileService|assessmentService|triggerAssessmentUpdate" internal/module/practice internal/app/composition`
- `cd code/backend && go test ./internal/app/... ./internal/module/practice/... -run 'Practice|Router|Composition' -count=1`

Review focus:

- `practice` 是否还保留隐藏的 assessment 直调入口
- composition 是否已经只保留真实 owner 依赖

### Slice 2: 收紧 guardrail 与 allowlist

目标：

- 删除 `allowedModuleDependencies` 里的 `practice -> assessment`
- 更新 `practice/architecture_test.go` 的 typed deps 断言
- 让模块边界测试阻止旧依赖回流

Validation:

- `cd code/backend && go test ./internal/module/... -run 'Architecture|Allowlist' -count=1`

Review focus:

- allowlist 是否只删掉真实已消失的依赖，没有残留 stale entry
- practice runtime 守卫是否表达成“没有 assessment 依赖”，而不是仅仅少改一个 import

### Slice 3: 文档对齐当前事实

目标：

- 当前架构事实不再把 `practice -> assessment` 描述成装配依赖
- 目标设计稿的 phase 3 进度补到当前切片

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

Review focus:

- 文档是否明确区分“已落地事实”和“仍未完成的 phase 3 余项”
- 是否避免把未完成的事件化范围误写成“整体 phase 3 已完成”

## Integration Checks

- 练习正确提交后仍然发布 `practice.flag_accepted`
- 人工评审通过后仍然发布 `practice.flag_accepted`
- `assessment` runtime 仍在注册 practice 事件消费者
- `practice` 背景任务关闭逻辑仍能正确回收异步得分更新

## Rollback / Recovery Notes

- 本切片是纯代码与文档改动，无 schema / migration 影响，可代码级回滚
- 如果事件链路验证失败，应整体回退这一刀，而不是只恢复 composition wiring 的一半

## Risks

- 如果某些路径仍隐式依赖 `triggerAssessmentUpdate` 的定时行为，移除后可能暴露“只剩事件链路”的真实时序差异
- 如果 practice 未来新增“画像更新但不发 `FlagAccepted`”的分支，这次 guardrail 不会自动覆盖；需要该分支单独建事件或明确归属
- 这次不会解决 `practice_readmodel` 和 application concrete allowlist 的剩余迁移债，它们仍是后续 phase 4/5 范围

## Verification Plan

1. `cd code/backend && go test ./internal/module/practice/... ./internal/module/assessment/... ./internal/app/... -run 'Practice|Assessment|Router|Composition' -count=1`
2. `cd code/backend && go test ./internal/module/... -run 'Architecture|Allowlist' -count=1`
3. `python3 scripts/check-docs-consistency.py`
4. `bash scripts/check-consistency.sh`
5. `bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- 目标边界明确：`practice` 只发事件，`assessment` 只消费事件，不再有同步画像 service 注入
- reuse point 明确：复用现有 `practice.flag_accepted` 与 `assessment.RegisterPracticeEventConsumers`，不新增并行桥接抽象
- 这刀同时解决行为与结构：不是仅仅“事件也能工作”，而是把旧的同步 owner 连接从 wiring、service、测试和 allowlist 一起删掉
- 本切片不会制造“合并后立刻还要再做第二轮 practice/assessment 解耦”的重复重构；剩余未完成项已经转入 phase 3 其他范围和 phase 4/5
