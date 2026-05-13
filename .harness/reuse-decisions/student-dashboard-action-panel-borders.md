# Reuse Decision

## Change type
component / page

## Existing code searched
- code/frontend/src/components/dashboard/student/StudentCategoryProgressPage.vue
- code/frontend/src/components/dashboard/student/StudentDifficultyPage.vue
- code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue
- code/frontend/src/assets/styles/journal-soft-surfaces.css
- code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts

## Similar implementations found
- `StudentDifficultyPage.vue` 已经有稳定的学生侧行动区边框语言：外层列表壳、主推荐项强调态、说明块外壳。
- `StudentCategoryProgressPage.vue` 和 `StudentRecommendationPage.vue` 已经在复用同一类外层列表边框公式，但当前只停留在各页面本地重复，没有收口成共享类。
- `journal-soft-surfaces.css` 已经集中维护学生侧按钮、空态、弱强调胶囊和 shell token，适合继续承接这类共享边框壳，而不是让每个 dashboard panel 再手工拼一套局部边框。

## Decision
refactor_existing

## Reason
这次不是新增一种新的视觉样式，而是把学生 dashboard `category` / `difficulty` 两个 panel 里重复的边框外壳和 item 边框收回到现有 `journal-soft-surfaces.css`。这样可以直接复用既有学生侧 token，避免 `category` 为了补 item 边框再复制一份 `difficulty` 的本地 border/background 公式。

## Files to modify
- .harness/reuse-decisions/student-dashboard-action-panel-borders.md
- code/frontend/src/assets/styles/journal-soft-surfaces.css
- code/frontend/src/components/dashboard/student/StudentCategoryProgressPage.vue
- code/frontend/src/components/dashboard/student/StudentDifficultyPage.vue
- code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts
