# Student Timeline Workspace Directory Surface Alignment Plan

> 状态：Current
> 事实源：`TrainingTimelinePanel.vue`、`style.css`、student dashboard / teacher student analysis raw-source tests

## Objective

- 把训练记录列表区域的外框、空态和分页节奏收口到现有 `workspace-directory-*` 共享 contract。
- 保持 `TrainingTimelinePanel.vue` 继续作为 student dashboard 与 teacher 学员详情共用的单点展示 owner。
- 保持时间线的指标卡、分组逻辑、分页逻辑和事件内容不变。

## Non-goals

- 本轮不改 timeline 顶部 hero、指标卡文案、事件类型映射和分页行为。
- 本轮不把时间线抽成新的 shared shell，也不把样式下沉到 dashboard 或 teacher 调用页。
- 本轮不扩到 recommendations、overview、review workspace 等其他区块。

## Source Inputs

- `code/frontend/src/entities/training-timeline/ui/TrainingTimelinePanel.vue`
- `code/frontend/src/style.css`
- `code/frontend/src/features/challenge-list/ui/ChallengeDirectoryPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue`
- `code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts`
- `code/frontend/src/__tests__/studentJournalSoftStyles.test.ts`
- `code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`

## Architecture Fit Check

- 学生侧列表区域的基础外框 contract 已经存在于 `style.css` 的 `workspace-directory-section / list / pagination`，当前 debt 在 timeline 面板仍局部持有同类壳层样式。
- `TrainingTimelinePanel.vue` 已经是 dashboard 和 teacher 学员分析共享组件，因此列表壳层 owner 也应留在这个组件，而不是散落到两个 consumer。
- 本轮最小正确边界是“timeline 复用现有共享列表壳层，并只在本组件内保留时间线分组和事件行的局部变量桥接”，而不是再造一套 timeline-specific directory shell。

## Task Breakdown

### Slice 1: 补齐流程文档与 startup gate

**Files**
- Modify: `.harness/reuse-decisions/student-timeline-workspace-directory-surface-alignment.md`
- Create: `docs/plan/impl-plan/2026-06-04-student-timeline-workspace-directory-surface-alignment-plan.md`

- [ ] Step 1: 确认 reuse decision 与实现 owner 一致，plan 明确只收口列表壳层 contract。
- [ ] Step 2: 通过 `check-task-intake` 的 reuse startup gate。

**Validation**
- `bash scripts/check-task-intake.sh --reuse-decision student-timeline-workspace-directory-surface-alignment`

### Slice 2: 收口 timeline 列表壳层到 workspace-directory contract

**Files**
- Modify: `code/frontend/src/entities/training-timeline/ui/TrainingTimelinePanel.vue`

- [ ] Step 1: 让训练记录 section 复用 `workspace-directory-section` 和 `list-heading`。
- [ ] Step 2: 让空态、列表壳层、分页分别复用 `workspace-directory-empty`、`workspace-directory-list`、`workspace-directory-pagination`。
- [ ] Step 3: 删除 timeline 本地已经可由 shared contract 承担的裸外框定义，只保留分组和事件行表现需要的局部变量桥接。

**Validation**
- raw-source 测试应明确断言 `workspace-directory-*` contract 已在 timeline owner 生效。

### Slice 3: 更新护栏测试并做最小验证

**Files**
- Modify: `code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts`
- Verify: `code/frontend/src/__tests__/studentJournalSoftStyles.test.ts`
- Verify: `code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts`
- Verify: `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`

- [ ] Step 1: 更新 `studentUserSurfaceAlignment`，让它保护新的 shared list shell owner。
- [ ] Step 2: 确认 dashboard 与 teacher 学员详情调用链仍只依赖 `TrainingTimelinePanel.vue`。
- [ ] Step 3: 跑最小测试与 typecheck，再跑 workflow completion gate。

**Validation**
- `cd code/frontend && pnpm exec vitest run src/pages/__tests__/studentUserSurfaceAlignment.test.ts src/__tests__/studentJournalSoftStyles.test.ts src/pages/dashboard/__tests__/DashboardView.test.ts src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `cd code/frontend && pnpm typecheck`
- `bash scripts/check-workflow-complete.sh`

## Review Focus

- timeline 的列表外框 contract 是否已经单点收口到 `TrainingTimelinePanel.vue`，而不是 consumer 覆盖。
- 本地 CSS 是否只保留 timeline 分组和事件线表现所必需的差异，不再重复 shared shell 基础样式。
- 测试是否真实保护了 `workspace-directory-*` 的 shared owner，而不是只让字符串对上。

## Rollback / Recovery

- 如果时间线的分组表现确实需要偏离通用列表壳层，优先在 `TrainingTimelinePanel.vue` 内通过局部变量桥接，不回退到 dashboard / teacher 页面双份覆盖。
- 如果后续发现更多学生侧列表区域存在同类 contract 漂移，另起任务继续按 owner family 收口，不在本轮顺手扩面。
