# Student Journal 按钮变体架构收口实施计划

## Objective

把学生仪表盘难度面板的普通行动按钮从页面局部配色修复，收口为共享 `journal-soft-surface` 按钮变体。页面只表达按钮语义，不负责按钮明暗主题实现。

## Non-goals

- 不重写全站 `ui-btn` / `journal-btn` 的完整按钮体系。
- 不调整按钮文案、布局和点击行为。
- 不修改教师端、平台端按钮视觉。

## Source Inputs

- `code/frontend/src/assets/styles/journal-soft-surfaces.css`
- `code/frontend/src/components/dashboard/student/StudentDifficultyPage.vue`
- `code/frontend/src/views/__tests__/studentJournalButtonStyles.test.ts`
- `code/frontend/src/views/__tests__/studentJournalSoftStyles.test.ts`
- `docs/architecture/frontend/06-components.md`

## Task Slices

1. Shared variant
   - 在 `journal-soft-surfaces.css` 新增 `journal-btn-secondary` 共享变体。
   - 该变体负责边框、背景、文字、hover、focus 的主题变量落点。

2. Page adoption
   - `StudentDifficultyPage.vue` 的非主推档位按钮切换为 `journal-btn-secondary`。
   - 移除页面局部 `difficulty-action-item__cta--secondary` 配色规则。

3. Guardrails
   - 更新按钮共享样式测试，要求共享样式声明 secondary 变体。
   - 更新学生 soft style 测试，禁止难度面板继续保留页面私有 secondary 按钮配色。

4. Harness and architecture docs
   - 更新 `docs/architecture/frontend/06-components.md`，记录学生 soft journal 按钮变体 owner。
   - 更新 `feedback/` 记录，把问题从“页面桥接”收口为“共享按钮变体”。

## Compatibility Impact

只新增共享 CSS 类并替换目标按钮 class。已有 `journal-btn-primary`、`journal-btn-outline` 不变。

## Validation

- `npm run test:run -- src/views/__tests__/studentJournalButtonStyles.test.ts src/views/__tests__/studentJournalSoftStyles.test.ts`
- `npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts`
- `npx prettier --check src/assets/styles/journal-soft-surfaces.css src/components/dashboard/student/StudentDifficultyPage.vue src/views/__tests__/studentJournalButtonStyles.test.ts src/views/__tests__/studentJournalSoftStyles.test.ts`
- `bash scripts/check-consistency.sh`

## Review Focus

- 页面是否只选择共享按钮语义类。
- secondary 变体是否保留边框并通过主题变量适配 dark mode。
- 是否避免新增页面私有 dark selector 或 hardcoded light color。
