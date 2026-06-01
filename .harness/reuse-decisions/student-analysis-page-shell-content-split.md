# Reuse Decision

## Change type
frontend refactor / page shell owner cleanup

## Existing code searched
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspacePage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisOverviewHeroPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `StudentAnalysisWorkspacePage.vue` 已经承接共享 page model 与 route-level dialog owner，当前更适合继续把 `StudentAnalysisPage.vue` 收口成真正的 UI shell。
- `StudentInsightPanel.vue` 已经承接学员洞察区块组合；`StudentAnalysisOverviewHeroPanel.vue` 已承接 overview hero 顶栏。

## Decision
refactor_existing

## Reason
当前 `StudentAnalysisPage.vue` 虽然已经不再直接持有 query / route 状态，但仍同时承担三类职责：

- tab 导航展示与键盘焦点处理
- 错误态 / workspace shell 边界
- hero + insight content 的区块装配

最小正确收口方式是：

- 让 `StudentAnalysisPage.vue` 退回真正的 shell owner，只保留 tab shell、错误边界和对下游子组件的桥接
- 把 tab 导航抽到 workspace 内部独立子组件
- 把 hero + insight content 装配抽到 workspace 内部独立 content 组件
- 通过本地共享 tab metadata 让 tab bar 与 panel content 使用同一份契约，而不是继续把这组 owner 混在 page shell 内部

本轮不做：

- 不改 `useStudentAnalysisPage.ts` 的 page model、路由同步或异步 workflow
- 不继续拆 `StudentInsightPanel.vue` 的 overview / recommendations / timeline 区块 owner
- 不改 teacher / platform route page 的外部 props / emits 契约

## Files to modify
- `.harness/reuse-decisions/student-analysis-page-shell-content-split.md`
- `docs/plan/impl-plan/2026-06-01-student-analysis-page-shell-content-split-plan.md`
- `docs/reviews/frontend/2026-06-01-student-analysis-page-shell-content-split-review.md`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspaceTabs.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspaceContent.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/studentAnalysisWorkspaceTabs.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/index.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `StudentAnalysisPage.vue` 只保留 workspace shell、错误态和对 tab/content 子组件的桥接。
- tab metadata 与键盘导航 owner 会收敛到 `StudentAnalysisWorkspaceTabs.vue` 和本地共享 helper。
- hero + insight content 的装配 owner 会收敛到 `StudentAnalysisWorkspaceContent.vue`，后续如果继续拆 `StudentInsightPanel.vue`，切口会更清楚。
