# Reuse Decision

## Change type

component / hook

## Existing code searched

- code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue
- code/frontend/src/components/teacher/StudentInsightPanel.vue
- code/frontend/src/components/teacher/TeacherClassReviewPanel.vue
- code/frontend/src/components/teacher/TeacherInterventionPanel.vue
- code/frontend/src/features/teacher-student-analysis/model/useTeacherInterventionRecommendations.ts
- code/frontend/src/components/common/AppEmpty.vue
- code/frontend/src/views/teacher/**tests**/TeacherClassStudents.test.ts

## Similar implementations found

- `StudentRecommendationPage.vue` 已经定义了学生侧“无推荐题”提示和唯一主 CTA，说明空推荐在产品上是正常状态，不需要伪造题目占位。
- `StudentInsightPanel.vue` 已经在教师学员分析页使用 `AppEmpty` 展示“暂无推荐题目”，说明教师侧详情页也接受明确空状态而不是静默隐藏。
- `TeacherInterventionPanel.vue` 和 `TeacherClassReviewPanel.vue` 已经各自有推荐题展示区块，适合直接在现有区块内补 empty / error fallback，而不是新增独立卡片或第二套 panel。
- `useTeacherInterventionRecommendations.ts` 已经是教师工作台“加载单个学生推荐题”的唯一异步 owner，推荐加载失败与空结果的区分应继续收口在这里。

## Decision

extend_existing

## Reason

这次不是新增推荐模块，而是把后端已经允许的“正常空推荐”状态，在教师侧两个现有消费入口补齐。沿用学生侧空态文案模式、教师侧现有推荐区块和现有 hook 的状态 owner，可以在最小 diff 下把“空结果”和“加载失败”区分清楚，避免再出现静默留白或错误提示混淆。

## Files to modify

- .harness/reuse-decisions/teacher-no-recommendation-empty-states.md
- code/frontend/src/components/teacher/TeacherClassReviewPanel.vue
- code/frontend/src/components/teacher/TeacherInterventionPanel.vue
- code/frontend/src/features/teacher-student-analysis/model/useTeacherInterventionRecommendations.ts
- code/frontend/src/views/teacher/**tests**/TeacherClassStudents.test.ts
