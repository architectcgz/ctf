# Reuse Decision

## Change type

repository / readmodel / service / tests / docs

## Existing code searched

- `code/backend/internal/module/assessment/infrastructure/repository.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/assessment/application/commands/profile_service_test.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service_test.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository.go`
- `code/backend/internal/model/awd.go`
- `docs/design/AWD能力画像回流方案.md`
- `docs/architecture/features/教学复盘建议生成架构.md`
- `docs/reviews/architecture/2026-05-14-teaching-review-thesis-gap-review.md`

## Similar implementations found

- `code/backend/internal/module/assessment/infrastructure/repository.go`
  - 已经负责 recommendation snapshot 的训练事实装配、`AWDSuccessCount` 统计和 difficulty band 所需的维度事实。
- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
  - 已经负责班级 teaching fact snapshot 的装配，并且已经把 AWD 活跃度 / success count 接入到班级快照。
- `code/backend/internal/module/assessment/infrastructure/report_repository.go`
  - 已经在复盘证据链里区分 `submitted_by_user_id`、`source=submission` 和 `score_gained`，可作为 AWD 个人正向证据口径参考。
- `code/backend/internal/module/assessment/application/commands/profile_service.go`
  - 已经消费 `contest.awd.attack_accepted` 事件，但当前持久化画像公式仍是普通练习 owner。

## Decision

refactor_existing

## Reason

这次不是新增一套 AWD 推荐模块，也不是重写 `skill_profiles` 持久化公式，而是把“个人 AWD 正向成功证据”补回现有 teaching fact snapshot owner，让 recommendation 和 class review 真正吃到竞赛维度事实。

最小正确方案是：

- 继续复用 `assessment/infrastructure.Repository` 作为个人 recommendation snapshot owner
- 继续复用 `teaching_query/infrastructure.Repository` 作为班级 teaching snapshot owner
- 复用现有 `submitted_by_user_id + source=submission + score_gained>0` 作为 AWD 个人正向成功证据口径
- 在 snapshot 装配阶段补齐：
  - 维度内 AWD 成功覆盖
  - AWD difficulty 覆盖
  - 基于 AWD challenge 发布覆盖率的 profile score 补充信号
- 不在这一刀改 `skill_profiles` 持久化表和 `AssessmentDimensionScoreRepository`，避免把普通题 points 口径与 AWD 独立题表强行混成一套伪统一分值
- 不在这一刀处理班级时间段 / 导出结构 owner，那是 thesis gap 的下一条独立切片

这样可以在不发明第二套 advice owner 的前提下，让教学推荐和班级复盘的维度事实真正吸收 AWD 个人成功证据，同时避免跨表分值语义被这刀打乱。

## Files to modify

- `.harness/reuse-decisions/awd-teaching-fact-backfill.md`
- `docs/plan/impl-plan/2026-05-14-awd-teaching-fact-backfill-implementation-plan.md`
- `code/backend/internal/module/assessment/infrastructure/repository.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service_test.go`
- `docs/architecture/features/教学复盘建议生成架构.md`
- `docs/design/AWD能力画像回流方案.md`

## After implementation

- recommendation / class review 的 teaching fact snapshot 会显式吸收 AWD 个人成功证据
- `progression-ready` 与弱项 / 覆盖判断会看到 AWD 维度成功覆盖，而不是只看到 `AWDSuccessCount`
- `skill_profiles` 持久化公式继续保持普通练习 owner，不在本次切片内改写
- 班级时间段 / 导出结构 owner 继续留在 thesis gap 的下一条独立切片
