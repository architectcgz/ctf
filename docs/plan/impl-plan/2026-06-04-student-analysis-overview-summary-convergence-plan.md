# Student Analysis Overview Summary Convergence Plan

> 状态：Current
> 事实源：`studentInsightSurface.css`、`StudentAnalysisOverviewHeroPanel.vue`、teacher detail raw-source tests

## Objective

- 把 `StudentAnalysisOverviewHeroPanel` 顶部 summary 的基础展示 contract 收口到 `studentInsightSurface.css`。
- 清理 student-analysis hero 仍散在组件本地的 header-attached summary 间距与响应式列数规则。
- 保持当前 student-analysis 顶部 summary 的视觉、文案和加载态不变。

## Non-goals

- 本轮不改 summary card 的色彩与骨架样式。
- 本轮不扩到 review workspace、writeups KPI 或 recommendations 列表。
- 本轮不重做 student-analysis 页面整体布局。

## Source Inputs

- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisOverviewHeroPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-shared/ui/studentInsightSurface.css`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`

## Architecture Fit Check

- student-analysis 顶部 glass / state surface 已经有共享 CSS owner，summary 顶部条带继续扩这个入口最合理。
- 当前 debt 不在缺少 shared 目录，而在 consumer 本地继续持有 `padding: 0` 和列数折叠 contract。
- raw-source 测试应改为保护 shared owner，而不是继续接受局部 `.summary-strip`。

## Task Breakdown

### Slice 1: 扩 student-analysis shared summary contract

**Files**
- Modify: `code/frontend/src/features/teaching/student-analysis-shared/ui/studentInsightSurface.css`

- [ ] Step 1: 增加顶部 summary 共享 grid class，承接 header attached summary 间距与列数变量。
- [ ] Step 2: 把窄屏折叠规则回收到 shared CSS。

**Validation**
- raw-source 测试应断言 shared CSS owner，而不是组件本地 `.summary-strip`。

### Slice 2: 迁移 overview hero consumer 与测试

**Files**
- Modify: `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisOverviewHeroPanel.vue`
- Modify: `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

- [ ] Step 1: 让 `StudentAnalysisOverviewHeroPanel` 改为显式声明新的 shared summary class。
- [ ] Step 2: 删除本地 `.summary-strip` 基础 contract，只保留 summary card 自身视觉。
- [ ] Step 3: 更新 teacher detail raw-source 护栏测试。

**Validation**
- `cd code/frontend && pnpm exec vitest run src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `cd code/frontend && pnpm typecheck`

## Review Focus

- 顶部 summary 的 `padding: 0` 和列数响应式是否已经回到 shared owner。
- hero 组件是否只保留 summary card 自身视觉 token，而不再持有基础 grid contract。
- 测试是否真正保护共享 owner，而不是仅修改字符串断言。

## Rollback / Recovery

- 如果 student-analysis hero 的 summary 与其他 section KPI family 差异过大，优先保留更窄的 student-analysis shared class，不回退到组件本地。
- 若后续发现别的 student-analysis hero 区块复用同类 contract，再按 shared owner 扩展，不在本轮顺手扩面。
