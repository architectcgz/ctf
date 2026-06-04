# Teacher Summary Contract Convergence Plan

> 状态：Current
> 事实源：`teacher-surface.css`、teacher dashboard summary consumers、student-analysis review tests、frontend tech debt backlog

## Objective

- 清理 `student-analysis` 中仍指向旧 KPI class 的测试护栏。
- 把 `TeacherDashboardPage`、`TeacherDashboardPortraitPanel`、`TeacherDashboardTrendPanel` 里重复的 summary strip 基础样式收口到 dashboard feature 共享 owner。
- 保持 teacher summary family 继续以 `teacher-surface.css` 为跨页面基础 owner，不新增第二套全局 teacher summary 规则。

## Non-goals

- 本轮不处理 admin / platform summary family。
- 本轮不重构 `ClassManagementPage`、`StudentManagementPage`、`TeacherInstanceHeroPanel` 的 summary markup；这三页已经主要依赖 `teacher-surface.css`。
- 本轮不扩到 `class-students` 顶部 summary 和 `review-archive` summary tone；它们保留在各自 feature / widget 语义内，后续单独收口。

## Source Inputs

- `code/frontend/src/assets/styles/teacher-surface.css`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPortraitPanel.vue`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardTrendPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.test.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Architecture Fit Check

- teacher 页通用 summary shell 已经在 `teacher-surface.css` 里定义，说明跨页面基础 owner 不缺；当前 debt 在 dashboard feature 内部仍重复声明 `summary-grid` / `summary-note` 基础 contract。
- `student-analysis` 当前实现已经切到 `student-insight-*` 共享 class，残留问题只在测试 owner 漂移，适合和本轮一并收掉。
- 这轮的最小正确边界是“在 dashboard feature 内补共享 owner，并让测试跟随真实 owner”，而不是把所有 teacher summary 立刻抽成新的全局 CSS。

## Task Breakdown

### Slice 1: 修正 student-analysis 旧测试类名漂移

**Files**
- Modify: `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.test.ts`
- Modify: `code/frontend/src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.test.ts`

- [ ] Step 1: 把 `insight-kpi-grid` 相关断言改成 `student-insight-kpi-grid`。
- [ ] Step 2: 确认这些测试不再锁定已经退场的旧 class 字面量。

**Validation**
- `cd code/frontend && pnpm exec vitest run src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.test.ts src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.test.ts`

### Slice 2: 收口 teacher dashboard feature 内部 summary note contract

**Files**
- Create: `code/frontend/src/features/teacher/dashboard/ui/teacherDashboardSummary.css`
- Modify: `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue`
- Modify: `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPortraitPanel.vue`
- Modify: `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardTrendPanel.vue`
- Modify: `code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts`

- [ ] Step 1: 抽出 dashboard overview summary 与 note summary 的共享 grid / card / copy contract。
- [ ] Step 2: 让 page / portrait / trend 三个 consumer 引用共享样式，只保留各自真正的局部布局差异。
- [ ] Step 3: 更新 raw-source 护栏测试，确认共享 owner 生效且旧局部重复规则退场。

**Validation**
- `cd code/frontend && pnpm exec vitest run src/pages/teacher/__tests__/TeacherDashboard.test.ts`

### Slice 3: 集成验证

**Validation**
- `cd code/frontend && pnpm exec vitest run src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.test.ts src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.test.ts src/pages/teacher/__tests__/TeacherDashboard.test.ts`
- `cd code/frontend && pnpm typecheck`

## Review Focus

- student-analysis 测试是否完全跟随真实共享 KPI owner，而不是局部绿灯。
- dashboard summary 共享 CSS 是否只承接重复基础 contract，没有重新发明第二套 teacher global surface。
- `TeacherDashboardPage`、`TeacherDashboardPortraitPanel`、`TeacherDashboardTrendPanel` 是否都只保留各自局部布局差异。

## Rollback / Recovery

- 如果 dashboard 三处 summary 语义差异超出当前共享 CSS 能承接的范围，优先退到 feature 内更窄的两个 shared class，不回退到三处各自复制。
- 如果测试暴露出更多 teacher summary family 的旧护栏漂移，本轮先记录为下一批，不扩到 admin / platform 线。
