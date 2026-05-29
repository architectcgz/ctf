# Challenge List Route Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-challenge-list-route-owner-cleanup-plan.md`
- Scope：
  - `useChallengeListPage.ts`
  - `ChallengeList.vue`
  - `ChallengeDirectoryPanel.vue`
  - `ChallengeDirectoryRow.vue`
  - `routeQueryTransport.ts`
- Classification check：同意按“challenge list route owner cleanup”单独切片；这条同时触及 query sync 和薄导航，不适合和上一批 notification / scoreboard 的 route owner 混做。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `ChallengeList` 这条不该再用“新 route wrapper 承接老 router”来换壳；真正需要收口的是导航 contract 和 query transport owner。
- shared transport 只能提供 query 读取与替换，不应顺手吃掉 challenge-specific parse、默认值和刷新策略。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/challenges/__tests__/ChallengeList.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/challenge-list-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-challenge-list-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-challenge-list-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/composables/routeQueryTransport.ts code/frontend/src/features/challenge-list/model/challengeListRoutes.ts code/frontend/src/features/challenge-list/model/index.ts code/frontend/src/features/challenge-list/model/useChallengeListPage.ts code/frontend/src/views/challenges/ChallengeList.vue code/frontend/src/components/challenge/ChallengeDirectoryPanel.vue code/frontend/src/entities/challenge/ui/ChallengeDirectoryRow.vue code/frontend/src/views/challenges/__tests__/ChallengeList.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- shared query transport 目前只提供 `query` 读取和 `replaceQuery()` transport，没有继续吞进 challenge-specific parse / refresh policy；如果后续继续增长，应再评估是否抽成更通用的 route query state owner，而不是在 transport 上叠业务分支。
- 这份 review 是同上下文 self-review；独立 reviewer gate 仍未满足。
