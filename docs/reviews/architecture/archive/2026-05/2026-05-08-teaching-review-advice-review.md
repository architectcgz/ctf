# 2026-05-08 教学复盘建议生成架构 Review

## Scope

- `docs/design/教学复盘建议优化方案.md`
- `docs/architecture/features/教学复盘建议生成架构.md`
- `docs/plan/impl-plan/2026-05-08-teaching-review-advice-implementation-plan.md`

## Verdict

- `Gate verdict: pass with residual risks`

## Checked Points

- `internal/teaching/advice` 已明确成为弱项维度识别、排序、证据充足性判断和建议规则的单一 owner。
- `RecommendationService` 已降回候选题查询、已解过滤和结果装配角色，不再保留独立弱项判定 owner。
- 输入模型已区分“原始教学事实快照”和 `advice` 派生结果，`confidence / is_weak / severity` 不再混入输入层。
- `severity` 已被提升为统一主语义字段，旧的 `Accent`、`Level` 展示型字段不再作为目标契约保留。
- 整体范围仍贴着毕业设计主线，没有扩成 LLM 教学助手、学习路径平台或新的前端专题工程。

## Findings

无 blocker。

## Residual Risks

- 当前性能约束仍停留在架构与计划层，后续实现 review 需要实际检查是否出现按学生 fan-out 读取事实的回退。
- 当前负向校验已进入实现计划，但仍需在实现阶段补齐测试，确认“证据不足维度”不会在 class review、archive、recommendation 任一链路中被误判成明确弱项。
