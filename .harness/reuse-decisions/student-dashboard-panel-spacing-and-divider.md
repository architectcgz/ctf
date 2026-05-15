# Reuse Decision

## Change type
component / layout / page

## Existing code searched
- code/frontend/src/components/dashboard/student/StudentOverviewStyleEditorial.vue
- code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue
- code/frontend/src/components/dashboard/student/StudentCategoryProgressPage.vue
- code/frontend/src/components/dashboard/student/StudentDifficultyPage.vue
- code/frontend/src/components/dashboard/student/StudentTimelinePage.vue
- code/frontend/src/assets/styles/theme.css
- code/frontend/src/assets/styles/workspace-shell.css
- code/frontend/src/style.css
- code/frontend/src/views/__tests__/pageTabsStyles.test.ts
- code/frontend/src/views/__tests__/spacingSystemTokens.test.ts
- code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts

## Similar implementations found
- `StudentTimelinePage.vue` 已经把学生 dashboard header、摘要卡片和下方 divider 的节奏做成了当前最稳定的参考面板，不应该让 `overview / recommendation / category / difficulty` 继续各自维护不同的 `mt / pt / border-top`。
- `style.css` 已经开始承接 `workspace-panel-header` 的标题、文案和块间距语义，说明 panel 内部垂直节奏的 owner 应该继续收口在共享层，而不是回退到页面私有样式。
- `theme.css` 与 `workspace-shell.css` 已经承担 workspace tab rail 到 panel 内容区的全局 spacing token，新的 header-body divider 也应沿这条 token 链扩展，而不是新开局部 hardcode。
- `pageTabsStyles.test.ts`、`spacingSystemTokens.test.ts`、`studentUserSurfaceAlignment.test.ts` 已经是这类 tab/panel 共享样式的回归断言入口，适合继续扩展，而不是新增一组分散的页面私有测试。

## Decision
refactor_existing

## Reason
这次不是新增新的 dashboard 版式，而是把学生仪表盘各个 tab 已经存在的 header、divider 和 tab 下方面板间距，统一回收到现有 workspace token 与 shared style owner。

最小正确做法是：
- 继续复用 `StudentTimelinePage.vue` 作为 header-body divider 节奏参考；
- 继续复用 `style.css` 里的 `workspace-panel-header` 共享 contract；
- 在 `theme.css` / `workspace-shell.css` 现有 `space-workspace-*` 体系里补充 panel divider gap；
- 把 `overview / recommendation / category / difficulty` 从各自的外层 `mt-6 / pt-5 / border-top / ::before` 收回到同一个 shared divider class。

这样可以让学生 dashboard 各 tab 的结构节奏保持同一个 owner，避免后续每个页面再次分叉出新的 spacing 例外。

## Files to modify
- .harness/reuse-decisions/student-dashboard-panel-spacing-and-divider.md
- code/frontend/src/assets/styles/theme.css
- code/frontend/src/assets/styles/workspace-shell.css
- code/frontend/src/style.css
- code/frontend/src/components/dashboard/student/StudentOverviewStyleEditorial.vue
- code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue
- code/frontend/src/components/dashboard/student/StudentCategoryProgressPage.vue
- code/frontend/src/components/dashboard/student/StudentDifficultyPage.vue
- code/frontend/src/components/dashboard/student/StudentTimelinePage.vue
- code/frontend/src/views/__tests__/pageTabsStyles.test.ts
- code/frontend/src/views/__tests__/spacingSystemTokens.test.ts
- code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts

## After implementation
- 学生 dashboard 的 tab 下方面板间距、header title/copy/actions/summary 节奏和 header-body divider 由共享 token 与 shared class 统一承接。
- 后续再调整学生 dashboard 的 panel 间距时，优先改 `theme.css` / `style.css` 的共享 token 和 class，而不是重新写页面私有 `mt`、`pt` 或外层边线。
