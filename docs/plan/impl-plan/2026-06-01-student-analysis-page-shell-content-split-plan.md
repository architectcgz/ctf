# Student Analysis Page Shell Content Split 计划

## Objective

- 把 `StudentAnalysisPage.vue` 收口成真正的 workspace shell owner。
- 将 tab 导航和 content 装配拆成明确的内部子组件，减少 page shell 内部混合职责。

## Non-goals

- 不改 `useStudentAnalysisPage.ts` 的 route/query 同步、请求编排或 review workflow。
- 不修改 teacher / platform route page 对 `StudentAnalysisPage` 的外部 props / emits 契约。
- 不继续拆 `StudentInsightPanel.vue` 或 `StudentAnalysisWorkspacePage.vue`。

## Source Inputs

- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisOverviewHeroPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspacePage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- 当前最小且能继续压 owner 面的切口不是再动 page model，而是让 `StudentAnalysisPage.vue` 退回 shell，去掉对 tab metadata / tabpanel content 的双重内聚。
- 这轮拆分仍然留在 `student-analysis-workspace/ui` 内部，不引入新的 feature 边界，因此 review 成本和回归范围可控。

## Task Slices

### Slice 1: 抽离 tab 导航 owner

- 目标：新增 `StudentAnalysisWorkspaceTabs.vue` 与本地共享 tab metadata，承接 tab 按钮列表与键盘导航。
- 风险：tab button id / panel id / `aria` 契约必须保持不变，否则 teacher runtime test 会回归。

### Slice 2: 抽离 content 装配 owner

- 目标：新增 `StudentAnalysisWorkspaceContent.vue`，承接 active panel、overview hero 和 `StudentInsightPanel` 组合。
- 风险：overview tab 当前要同时显示 hero 与 overview section，不能在拆分时改变现有呈现。

### Slice 3: 收口 page shell 与测试护栏

- 目标：让 `StudentAnalysisPage.vue` 只保留 shell、错误边界和对子组件事件桥接，并同步 teacher/platform source-boundary 测试与 backlog 进展。
- 风险：如果只移动模板不补边界断言，后续容易再次把 tab metadata 或 content assembly 写回 page shell。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision student-analysis-page-shell-content-split`
- `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/student-analysis-page-shell-content-split.md docs/plan/impl-plan/2026-06-01-student-analysis-page-shell-content-split-plan.md docs/reviews/frontend/2026-06-01-student-analysis-page-shell-content-split-review.md code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspaceTabs.vue code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspaceContent.vue code/frontend/src/features/teaching/student-analysis-workspace/ui/studentAnalysisWorkspaceTabs.ts code/frontend/src/features/teaching/student-analysis-workspace/ui/index.ts code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review Focus

- `StudentAnalysisPage.vue` 是否真正退回 shell owner，而不是继续混放 tab metadata / tabpanel content。
- tab button / panel 的可访问性 id、焦点导航与现有 external contract 是否保持稳定。
- `StudentAnalysisWorkspaceContent.vue` 是否只承接内容装配，不把 route-aware page model 或 async workflow 下沉进去。
