# Reuse Decision

## Change type
frontend refactor / router owner cleanup

## Existing code searched
- code/frontend/src/features/auth/model/useAuth.ts
- code/frontend/src/features/auth/model/useLoginPage.ts
- code/frontend/src/features/auth/model/useLoginPage.test.ts
- code/frontend/src/features/auth/model/useRegisterPage.ts
- code/frontend/src/features/auth/model/useRegisterPage.test.ts
- code/frontend/src/features/auth/model/useLoginViewPage.ts
- code/frontend/src/widgets/layout-shell/model/useLayoutSessionActionsBridge.ts
- code/frontend/src/views/auth/__tests__/LoginView.test.ts
- code/frontend/src/views/auth/__tests__/RegisterView.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `useNotificationDrawer.ts`、`useChallengeManagePresentation.ts` 这类下游 helper 已改成消费 callback / signal，不再自己持有 `vue-router`。
- `usePlatformUserManagePage.ts`、`useClassStudentsPage.ts` 这类 page owner 已经接住 query / redirect 导航，再把纯 helper 降回非 route owner。
- `widgets/layout-shell/useLayoutNotificationDrawerBridge.ts` 已作为 layout bridge 持有跨 feature 的 route 跳转 owner。

## Decision
refactor_existing

## Reason
`useAuth.ts` 当前把三类不同 workflow 的导航揉在一起：

- 登录成功后的 redirect
- 注册成功后的角色首页跳转
- 退出登录后的 `/login` 跳转

这让一个通用 auth helper 直接持有 `vue-router`，owner 落点偏高，也把 page / layout 的导航责任藏进 feature 里。

这轮按最小正确边界收口：

- `useAuth.ts` 退回纯 auth side-effect owner，只负责 API、auth store 和 toast
- 登录 redirect 回到 `useLoginPage.ts`
- 注册成功跳转回到 `useRegisterPage.ts`
- 退出登录跳转回到 `useLayoutSessionActionsBridge.ts`
- 删除 `features/auth/model/useAuth.ts -> vue-router` 这条 allowlist，并只为真正需要 router 的 page owner 增补条目

本轮不做：

- 不改 `useLoginViewPage.ts` 当前对 `redirect` query 的读取与 sanitize owner
- 不改登录 / 注册页面视觉和表单交互
- 不继续处理其它 `featureRouterImportAllowlist` 条目

## Files to modify
- .harness/reuse-decisions/auth-router-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-auth-router-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-auth-router-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/auth/model/useAuth.ts
- code/frontend/src/features/auth/model/useAuth.test.ts
- code/frontend/src/features/auth/model/useLoginPage.ts
- code/frontend/src/features/auth/model/useLoginPage.test.ts
- code/frontend/src/features/auth/model/useLoginViewPage.ts
- code/frontend/src/features/auth/model/useLoginViewPage.test.ts
- code/frontend/src/features/auth/model/useRegisterPage.ts
- code/frontend/src/features/auth/model/useRegisterPage.test.ts
- code/frontend/src/widgets/layout-shell/model/useLayoutSessionActionsBridge.ts
- code/frontend/src/widgets/layout-shell/model/useLayoutSessionActionsBridge.test.ts

## After implementation
- `useAuth.ts` 不再 import `vue-router`
- login / register / logout 的导航 owner 分别回到真实 workflow owner
- `featureRouterImportAllowlist` 收掉 `useAuth.ts`，并仅保留 page owner 层的 router 例外
- 登录、注册、退出登录流程的现有用户可见行为保持不变
