# Student Analysis Page Owner Cleanup 计划

## Objective

- 把 teacher / platform 学员分析页的 page shell owner 从 `pages/teacher/TeacherStudentAnalysisRoutePage.vue` 与 `pages/platform/PlatformStudentAnalysisRoutePage.vue` 收回 `features/teaching/student-analysis-workspace/ui`。
- 保持 `useStudentAnalysisPage()` 的共享 owner 命名，不误改成单角色 page model。
- 保持学员分析页的 tab、review workspace、writeup moderation、review archive export 和导出默认班级行为不变。

## Non-goals

- 不移动文件路径。
- 不改 `useStudentAnalysisPage.ts` 的数据加载、route target、review workspace query 或 moderation workflow。
- 不改 `StudentAnalysisPage.vue` 的展示结构、tab 交互或 section 组合方式。
- 不改 `features/teaching/class-report-export` 的导出流程 owner。

## Source Inputs

- `code/frontend/src/pages/teacher/TeacherStudentAnalysisRoutePage.vue`
- `code/frontend/src/pages/platform/PlatformStudentAnalysisRoutePage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- 这条线优先收口共享 route page owner，不继续拆 student analysis workspace workflow。
- `useStudentAnalysisPage()` 同时被 teacher / platform route 消费，因此这次不做角色化命名，避免把共享 page model 错改成单角色 owner。
- route page 退回薄壳后，共享 feature page 要成为唯一组合层；不能只是把模板复制到第二个壳文件里，继续让 route page 持有 `ClassReportExportDialog` 或 page model。

## Task Slices

### Slice 1: 新增共享 student analysis page shell

- 目标：在 `features/teaching/student-analysis-workspace/ui` 下新增 `StudentAnalysisWorkspacePage.vue`，由它直接调用共享 `useStudentAnalysisPage()` 并组合 `StudentAnalysisPage` 与导出对话框。
- 风险：需要保持 teacher / platform 两侧的导出默认班级、review archive、writeup moderation 和 tab 事件契约不变。

### Slice 2: teacher / platform route page 退回薄壳

- 目标：`TeacherStudentAnalysisRoutePage.vue` 与 `PlatformStudentAnalysisRoutePage.vue` 只渲染共享 feature page，并分别传入自己的根类名。
- 风险：raw-source 护栏需要同时更新 teacher / platform 两侧，否则后续容易回流到 route page 直连 page model。

### Slice 3: 补共享 owner 护栏与 backlog 进展

- 目标：更新 teacher / platform 测试与 backlog 进展，明确 student analysis page model 继续保持共享 owner 命名。
- 风险：只补 owner 护栏，不重写已有运行态用例。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision student-analysis-page-owner-cleanup`
- `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `git diff --check -- .harness/reuse-decisions/student-analysis-page-owner-cleanup.md docs/plan/archive/impl-plan/2026-06/2026-06-01-student-analysis-page-owner-cleanup-plan.md code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspacePage.vue code/frontend/src/features/teaching/student-analysis-workspace/ui/index.ts code/frontend/src/pages/teacher/TeacherStudentAnalysisRoutePage.vue code/frontend/src/pages/platform/PlatformStudentAnalysisRoutePage.vue code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Review Focus

- teacher / platform route page 是否真正退回薄壳，而不是继续持有导出弹窗或 page model。
- 共享 feature page 是否成为 student analysis 的唯一组合层。
- 保持共享 page model 命名后，teacher / platform 两侧的 tab、review workspace、writeup moderation 与导出入口是否保持不变。

## Rollback / Recovery

- 如果 route page 或 raw-source 护栏还漏了旧组合方式，可以继续补公共导出或断言。
- 不能回退成 teacher / platform route page 继续直接组合 `ClassReportExportDialog` 和 `useStudentAnalysisPage()`。
