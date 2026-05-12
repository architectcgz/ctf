# Reuse Decision

## Change type
component / layout

## Existing code searched
- code/frontend/src/components/teacher/TeacherClassTrendPanel.vue
- code/frontend/src/components/teacher/TeacherClassReviewPanel.vue
- code/frontend/src/components/teacher/TeacherClassInsightsPanel.vue
- code/frontend/src/components/teacher/TeacherInterventionPanel.vue
- code/frontend/src/components/teacher/teacher-panel-shell.css
- code/frontend/src/components/teacher/teacher-workspace-subpanel.css
- code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts
- code/frontend/src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts

## Similar implementations found
- `teacher-panel-shell.css` 已经统一管理标题、subtitle、panel header 的共享样式，适合在这里兼容“没有 eyebrow 时标题首行贴顶”的场景。
- `teacher-workspace-subpanel.css` 是班级工作区 panel 的共享 wrapper 样式 owner，局部 eyebrow 深选择器应该随结构移除一并清掉。
- `teacherEyebrowSharedStyles.test.ts` 和 `teacherWorkspaceSubpanelAdoption.test.ts` 已经覆盖 eyebrow / subpanel 的共享约束，适合在现有测试里补结构回归。

## Decision
refactor_existing

## Reason
这次不是做新的标题模式，而是把班级工作区四个现有 panel 的 eyebrow 结构去掉。沿用共享 panel shell 处理标题间距，比在每个 panel 内各自补 margin 修正更稳，也能顺手清掉已经无用的 subpanel eyebrow 覆盖规则。

## Files to modify
- .harness/reuse-decisions/class-panel-eyebrow-removal.md
- code/frontend/src/components/teacher/TeacherClassTrendPanel.vue
- code/frontend/src/components/teacher/TeacherClassReviewPanel.vue
- code/frontend/src/components/teacher/TeacherClassInsightsPanel.vue
- code/frontend/src/components/teacher/TeacherInterventionPanel.vue
- code/frontend/src/components/teacher/teacher-panel-shell.css
- code/frontend/src/components/teacher/teacher-workspace-subpanel.css
- code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts
- code/frontend/src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts
