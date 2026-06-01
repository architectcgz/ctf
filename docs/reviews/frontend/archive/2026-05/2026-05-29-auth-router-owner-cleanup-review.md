# Auth Router Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-auth-router-owner-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/auth-router-owner-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-auth-router-owner-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-auth-router-owner-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/auth/model/useAuth.ts`
  - `code/frontend/src/features/auth/model/useAuth.test.ts`
  - `code/frontend/src/features/auth/model/useLoginPage.ts`
  - `code/frontend/src/features/auth/model/useLoginPage.test.ts`
  - `code/frontend/src/features/auth/model/useLoginViewPage.ts`
  - `code/frontend/src/features/auth/model/useLoginViewPage.test.ts`
  - `code/frontend/src/features/auth/model/useRegisterPage.ts`
  - `code/frontend/src/features/auth/model/useRegisterPage.test.ts`
  - `code/frontend/src/widgets/layout-shell/model/useLayoutSessionActionsBridge.ts`
  - `code/frontend/src/widgets/layout-shell/model/useLayoutSessionActionsBridge.test.ts`
- Classification check：同意按单条 feature router owner cleanup 处理；`useAuth.ts` 不属于合理的长期 route-aware feature owner。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useAuth.ts` 作为跨登录 / 注册 / 退出登录复用的 auth helper，应只保留 auth side effect，不应继续直接认识 `vue-router`。
- 登录和注册成功后的跳转属于 page workflow owner，回到 `useLoginPage.ts`、`useRegisterPage.ts` 更合理。
- 退出登录后跳回登录页属于 layout session action，回到 `useLayoutSessionActionsBridge.ts` 更合理。
- `useLoginViewPage.ts` 如果继续持有 redirect query 读取与 sanitize，就会让 auth 线仍然留着一个非完整 page owner 的 router 例外；本轮一起把它降成纯 target helper 是对的。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/features/auth/model/useAuth.test.ts src/features/auth/model/useLoginPage.test.ts src/features/auth/model/useLoginViewPage.test.ts src/features/auth/model/useRegisterPage.test.ts src/widgets/layout-shell/model/useLayoutSessionActionsBridge.test.ts src/views/auth/__tests__/LoginView.test.ts src/views/auth/__tests__/RegisterView.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/auth-router-owner-cleanup.md docs/plan/impl-plan/2026-05-29-auth-router-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-auth-router-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/auth/model/useAuth.ts code/frontend/src/features/auth/model/useAuth.test.ts code/frontend/src/features/auth/model/useLoginPage.ts code/frontend/src/features/auth/model/useLoginPage.test.ts code/frontend/src/features/auth/model/useLoginViewPage.ts code/frontend/src/features/auth/model/useLoginViewPage.test.ts code/frontend/src/features/auth/model/useRegisterPage.ts code/frontend/src/features/auth/model/useRegisterPage.test.ts code/frontend/src/widgets/layout-shell/model/useLayoutSessionActionsBridge.ts code/frontend/src/widgets/layout-shell/model/useLayoutSessionActionsBridge.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `featureRouterImportAllowlist` 还有大量条目，需要继续按“page owner / non-page owner”逐条收口。
