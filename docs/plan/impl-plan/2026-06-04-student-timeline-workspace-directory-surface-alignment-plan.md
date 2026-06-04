# Student Timeline Workspace Directory Surface Alignment Plan

> 状态：Current
> 事实源：`TrainingTimelinePanel.vue`、`TrainingTimelineContent.vue`、`StudentInsightTimelineSection.vue`、`style.css`、student dashboard / teacher student analysis raw-source tests

## Objective

- 把训练记录重构成“共享内容 owner + 学生页壳 + 教师洞察壳”的明确边界。
- 把训练记录列表区域的外框、空态和分页节奏继续收口到现有 `workspace-directory-*` 共享 contract。
- 保持时间线的指标卡、分组逻辑、分页逻辑和事件内容不变。

## Non-goals

- 本轮不改 timeline 顶部文案、事件类型映射和分页行为。
- 本轮不扩到 recommendations、overview、review workspace 等其他区块。
- 本轮不改学生 dashboard 其他 panel 的 embedded 形态。

## Source Inputs

- `code/frontend/src/entities/training-timeline/ui/TrainingTimelinePanel.vue`
- `code/frontend/src/entities/training-timeline/ui/TrainingTimelineContent.vue`
- `code/frontend/src/style.css`
- `code/frontend/src/features/challenge-list/ui/ChallengeDirectoryPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightTimelineSection.vue`
- `code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts`
- `code/frontend/src/__tests__/studentJournalSoftStyles.test.ts`
- `code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`

## Architecture Fit Check

- 学生侧列表区域的基础外框 contract 已经存在于 `style.css` 的 `workspace-directory-section / list / pagination`，当前 debt 不在“有没有复用”，而在 `TrainingTimelinePanel.vue` 同时承担学生页壳与教师工作台壳。
- `embedded` 已经不是单纯样式开关，而是在切两套布局语义：page shell、padding、section gap、divider 与 list owner。
- 本轮最小正确边界是：
  - `TrainingTimelineContent.vue` 单点持有时间线 header、指标卡、列表壳、分页和事件分组表现
  - `TrainingTimelinePanel.vue` 回归学生 dashboard 页面壳
  - `StudentInsightTimelineSection.vue` 承接教师学员分析页的 section 壳与 loading skeleton

## Task Breakdown

### Slice 1: 补齐流程文档与 startup gate

**Files**
- Modify: `.harness/reuse-decisions/student-timeline-workspace-directory-surface-alignment.md`
- Create: `docs/plan/impl-plan/2026-06-04-student-timeline-workspace-directory-surface-alignment-plan.md`

- [ ] Step 1: 确认 reuse decision 与实现 owner 一致，plan 明确只收口列表壳层 contract。
- [ ] Step 2: 通过 `check-task-intake` 的 reuse startup gate。

**Validation**
- `bash scripts/check-task-intake.sh --reuse-decision student-timeline-workspace-directory-surface-alignment`

### Slice 2: 拆 timeline 内容 owner 并移除 mixed embedded 语义

**Files**
- Modify: `code/frontend/src/entities/training-timeline/ui/TrainingTimelinePanel.vue`
- Create: `code/frontend/src/entities/training-timeline/ui/TrainingTimelineContent.vue`
- Create: `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightTimelineSection.vue`
- Modify: `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue`

- [ ] Step 1: 抽出 `TrainingTimelineContent.vue`，单点持有时间线 header、指标卡、列表壳和分页。
- [ ] Step 2: 让 `TrainingTimelinePanel.vue` 只保留学生页壳，不再接受 `embedded`。
- [ ] Step 3: 新增 `StudentInsightTimelineSection.vue`，承接教师学员分析页的 section 节奏和 loading skeleton。
- [ ] Step 4: 保持空态、列表壳层、分页继续复用 `workspace-directory-empty`、`workspace-directory-list`、`workspace-directory-pagination`。

**Validation**
- raw-source 测试应明确断言 `workspace-directory-*` contract 已在 `TrainingTimelineContent.vue` 生效。

### Slice 3: 更新护栏测试并做最小验证

**Files**
- Modify: `code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts`
- Modify: `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- Modify: `code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts`
- Modify: `code/frontend/src/__tests__/metricPanelSurfaceOwnership.test.ts`
- Verify: `code/frontend/src/__tests__/studentJournalSoftStyles.test.ts`
- Verify: `code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts`
- Verify: `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`

- [ ] Step 1: 更新 raw-source 护栏，让内容断言指向 `TrainingTimelineContent.vue`，壳层断言分别保护学生壳与教师壳。
- [ ] Step 2: 确认 dashboard 继续依赖 `TrainingTimelinePanel.vue`，teacher 学员详情改为依赖 `StudentInsightTimelineSection.vue`。
- [ ] Step 3: 跑最小测试与 typecheck，再跑 workflow completion gate。

**Validation**
- `cd code/frontend && pnpm exec vitest run src/pages/__tests__/studentUserSurfaceAlignment.test.ts src/__tests__/studentJournalSoftStyles.test.ts src/pages/dashboard/__tests__/DashboardView.test.ts src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `cd code/frontend && pnpm typecheck`
- `bash scripts/check-workflow-complete.sh`

## Review Focus

- timeline 的列表外框 contract 是否已经单点收口到 `TrainingTimelineContent.vue`，而不是 consumer 覆盖。
- `TrainingTimelinePanel.vue` 与 `StudentInsightTimelineSection.vue` 是否各自只承载自己的壳层职责。
- 测试是否真实保护了“共享内容 + 两个壳”的 owner 拆分，而不是只让字符串对上。

## Rollback / Recovery

- 如果时间线的分组表现确实需要偏离通用列表壳层，优先在 `TrainingTimelineContent.vue` 内通过局部变量桥接，不回退到 dashboard / teacher 页面双份覆盖。
- 如果后续发现更多学生侧列表区域存在同类 contract 漂移，另起任务继续按 owner family 收口，不在本轮顺手扩面。
