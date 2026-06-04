# Reuse Decision

## Change type
- page
- component
- layout
- style
- test

## Existing code searched
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-shared/ui/StudentInsightStateSurface.vue`
- `code/frontend/src/features/teaching/student-analysis-shared/ui/studentInsightSurface.css`
- `code/frontend/src/assets/styles/workspace-shell.css`
- `code/frontend/src/assets/styles/surface-shell-background.css`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`

## Similar implementations found
- `StudentInsightStateSurface` 已经有 `surface="glass" | "plain"` 的状态壳概念，但目前它和 loading surface 更偏“组件状态壳”，还不是独立的局部 glass 区域体系。
- `surface-shell-background.css` 当前把 `.workspace-shell` 统一做成带 radial gradient 的 shell，因此学员详情页一旦继承 `workspace-shell`，整页都会带玻璃感。
- 推荐列表自己私有实现了一套 glass 渐变/高光/阴影，与 `StudentInsightStateSurface` 的 glass surface 语义重复。

## Decision
refactor_existing

## Reason
- 这次不是新增视觉风格，而是把“页壳 surface”和“局部 glass region”两个 owner 拆开。
- 最小正确改法是：
  - 给 `workspace-shell` 增加显式 plain modifier，让页面可以关闭整页 glass 背景。
  - 给 student insight 共享 surface 抽出独立 `glass region` 类，供列表/区域显式挂载。
  - 当前页先接入这套抽象，避免继续在推荐列表里写一份私有 glass CSS。
- 不改 `metric-panel` 全局体系，不把这次范围扩大成所有 dashboard / timeline / directory 的统一重构。

## Files to modify
- `.harness/reuse-decisions/student-analysis-glass-region-separation.md`
- `code/frontend/src/assets/styles/workspace-shell.css`
- `code/frontend/src/features/teaching/student-analysis-shared/ui/studentInsightSurface.css`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue`
- `code/frontend/src/entities/training-timeline/ui/TrainingTimelineContent.vue`
- `code/frontend/src/assets/styles/workspace-glass.css`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts`
- `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`

## After implementation
- 以后如果某个教师页只想让局部列表/卡片带 glass，不再通过整页 `workspace-shell` 背景兜底；优先使用 `workspace-shell--plain` + 显式 glass region。
