# Platform Challenges Route Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-platform-challenges-route-owner-cleanup-plan.md`
- Scope：
  - `platformChallengeRoutes.ts`
  - `useChallengeManagePage.ts`
  - `usePlatformChallengeRoutePage.ts`
  - `ChallengeManage.test.ts`
  - `ChallengeTopologyStudio.test.ts`
  - `ChallengeWriteup.test.ts`
  - `architectureAllowlist.ts`
- Classification check：同意按“platform challenges route owner cleanup”单独切片；这组只是在同一 feature 里收 route param 读取与薄导航，不需要再引入新的 feature 外 wrapper。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `platformChallengeRoutes.ts` 把 import preview、detail、topology、writeup panel、writeup editor、import manage 这些目标路由集中在 feature 内部描述，比继续在 page model 里散着拼 route object 更清楚，也避免跨 feature 借路由 helper。
- `useChallengeManagePage.ts` 改为只持有列表、排序、筛选与题目动作 owner，并通过 `routeNavigationTransport` 执行导航；route 语义仍留在本 feature，而没有漂到 shared transport。
- `usePlatformChallengeRoutePage.ts` 现在只通过 `routeQueryTransport` 读取 `challengeId`，并通过本地 route target helper 返回详情或跳转题解编辑页；topology / writeup 的 mode 语义继续留在 route page owner，没有被抽成只包 router 的中间层。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeManage.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/platform/__tests__/ChallengeWriteup.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-challenges-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-platform-challenges-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-platform-challenges-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-challenges/model/index.ts code/frontend/src/features/platform-challenges/model/platformChallengeRoutes.ts code/frontend/src/features/platform-challenges/model/useChallengeManagePage.ts code/frontend/src/features/platform-challenges/model/usePlatformChallengeRoutePage.ts code/frontend/src/views/platform/__tests__/ChallengeManage.test.ts code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `platformChallengeRoutes.ts` 后续如果继续增长，应继续只保留 route target contract，不要把 mode 判定、加载或错误策略塞进去。
- 这份 review 是同上下文 self-review；独立 reviewer gate 仍未满足。
