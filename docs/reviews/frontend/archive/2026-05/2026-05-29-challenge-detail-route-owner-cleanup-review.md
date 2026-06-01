# Challenge Detail Route Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-challenge-detail-route-owner-cleanup-plan.md`
- Scope：
  - `challengeDetailRoutes.ts`
  - `useChallengeDetailPage.ts`
  - `ChallengeDetail.test.ts`
  - `architectureAllowlist.ts`
- Classification check：同意按“challenge detail route owner cleanup”单独切片；这条只涉及 params 读取和错误态返回列表导航，不需要新增 route wrapper。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useChallengeDetailPage.ts` 改为通过 `routeQueryTransport.ts` 读取 `challengeId`、通过 `routeNavigationTransport.ts` 执行返回列表导航，是这条最小且边界清楚的收口方式；challenge detail 的数据加载、预取和实例 workflow owner 没有被打散。
- `challengeDetailRoutes.ts` 只承接 `Challenges` 这一条薄 route target，没有把 challenge detail 自己的业务状态同步逻辑继续沉到 shared。
- `ChallengeDetail.test.ts` 既补了 raw-source 护栏，也补了错误态点击“返回题目列表”后的真实命名路由命中，能覆盖这次 transport owner 迁移的核心回归面。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/challenges/__tests__/ChallengeDetail.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/challenge-detail-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-challenge-detail-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-challenge-detail-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/challenge-detail/model/challengeDetailRoutes.ts code/frontend/src/features/challenge-detail/model/index.ts code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `useUrlSyncedTabs()` 仍通过 `window.location` 维护 panel query；本轮只收 `useChallengeDetailPage.ts` 的 direct router import，不处理这层长期演进。
- 这份 review 是同上下文 self-review；独立 reviewer gate 仍未满足。
