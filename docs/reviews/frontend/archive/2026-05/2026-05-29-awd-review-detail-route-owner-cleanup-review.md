# AWD Review Detail Route Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-awd-review-detail-route-owner-cleanup-plan.md`
- Scope：
  - `awdReviewDetailRoutes.ts`
  - `useAwdReviewDetailPage.ts`
  - `PlatformAwdReviewDetail.test.ts`
  - `architectureAllowlist.ts`
- Classification check：同意按“awd review detail route owner cleanup”单独切片；这条以 params/query 读取和返回索引导航为主，不需要再引入额外 page wrapper。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useAwdReviewDetailPage.ts` 改为通过 `routeQueryTransport.ts` 读取 `contestId / round`、通过 `replaceQuery()` 写回 round query，是这条最小且边界清楚的收口方式；当前页 query owner 仍留在 detail page。
- 返回索引页动作下沉到本地 `awdReviewDetailRoutes.ts` 后，detail page 继续只持有数据加载、summary 聚合、team drawer、export polling 和 breadcrumb owner，没有被拆成新的中间层。
- `route` 本体从返回值里移除后，消费面没有丢失行为，说明这层 page model 对外 contract 更干净了。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/awd-review-detail-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-awd-review-detail-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-awd-review-detail-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/awd-review-detail-workspace/model/awdReviewDetailRoutes.ts code/frontend/src/features/awd-review-detail-workspace/model/index.ts code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `teachingWorkspaceRouting.ts` 仍是 AWD review route name 的单点来源；本轮只收 transport owner，不处理 route naming 的长期归位。
- 这份 review 是同上下文 self-review；独立 reviewer gate 仍未满足。
