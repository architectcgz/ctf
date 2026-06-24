# SectionCard Style Contract Convergence Plan

> 状态：Current
> 事实源：`SectionCard` shared owner、student analysis consumer、review archive consumer、`:deep` guard allowlist

## Objective

- 把 student analysis / review archive 这条链上对 `SectionCard` 的深度样式覆盖收口成显式组件 contract。
- 删除这批 consumer 上的 `:deep(.section-card*)`。
- 保持教师分析页与复盘归档页现有视觉和交互不变，只改变样式 owner。

## Non-goals

- 不处理 topology studio 的 `SectionCard` 深度样式，这一组视觉语义与 teacher surfaces 不同，另开切片。
- 不处理 `ModalTemplateShell`、`SlideOverDrawer`、`CFocusedInputDialog`、`CActionMenu` 的 slot-style contract。
- 不改 student analysis / review archive 的数据流、route owner、事件 contract 或业务文案。

## Source Inputs

- `code/frontend/src/shared/ui/common/SectionCard.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspaceContent.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveWorkspace.vue`
- `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveSummarySection.vue`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `code/frontend/scripts/vue-deep-allowlist.json`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Architecture Fit Check

- 当前 debt 的真实 owner 不在 student analysis page 壳，而在 `SectionCard` 的共享 contract 缺失。
- 这轮不把“父组件覆盖子组件内部结构”继续当作正常样式模式，而是收口到 `SectionCard` 自己可声明的变体 / CSS variable contract。
- touched surface 上的已知 debt 是 student analysis / review archive 对 `.section-card*` 的深度覆盖；本轮要把这组 debt 从当前 touched surface 清掉，不能只加新 allowlist。

## File Ownership Map

- `code/frontend/src/shared/ui/common/SectionCard.vue`
  - 负责共享 section 壳结构与显式样式 contract。
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/*`
  - 负责 student analysis 的 teacher-flat section card 使用方式，不再负责反向改 `SectionCard` 内部 DOM。
- `code/frontend/src/features/teaching/student-analysis-review/ui/*`
  - 负责 review / writeup / evidence 相关 section card consumer 选择。
- `code/frontend/src/widgets/review-archive-workspace/*`
  - 负责 review archive 的 teacher-surface section card consumer 选择。
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
  - 负责原始源码级 guardrail，确认这批页面不再回流 `.section-card*` 深度覆盖。

## Task Breakdown

### Slice 1: 给 SectionCard 增加显式样式 contract

**Files**
- Modify: `code/frontend/src/shared/ui/common/SectionCard.vue`

- [ ] Step 1: 为 `SectionCard` 增加稳定的变体或 CSS variable contract，覆盖 header / body / border / surface 这些当前被 `:deep` 反向修改的点。
- [ ] Step 2: 保持默认样式对 topology 和其他现有 consumer 向后兼容。
- [ ] Step 3: 明确 teacher-flat / teacher-surface 两组语义，不把 review archive 和 student analysis 混成无名本地类。

**Validation**
- 读取 `SectionCard.vue`，确认默认 contract 仍可被旧 consumer 使用。
- 聚焦 raw-source 测试覆盖 teacher variants。

### Slice 2: 把 student analysis consumer 改成显式 SectionCard contract

**Files**
- Modify: `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- Modify: `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspaceContent.vue`
- Modify: `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.vue`
- Modify: `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue`
- Modify: `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue`
- Modify: `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue`
- Modify: `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue`
- Modify: `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

- [ ] Step 1: 在实际使用 `SectionCard` 的 consumer 上显式声明 teacher-flat variant。
- [ ] Step 2: 把 overview / tab section 的局部差异改成 root class + CSS variable override，不再写 `:deep(.section-card*)`。
- [ ] Step 3: 删除 `StudentInsightPanel.vue` 和 `StudentAnalysisWorkspaceContent.vue` 里对 `.section-card*` 的父级深度覆盖。
- [ ] Step 4: 更新 teacher detail surface raw-source 断言，让它检查新 contract，而不是继续锁定 `:deep` 字面量。

**Validation**
- `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `cd code/frontend && npm run check:vue-deep`

### Slice 3: 把 review archive consumer 改成显式 SectionCard contract

**Files**
- Modify: `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveWorkspace.vue`
- Modify: `code/frontend/src/widgets/review-archive-workspace/ReviewArchiveSummarySection.vue`
- Modify: `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- Modify: `code/frontend/scripts/vue-deep-allowlist.json`
- Create: `docs/reviews/frontend/2026-06-04-section-card-style-contract-convergence-review.md`

- [ ] Step 1: 在 review archive summary 的 `SectionCard` consumer 上显式声明 teacher-surface variant。
- [ ] Step 2: 删除 workspace 对 `.section-card*` 的深度覆盖，保留页面级 token owner。
- [ ] Step 3: 更新 allowlist 与 review 证据，确认这批 `:deep` 已退场。

**Validation**
- `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `cd code/frontend && npm run check:vue-deep`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-frontend-architecture.sh --quick`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`

## Recommended Execution Order

1. 先扩 `SectionCard` contract。
2. 再迁移 student analysis consumer。
3. 最后迁移 review archive consumer，并同步 allowlist / review 证据。

## Review Focus

- `SectionCard` 新 contract 是否真的覆盖了当前被 `:deep` 修改的 header / body / border / surface 需求。
- student analysis / review archive 是否已经改为“声明 contract”，而不是换个位置继续反向改内部 class。
- 默认 `SectionCard` 是否保持对 topology 等现有 consumer 的兼容。
- 原始源码断言与 `:deep` allowlist 是否真实反映了收口后的现状。

## Rollback / Recovery

- 如果 `SectionCard` contract 无法同时兼容默认样式与 teacher variants，优先拆成更窄的显式 variant，不回退到新的 `:deep`。
- 如果 student analysis 与 review archive 的视觉需求差异超出当前 batch，可先保留同一 contract 下的 CSS variable override，不在本轮扩展到 topology。
