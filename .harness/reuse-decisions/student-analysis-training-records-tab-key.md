# Reuse Decision

## Change type
- page
- component
- model
- test

## Existing code searched
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/studentAnalysisWorkspaceTabs.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/studentInsightShared.ts`
- `code/frontend/src/shared/model/navigation/useRouteQueryTabs.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `code/frontend/src/pages/__tests__/studentAnalysisRouteTestSupport.ts`

## Similar implementations found
- 学员分析页已经把 tab 的展示文案与内部 key 分开，当前只是 `timeline -> 训练记录` 这组命名没有继续收口。
- `useRouteQueryTabs.ts` 已经负责 `panel` query 的读取与回写，因此兼容旧 key 的最佳位置是在学员分析页自己的 route-tab owner，而不是改共享路由工具。

## Decision
refactor_existing

## Reason
- 这次不是新增一个 panel，也不是重做 training timeline 实体，而是把“训练记录”这个产品语义从 tab key / panel id / query 参数上收口出来。
- 最小正确改法是只改学员分析页自己的 tab key 和 query 兼容逻辑，保留 `TrainingTimelineContent`、`TimelineEvent`、`getStudentTimeline` 等训练时间线实体命名不动，避免把页面命名调整扩大成跨域重命名。
- 旧 `panel=timeline` 需要继续兼容，否则历史链接会直接失效。

## Files to modify
- `.harness/reuse-decisions/student-analysis-training-records-tab-key.md`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/studentAnalysisWorkspaceTabs.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/studentInsightShared.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`

## After implementation
- 以后遇到“页面文案已经改成业务语义，但 URL / panel key 仍泄漏旧展示抽象”的情况，优先在页面 owner 层做 key 收口与兼容迁移，不直接扩散到共享 entity 命名。
