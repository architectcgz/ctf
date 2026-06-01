# User Governance Panel Route Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-user-governance-panel-route-owner-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/user-governance-panel-route-owner-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-user-governance-panel-route-owner-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-user-governance-panel-route-owner-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.ts`
  - `code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.test.ts`
  - `code/frontend/src/features/platform-user-management/model/usePlatformUserManagePage.ts`
  - `code/frontend/src/features/platform-user-management/model/index.ts`
  - `code/frontend/src/features/platform-user-management/ui/UserGovernancePage.vue`
  - `code/frontend/src/views/platform/UserManage.vue`
  - `code/frontend/src/views/platform/__tests__/UserManage.test.ts`
- Classification check：预期按单一 feature 的 router owner cleanup 处理；`useUserGovernancePanelRoute.ts` 不是合理长期 router owner。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useUserGovernancePanelRoute.ts` 继续持有 `vue-router` 会让 UI shell 间接带上 route owner 身份，这个落点不对。
- `usePlatformUserManagePage.ts` 已经是 `UserManage` 的 page owner，把 panel query 读写并回这里更符合当前架构。
- 这轮的结构收益在于“allowlist owner 从 helper 迁到 page owner”，而不是纯数量下降。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/features/platform-user-management/model/useUserGovernancePanelRoute.test.ts src/views/platform/__tests__/UserManage.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/user-governance-panel-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-user-governance-panel-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-user-governance-panel-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.ts code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.test.ts code/frontend/src/features/platform-user-management/model/usePlatformUserManagePage.ts code/frontend/src/features/platform-user-management/model/index.ts code/frontend/src/features/platform-user-management/ui/UserGovernancePage.vue code/frontend/src/views/platform/UserManage.vue code/frontend/src/views/platform/__tests__/UserManage.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `usePlatformUserManagePage.ts` 后续若继续增长，仍可能需要再拆局部 workflow helper。
