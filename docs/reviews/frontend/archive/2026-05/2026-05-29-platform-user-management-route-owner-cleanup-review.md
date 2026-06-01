# Platform User Management Route Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-platform-user-management-route-owner-cleanup-plan.md`
- Scope：
  - `usePlatformUserManagePage.ts`
  - `UserManage.test.ts`
  - `architectureAllowlist.ts`
- Classification check：同意按“platform user management route owner cleanup”单独切片；这条只有 `panel` query hydrate 与 replace transport，不需要再引入 user-management-specific route wrapper。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `platform-user-management` 这条的核心不是把 panel owner 从 page model 再搬出去，而是把 `useRoute / useRouter` transport 从 page model 下沉到已有的共享 query transport。
- 复用 `routeQueryTransport.ts` 是合理的，因为它只承接 query 读取与 `replaceQuery()`，没有把 user governance 自己的 panel normalize、mounted refresh 或筛选节流平移进 shared。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/UserManage.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-user-management-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-platform-user-management-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-platform-user-management-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-user-management/model/usePlatformUserManagePage.ts code/frontend/src/views/platform/__tests__/UserManage.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `routeQueryTransport.ts` 当前仍只应承接 transport；如果后续用户治理页继续增长出更复杂的 query canonicalization，不应继续堆到 shared composable。
- 这份 review 是同上下文 self-review；独立 reviewer gate 仍未满足。
