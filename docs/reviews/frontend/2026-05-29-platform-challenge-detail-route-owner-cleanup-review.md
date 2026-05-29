# Platform Challenge Detail Route Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-platform-challenge-detail-route-owner-cleanup-plan.md`
- Scope：
  - `platformChallengeDetailRoutes.ts`
  - `usePlatformChallengeDetailPage.ts`
  - `ChallengeDetail.test.ts`
  - `architectureAllowlist.ts`
- Classification check：同意按“platform challenge detail route owner cleanup”单独切片；这条只是在 detail page owner 内收 params 读取与薄导航，不需要再造新的 feature 外 wrapper。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- detail 页本地新增 `platformChallengeDetailRoutes.ts` 是合理的；它描述的是 challenge detail workflow 自己的返回题库、去拓扑和去题解动作，不应依赖 `platform-challenges` feature 的内部 helper。
- `usePlatformChallengeDetailPage.ts` 改为通过 `routeQueryTransport.ts` 读取 `challengeId`、通过 `routeNavigationTransport.ts` 执行导航后，Flag draft、附件下载和失败跳回计时器 owner 都仍留在 page model，没有被错误拆散。
- 加载失败后的延迟跳回现在也复用本地 route target contract，避免成功路径和失败路径的导航语义分叉。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeDetail.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-challenge-detail-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-platform-challenge-detail-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-platform-challenge-detail-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-challenge-detail/model/index.ts code/frontend/src/features/platform-challenge-detail/model/platformChallengeDetailRoutes.ts code/frontend/src/features/platform-challenge-detail/model/usePlatformChallengeDetailPage.ts code/frontend/src/views/platform/__tests__/ChallengeDetail.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 目前 challenge authoring 相关 route target 分别落在 `platform-challenges` 和 `platform-challenge-detail` 两个 feature，本轮是按 owner 分开收口；如果后续再出现第三处相同导航，再评估是否需要抽成更中立的 authoring route contract owner。
- 这份 review 是同上下文 self-review；独立 reviewer gate 仍未满足。
