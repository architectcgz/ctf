# Student Analysis Review KPI Grid Convergence Plan

> 状态：Current
> 事实源：`studentInsightSections.css`、review KPI section consumers、teacher detail raw-source tests

## Objective

- 把 student-analysis review 三处 KPI grid 的列数和响应式折叠 contract 收口到 `studentInsightSections.css`。
- 清理 `writeups / manual review / attack sessions` 里仍散在本地的列数规则与 utility class。
- 保持三处 section 的 KPI 卡片内容、顺序和视觉风格不变。

## Non-goals

- 本轮不改 KPI card 的文案、颜色、图标和内容组织。
- 本轮不重构 writeup 列表、review workspace 或 state surface。
- 本轮不扩到 overview hero summary 或 recommendations 列表。

## Source Inputs

- `code/frontend/src/features/teaching/student-analysis-shared/ui/studentInsightSections.css`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.test.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.test.ts`

## Architecture Fit Check

- `studentInsightSections.css` 已经是 review section KPI family 的共享入口，当前 debt 不在缺入口，而在 consumer 仍继续持有列数 contract。
- `writeup-kpi-grid` 和 `md:grid-cols-*` 都是基础 grid contract，应该回到 shared owner，而不是留在各 section 本地。
- `attack sessions` 目前混入 `teacher-summary-grid` 只是为了 4 列，不属于真实 teacher summary family。

## Task Breakdown

### Slice 1: 扩 shared KPI grid contract

**Files**
- Modify: `code/frontend/src/features/teaching/student-analysis-shared/ui/studentInsightSections.css`

- [ ] Step 1: 为 review KPI grid 增加显式 shared 列数 class，例如 3 列 / 4 列变体。
- [ ] Step 2: 把窄屏折叠规则收回 shared CSS。

**Validation**
- raw-source 测试应检查 shared class，而不是本地 `writeup-kpi-grid` 或 `md:grid-cols-*`。

### Slice 2: 迁移 consumers 与测试

**Files**
- Modify: `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue`
- Modify: `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue`
- Modify: `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue`
- Modify: `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- Modify: `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.test.ts`
- Modify: `code/frontend/src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.test.ts`

- [ ] Step 1: 三处 KPI grid 改为显式 shared column class。
- [ ] Step 2: 删除 `writeup-kpi-grid` 与相关列数媒体查询。
- [ ] Step 3: 去掉 `attack sessions` 对 `teacher-summary-grid` 和 `md:grid-cols-*` 的列数依赖。
- [ ] Step 4: 更新测试护栏。

**Validation**
- `cd code/frontend && pnpm exec vitest run src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.test.ts src/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.test.ts`
- `cd code/frontend && pnpm typecheck`

## Review Focus

- KPI grid 的 3 列 / 4 列与窄屏折叠是否已经统一回到 shared owner。
- `attack sessions` 是否不再错误依赖 teacher summary family 只为拿列数。
- 测试是否真实保护 shared owner，而不是只改字符串。

## Rollback / Recovery

- 如果 writeups 与 attack sessions 的 KPI family 需求差异超出本轮范围，优先保留更窄的 shared variant class，不回退到 consumer 本地媒体查询。
- 若后续还有更多 review KPI 变体，再沿 `studentInsightSections.css` 扩展，不在本轮顺手扩到无关 section。
