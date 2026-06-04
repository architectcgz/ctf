# Reuse Decision

## Change type
component styling / component template

## Existing code searched
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.test.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/studentReviewWorkspacePresentation.test.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/studentInsightShared.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`

## Similar implementations found
- `StudentInsightAttackSessionsSection.vue` 已经把证据链工作区包在 `SectionCard variant="teacher-flat"` 里，说明内容区可以继续走扁平化结构，不需要额外摘要条再叠一层信息。
- `StudentReviewWorkspace.vue` 现有会话列表已经使用分隔线和时间线样式，适合直接在这一层收紧垂直节奏，而不是新增另一套列表外壳。

## Decision
extend_existing

## Reason
- 用户要求移除证据链中的一块摘要条，并压缩会话列表高度。
- 最小正确改动是直接在 `StudentReviewWorkspace.vue` 删除观察摘要条渲染，并收紧现有 `attack-session` / `attack-event` 的间距、字号和时间线节奏，不改数据流、筛选逻辑或 SectionCard owner。

## Files to modify
- `.harness/reuse-decisions/student-review-evidence-panel-tighten.md`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.test.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.test.ts`

## After implementation
- 证据链面板不再显示额外的观察摘要条。
- 摘要卡不再落在 list 容器内部，而是放到 list 上方。
- 攻击会话列表维持现有信息结构，但首屏占高更低，列表更紧凑。
