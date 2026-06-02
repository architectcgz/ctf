# Reuse Decision

## Change type
frontend style alignment / scoreboard header background

## Existing code searched
- `code/frontend/src/features/scoreboard/ui/ScoreboardWorkspaceShell.vue`
- `code/frontend/src/style.css`
- `code/frontend/dist/assets/ScoreboardWorkspaceShell-C883_qjt.css`
- `/usr/share/nginx/html/assets/ScoreboardWorkspaceShell-BKi9xY9Q.css` in `ctf-frontend-new-15173`

## Similar implementations found
- `5173` 当前实际生效的 `ScoreboardWorkspaceShell-BKi9xY9Q.css` 把视觉底色挂在 `.student-directory-list-heading__body`，而不是整个 `header`。
- `student-directory-shell__head` 的基础布局和分隔线由 `code/frontend/src/style.css` 统一 owner，不适合为了单页观感局部改成整块面板底色。

## Decision
refactor_existing

## Reason
这次目标不是重做 scoreboard 目录头部样式，而是让 `15173` 和 `5173` 的现有视觉 owner 对齐。

最小正确改动是：

- 保持 `student-directory-shell__head` 继续只负责布局、底边和间距。
- 让 `ScoreboardWorkspaceShell.vue` 只给 `.student-directory-list-heading__body` 提供和 `5173` 一致的背景、圆角和内边距。
- 重新 build 并同步到 `ctf-frontend-new-15173`，确保容器不再继续引用旧的 scoreboard CSS。

本轮不做：

- 不调整 scoreboard 列表布局、筛选器或卡片样式。
- 不扩展到其他 `student-directory-shell` 页面。
- 不改共享 token 或全局目录壳背景规则。

## Files to modify
- `.harness/reuse-decisions/scoreboard-header-bg-align.md`
- `code/frontend/src/features/scoreboard/ui/ScoreboardWorkspaceShell.vue`

## After implementation
- `15173` 的 scoreboard 标题底色 owner 会和 `5173` 保持一致。
- 目标视觉变化由 `student-directory-list-heading__body` 提供，不再依赖整块 `header` override。
- 容器里的 scoreboard route 资源会切到最新构建产物。
