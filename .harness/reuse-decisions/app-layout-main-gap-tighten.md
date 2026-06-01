# Reuse Decision

## Change type
layout spacing refinement / shared workspace shell

## Existing code searched
- `code/frontend/src/views`
- `code/frontend/src/components`
- `code/frontend/src/features`
- `code/frontend/src/widgets`
- `code/frontend/src/composables`
- `code/frontend/src/api`
- `code/frontend/src/assets/styles/workspace-shell.css`
- `code/frontend/src/assets/styles/journal-user-shell.css`
- `code/frontend/src/pages/scoreboard/ScoreboardViewRoutePage.vue`
- `code/frontend/src/router/routes/studentRoutes.ts`

## Similar implementations found
- `studentRoutes.ts`、`teacherRoutes.ts`、`platformRoutes.ts` 里的业务页几乎都落在 `workspace-shell` 上，student/admin 还会再叠加 `journal-hero` 视觉层。
- `workspace-shell.css` 是所有页面壳共用的默认边框、阴影、tabs 间距 owner。
- `journal-user-shell.css`、`journal-admin-shell.css` 会在共享壳之上再追加 hero 阴影；如果要全局统一成“简单边线分隔”，这些额外 hero 阴影也要一起收口。

## Decision
refactor_existing

## Reason
这次目标已经明确成“所有页统一修改”，因此不应该再依赖 `AppLayout` 的 route bridge，也不需要页面变体；最小正确改动就是直接修改共享页面壳默认规则。

最小正确改动是：

- 撤回 `AppLayout.vue` 里前面用于试探的 bleed bridge 改法，避免布局层反向 deep 改页面实现。
- 在 `workspace-shell.css` 直接把共享页面壳改成贴边默认值：去掉顶边、左边和默认阴影，并收掉 `top-tabs` 自带的顶部偏移。
- 在 `journal-user-shell.css`、`journal-admin-shell.css` 去掉额外 hero 阴影，避免 student/admin 页面还继续像一张悬浮卡片。
- 保持各页自己的 `content-pane`、tabs 和正文节奏不动，只收口“页面壳与顶栏 / 侧栏如何相接”这一层。

本轮不做：

- 不改 `TopNav` 自身高度或边框。
- 不改非 bleed 布局页面。
- 不重写各页自己的 tabs / content / table / card 内部 spacing。
- 不把 route page 模板结构改成另一套布局组件。

## Files to modify
- `.harness/reuse-decisions/app-layout-main-gap-tighten.md`
- `code/frontend/src/assets/styles/workspace-shell.css`
- `code/frontend/src/assets/styles/journal-user-shell.css`
- `code/frontend/src/assets/styles/journal-admin-shell.css`

## After implementation
- 所有共享 `workspace-shell` 页面会统一改成更贴顶栏 / 侧栏的默认壳，不再像左上缺一块。
- 分隔主要靠顶栏底边和页壳自身的简单边线，而不是靠顶层卡片阴影制造层次。
- 后续页面只要继续复用共享 shell，就会自动吃到这次全局规则，不需要再走 route bridge 或逐页补丁。
