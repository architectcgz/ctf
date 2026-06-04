# Reuse Decision

## Change type
frontend architecture / component owner / layout

## Existing code searched
- code/frontend/src/entities/training-timeline/ui/TrainingTimelinePanel.vue
- code/frontend/src/features/student-dashboard/ui/StudentRecommendationPage.vue
- code/frontend/src/features/student-dashboard/ui/StudentCategoryProgressPage.vue
- code/frontend/src/features/student-dashboard/ui/StudentDifficultyPage.vue
- code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupEditorPage.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightTimelineSection.vue
- code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue
- code/frontend/src/style.css

## Similar implementations found
- `TrainingTimelinePanel.vue` 刚完成过一次 owner 拆分：共享内容 owner + 调用方各自页面壳 / section 壳。
- 学生 dashboard 其余 panel 仍存在 `embedded` 入口，需要判断它们是不是同类 mixed owner 问题。
- `ChallengeWriteupEditorPage.vue` 也有 `embedded`，但其使用场景和壳层职责需要单独确认，不应先验地一刀切。

## Decision
extend_existing

## Reason
这次任务不是引入新模式，而是把“不要用 `embedded` 切两套布局语义”固化为项目规则，并把同类现存点按时间线的收口方式继续清理。最小正确路径是：
1. 先扫描所有 `embedded` 使用点
2. 只对 mixed owner 场景做结构收口
3. 把规则沉淀到项目 harness / AGENTS / feedback

## Files to modify
- .harness/reuse-decisions/embedded-layout-owner-retirement.md
- ctf/AGENTS.md
- feedback/*.md
- code/frontend/src/**（仅命中 mixed owner 场景的文件）
- code/frontend/src/**/__tests__/*.test.ts（相关护栏）
- docs/plan/impl-plan/*.md

## After implementation
- 项目会明确禁止新增“`embedded` 同时切页面壳与内容 owner”的实现方式。
- 现存同类点会尽量改成“共享内容组件 + 调用方各自壳层”。
- 仅承担局部视觉嵌入态且不改变 owner 语义的 `embedded`，可保留但需有明确边界。
