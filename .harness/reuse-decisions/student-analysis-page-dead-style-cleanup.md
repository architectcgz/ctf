# Reuse Decision

## Change type
component / layout

## Existing code searched
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisOverviewHeroPanel.vue`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisOverviewHeroPanel.vue`

## Decision
refactor_existing

## Reason
- `StudentAnalysisPage.vue` 经过前两刀收口后，模板已经不再挂载 class switch、student directory、context rail 等旧区块，但 scoped style 仍残留对应定义。
- 这些样式是页面内部私有样式，不存在跨文件复用价值；最小正确做法是继续收口当前组件本体，把死样式直接清掉，而不是再抽公共样式层。
- 本轮不改路由、数据 owner 或交互，只清理当前页面内已经失效的局部样式定义。

## Files to modify
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## After implementation
- 如果后续继续清理 `StudentAnalysisPage.vue` 的未使用 props，再单开一刀处理对上层 route view 的传参收口。
