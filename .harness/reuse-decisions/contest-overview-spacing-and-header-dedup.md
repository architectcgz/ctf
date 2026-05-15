# Reuse Decision

## Change type
page / component / layout

## Existing code searched
- `code/frontend/src/components/contests/ContestOverviewPanel.vue`
- `code/frontend/src/views/contests/ContestDetail.vue`
- `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`
- `code/frontend/src/assets/styles/page-tabs.css`
- `code/frontend/src/style.css`
- `code/frontend/src/components/dashboard/student/StudentOverviewStyleEditorial.vue`
- `code/frontend/src/components/contests/ContestChallengeWorkspacePanel.vue`

## Similar implementations found
- `code/frontend/src/style.css` 已经提供 `workspace-panel-header`、`workspace-panel-divider` 作为工作区标题节奏和分隔线 owner，竞赛总览不应该继续维持独立的 header rail 和 divider 间距。
- `code/frontend/src/assets/styles/page-tabs.css` 已经提供 `workspace-tab-heading` / `workspace-tab-heading__main` 的共享标题结构，这次只需要在规则区做局部对齐覆盖，而不是再做一套 contest 专属 heading 组件。
- `code/frontend/src/components/dashboard/student/StudentOverviewStyleEditorial.vue` 已经验证“共享 divider + 页面局部 spacing token 微调”的模式可行，适合沿用到 contest overview。
- `code/frontend/src/components/contests/ContestChallengeWorkspacePanel.vue` 说明 contest 相关页面已经依赖 `ContestDetail.vue` 提供共享 section / divider 语义，overview 继续复用这条链比新增独立壳层更稳。

## Decision
refactor_existing

## Reason
这次不是新增竞赛总览组件体系，而是把 `ContestOverviewPanel` 收回现有 workspace header、divider 和 section spacing contract。重复的 header 元信息和 score rail 应该删除，分隔线应复用共享 token，规则区只做局部紧凑变体来修正单段文案的视觉松散与顶对齐问题。这样改动最小，也不会把 contest 页面继续推向一套新的局部布局语言。

## Files to modify
- `.harness/reuse-decisions/contest-overview-spacing-and-header-dedup.md`
- `code/frontend/src/components/contests/ContestOverviewPanel.vue`
- `code/frontend/src/views/contests/ContestDetail.vue`
- `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`

## After implementation
- 后续如果继续收竞赛详情其他 tab 的区块间距，优先复用 `workspace-panel-header`、`workspace-panel-divider` 和现有 section 变体，不要恢复 contest 专属分隔线和 header rail。
