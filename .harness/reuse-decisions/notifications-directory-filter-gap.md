# Reuse Decision

## Change type
frontend spacing token extraction / notification directory layout

## Existing code searched
- `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`
- `code/frontend/src/features/notifications/ui/NotificationCategoryFilter.vue`
- `code/frontend/src/style.css`
- `code/frontend/src/features/scoreboard/ui/ScoreboardWorkspaceShell.vue`
- `code/frontend/src/widgets/contest-list-workspace/ContestListWorkspace.vue`

## Similar implementations found
- `style.css` 里 `student-directory-shell__head + ...` 与 `student-directory-filters + ...` 已经负责目录壳层内部区块间距，但当前直接写死 `var(--space-4)`。
- `notifications` 页面复用了 `student-directory-shell` 与 `student-directory-filters`，说明间距 owner 应该仍在共享目录壳层，而不是单独塞进 `NotificationCategoryFilter`。
- 其他目录页同样复用这套组合选择器，适合先把间距抽成共享 token，再由具体页面按需覆写。

## Decision
refactor_existing

## Reason
这次问题不是分类选择框组件本身，而是共享目录壳层缺少“标题分隔线到下一块内容”的抽象 token，导致页面只能继承固定间距，无法按场景细调。

最小正确改动是：

- 给 `student-directory-shell` 提供共享的内部区块间距 token。
- 让头部到筛选区、筛选区到列表区都使用这个 token。
- 在 `notifications` 目录壳层局部覆写该 token。
- 对 `notification-filter-section` 内部“分类筛选块 / 统计块”之间的间距保留页面 owner，并通过局部 token 明确控制，不把它错误提升到共享目录壳层。

本轮不做：

- 不改 `NotificationCategoryFilter` 结构。
- 不改全站所有目录页的视觉值。
- 不改 workspace page header 或 content pane 的共享节奏。

## Files to modify
- `.harness/reuse-decisions/notifications-directory-filter-gap.md`
- `code/frontend/src/style.css`
- `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`

## After implementation
- 通知页目录头部分隔线与分类筛选之间会有更合理的呼吸感。
- 通知页筛选块与统计块之间的间距会由页面局部 token 明确控制。
- 这类目录壳层间距会有明确 token owner，后续页面可以按需覆写。
- 不会把 spacing 补丁散落到筛选子组件内部。
