# Reuse Decision

## Change type
component / page

## Existing code searched
- code/frontend/src/components/teacher/TeacherClassReviewPanel.vue
- code/frontend/src/components/teacher/TeacherClassInsightsPanel.vue
- code/frontend/src/components/teacher/TeacherInterventionPanel.vue
- code/frontend/src/components/teacher/teacher-panel-shell.css
- code/frontend/src/assets/styles/teacher-surface.css
- code/frontend/src/assets/styles/journal-notes.css
- code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts

## Similar implementations found
- `teacher-panel-shell.css` 已经为教师 detail panel 定义 `--panel-border` 和 `--panel-divider`，用于 panel 外壳、空态和图表容器边框。
- `journal-notes.css` 里的 `showcase-panel-card` / `showcase-panel-card--minimal-wire` 已经把拆分卡片边框统一收口到 `--showcase-panel-border`，默认也是共享边框 token。
- `teacher-surface.css` 已经提供 `--teacher-card-border` 作为教师工作区卡片边框 token，说明当前页面不需要继续在局部组件里手工混入 accent 颜色生成卡片边框。

## Decision
refactor_existing

## Reason
这次不是新增卡片样式，而是把班级工作区 review / insight / action 三个 panel 中漂移出去的局部边框公式收回到现有共享 token。沿用 `teacher-panel-shell.css` 和 `showcase-panel-card` 的边框 token，可以保持教师工作区卡片边框一致，同时把强调色限制在内容层和强调线，不再让每个卡片自己定义一套边框颜色。

## Files to modify
- .harness/reuse-decisions/class-workspace-card-border-token.md
- code/frontend/src/components/teacher/TeacherClassReviewPanel.vue
- code/frontend/src/components/teacher/TeacherClassInsightsPanel.vue
- code/frontend/src/components/teacher/TeacherInterventionPanel.vue
- code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts
