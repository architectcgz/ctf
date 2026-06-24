# Embedded Layout Owner Retirement Plan

> 状态：Current
> 事实源：`StudentRecommendationPage.vue`、`StudentCategoryProgressPage.vue`、`StudentDifficultyPage.vue`、`StudentOverviewStyleEditorial.vue`、`TrainingTimelinePanel.vue`、student dashboard raw-source tests

## Objective

- 清理学生 dashboard 中仍然使用 `embedded` 同时切页面壳与内容 owner 的 panel。
- 把这些 panel 统一收口为“共享内容组件 + dashboard 调用页壳”的结构。
- 将“不要用 `embedded` 切两套布局语义”沉淀到项目 harness 规则。

## Non-goals

- 本轮不处理 `ChallengeWriteupEditorPage.vue` 这类明确存在完整页 / 嵌入编辑器双入口的组件。
- 本轮不扩展到非 dashboard 学生页、教师页或平台页的所有 `embedded` 搜索结果。
- 本轮不改变现有 panel 的交互文案、排序逻辑、CTA 行为和数据契约。

## Source Inputs

- `code/frontend/src/features/student-dashboard/ui/StudentRecommendationPage.vue`
- `code/frontend/src/features/student-dashboard/ui/StudentCategoryProgressPage.vue`
- `code/frontend/src/features/student-dashboard/ui/StudentDifficultyPage.vue`
- `code/frontend/src/features/student-dashboard/ui/StudentOverviewStyleEditorial.vue`
- `code/frontend/src/entities/training-timeline/ui/TrainingTimelinePanel.vue`
- `code/frontend/src/features/student-dashboard/model/useStudentDashboardPanelBindings.ts`
- `code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts`
- `code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts`
- `code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts`
- `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`
- `ctf/AGENTS.md`
- `feedback/AGENTS.md`

## Architecture Fit Check

- 时间线已经证明 `embedded` 在这些 dashboard panel 里不是单纯视觉嵌入，而是在切换完整页壳、padding、divider、summary grid 和 body 节奏。
- 这类组件属于 student dashboard 的内容面板，不应该自己再持有“完整页”和“tab 内嵌面板”两套 owner 语义。
- 最小正确边界是：
  - 每个 panel 抽一个共享内容组件，持有 header + body + 局部表现
  - dashboard 调用层只负责 tab 面板壳
  - 如果未来有完整页入口，再单独用页面壳包共享内容，而不是回到 `embedded`

## Task Breakdown

### Slice 1: dashboard 现存 mixed-owner embedded 清理

**Files**
- Modify: `code/frontend/src/features/student-dashboard/ui/StudentRecommendationPage.vue`
- Modify: `code/frontend/src/features/student-dashboard/ui/StudentCategoryProgressPage.vue`
- Modify: `code/frontend/src/features/student-dashboard/ui/StudentDifficultyPage.vue`
- Modify: `code/frontend/src/features/student-dashboard/ui/StudentOverviewStyleEditorial.vue`
- Modify: `code/frontend/src/features/student-dashboard/model/useStudentDashboardPanelBindings.ts`
- Add: `code/frontend/src/features/student-dashboard/ui/*Content.vue`（按实际拆分命名）

- [ ] Step 1: 抽出 4 个 panel 的共享内容 owner。
- [ ] Step 2: 去掉 dashboard panel bindings 中对这些 panel 的 `embedded: true` 传参。
- [ ] Step 3: 让 dashboard tab 面板直接消费共享内容 owner，而不是让 panel 自己切两套壳。

**Validation**
- raw-source 测试应证明 dashboard panel 已不再通过 `embedded` 切 page shell。

### Slice 2: harness 规则沉淀

**Files**
- Modify: `ctf/AGENTS.md`
- Add or Modify: `feedback/2026-06-04-embedded-should-not-switch-layout-owner.md`
- Modify: `.harness/reuse-decisions/embedded-layout-owner-retirement.md`

- [ ] Step 1: 在项目 AGENTS 里写明：前端不得新增用 `embedded` 切 page shell / section shell / 内容 owner 的实现。
- [ ] Step 2: 在 feedback 里记录这次经验，明确例外范围和 owner 判断规则。

**Validation**
- `bash scripts/check-consistency.sh`

### Slice 3: 护栏测试和完成验证

**Files**
- Modify: `code/frontend/src/pages/__tests__/studentUserSurfaceAlignment.test.ts`
- Modify: `code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts`
- Modify: `code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts`
- Modify: `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`
- Modify: `code/frontend/src/__tests__/studentJournalSoftStyles.test.ts`

- [ ] Step 1: 更新 raw-source 护栏到“共享内容 owner + dashboard 壳”结构。
- [ ] Step 2: 跑最小测试、typecheck 和 workflow completion gate。

**Validation**
- `cd code/frontend && pnpm exec vitest run src/pages/dashboard/__tests__/DashboardView.test.ts src/pages/__tests__/studentUserSurfaceAlignment.test.ts src/pages/__tests__/workspacePageHeaderStyles.test.ts src/pages/__tests__/workspaceShellStyles.test.ts src/__tests__/studentJournalSoftStyles.test.ts`
- `cd code/frontend && pnpm typecheck`
- `bash scripts/check-workflow-complete.sh`

## Review Focus

- 是否还有 dashboard panel 在用 `embedded` 同时切 page shell 与内容节奏。
- 新抽出的内容组件是否只持有内容 owner，没有重新引入新的壳层分支。
- AGENTS / feedback 是否足够明确，能阻止后续继续新增同类实现。

## Rollback / Recovery

- 如果某个 panel 后续确实需要“完整页 + 内嵌 panel”双入口，优先用共享内容组件配两个显式壳，而不是恢复 `embedded`。
- 对 `ChallengeWriteupEditorPage.vue` 这类暂未纳入的双入口组件，后续单独起任务分析，不在本轮混改。
