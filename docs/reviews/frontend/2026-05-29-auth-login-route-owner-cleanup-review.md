# Auth Login Route Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-auth-login-route-owner-cleanup-plan.md`
- Scope：
  - `useLoginPage.ts`
  - `useLoginPage.test.ts`
  - `router/guards.ts`
  - `utils/redirectPath.ts`
  - `architectureAllowlist.ts`
- Classification check：同意按“auth login route owner cleanup”单独切片；这条既包含 route transport owner，也包含 redirect sanitize owner 的回位，单独切干净比保留中间态更合理。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useLoginPage.ts` 改为通过 `routeQueryTransport.ts` 读取 redirect query、通过 `routeNavigationTransport.ts` 执行登录成功跳转后，登录页继续只持有表单态、提交 guard、probe 提示和默认 dashboard fallback owner，不再直接碰 router。
- `sanitizeRedirectPath()` 提升到 `utils/redirectPath.ts` 后，router guards 与 login page 复用同一实现，消除了 feature 对 `router/guards.ts` 的反向依赖，同时没有丢掉 open redirect 防护。
- `useLoginPage.test.ts` 既覆盖了默认 dashboard、显式 redirect，也补了恶意 redirect 输入的 sanitize 行为，因此这轮最关键的交互与安全语义都有回归护栏。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/features/auth/model/useLoginPage.test.ts src/router/__tests__/guards.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/auth-login-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-auth-login-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-auth-login-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/auth/model/useLoginPage.ts code/frontend/src/features/auth/model/useLoginPage.test.ts code/frontend/src/router/guards.ts code/frontend/src/utils/redirectPath.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `resolveLoginRedirectTarget()` 仍留在 `useLoginViewPage.ts` 这个历史命名 helper 文件中；本轮只收 allowlist，不顺带处理命名债。
- 这份 review 是同上下文 self-review；独立 reviewer gate 仍未满足。
