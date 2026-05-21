# Reuse Decision

## Change type

service / view / component / test

## Existing code searched

- `code/backend/internal/module/assessment/application/queries/*.go`
- `code/backend/internal/module/assessment/application/queries/*test.go`
- `code/backend/internal/module/assessment/ports/*.go`
- `code/backend/internal/module/assessment/infrastructure/repository.go`
- `code/frontend/src/views/profile/*.vue`
- `code/frontend/src/components/teacher/*.vue`
- `code/frontend/src/utils/skillProfile.ts`

## Similar implementations found

- `assessment/application/queries/recommendation_service.go` 已经是学生推荐逻辑的唯一 owner，适合继续收口六维弱项判断
- `assessment/application/commands/profile_service.go` 已经稳定产出六维 `SkillProfile` 分数，不需要另起新的画像或测评服务
- `frontend/src/views/profile/SkillProfile.vue` 和 `frontend/src/components/teacher/StudentInsightPanel.vue` 已经承载学生端与教师端六维图展示，不需要新增页面或并行组件

## Decision

refactor_existing

## Reason

这次目标不是增加新的画像模块或新的推荐链路，而是把已有推荐逻辑从“教学复盘推断 + 复杂包装”收回到“六维分数驱动”的更直接口径：

- 后端继续复用 `RecommendationService`、`SkillProfile` 和 challenge recommendation repository
- 前端继续复用现有学生画像页和教师洞察面板
- 测试继续放在既有 assessment query tests 与 skill profile view tests 中

这样可以用最小 diff 修正论文与系统口径不一致的问题，同时避免再引入新的 DTO、service 或页面分支。

## Files to modify

- `.harness/reuse-decisions/assessment-six-dimension-recommendation-simplification.md`
- `docs/plan/impl-plan/2026-05-21-assessment-six-dimension-recommendation-simplification-implementation-plan.md`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
- `code/frontend/src/views/profile/SkillProfile.vue`
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- `code/frontend/src/views/profile/__tests__/SkillProfile.test.ts`

## After implementation

- 六维弱项与推荐目标以 `SkillProfile` 低分维度为准
- 推荐摘要展示真实维度覆盖率，而不是反向置信度
- 学生端与教师端页面统一为“六维学习画像 / 六维训练分布”表述
