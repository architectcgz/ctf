# Reuse Decision

## Change type
component / composition

## Existing code searched
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue`
- `code/frontend/src/components/platform/class/ClassManageHeroPanel.vue`
- `code/frontend/src/components/platform/student/StudentManageHeroPanel.vue`
- `code/frontend/src/components/teacher/review-archive/ReviewArchiveHero.vue`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/platform/class/ClassManageHeroPanel.vue`
- `code/frontend/src/components/platform/student/StudentManageHeroPanel.vue`
- `code/frontend/src/components/teacher/review-archive/ReviewArchiveHero.vue`

## Decision
refactor_existing

## Reason
- `StudentAnalysisPage.vue` 现在已经把 tab 选择和单实例 `StudentInsightPanel` 挂载收回到明确 owner，剩下最稳定的大块是 overview tab 的 hero 壳层：标题、快捷动作、summary 指标卡和分隔线。
- 仓库里已有多处 `*HeroPanel.vue` 作为稳定展示壳的收口模式，说明当前最小改动不是再造 page model 或 composable，而是继续沿用“父页 owner + 子组件承接稳定展示区”的拆分方式。
- `StudentAnalysisPage.vue` 仍负责 tab / query 同步、页面级数据装配和事件桥接；新组件只接收现成 props 并通过 emits 触发页面动作，不应直接拥有路由、请求或工作流状态。

## Files to modify
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisOverviewHeroPanel.vue`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`

## After implementation
- 如果教师端 class-management 下的 hero 抽层模式继续稳定复用，再考虑把它补进 `.harness/reuse-index/`。
