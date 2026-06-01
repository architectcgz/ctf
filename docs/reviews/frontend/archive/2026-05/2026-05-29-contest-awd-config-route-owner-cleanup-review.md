# Contest AWD Config Route Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-contest-awd-config-route-owner-cleanup-plan.md`
- Scope：
  - `contestAwdConfigRoutes.ts`
  - `useContestAwdConfigPage.ts`
  - `ContestAwdConfig.test.ts`
  - `architectureAllowlist.ts`
- Classification check：同意按“contest awd config route owner cleanup”单独切片；这条只收 params/query 与返回工作台导航，不需要再引入新的 page wrapper。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useContestAwdConfigPage.ts` 改成通过 `routeQueryTransport.ts` 读取 `contestId` 和 `service`，以及通过 `replaceQuery()` 回写 service query，是这条最小且清晰的收口方式；没有把 AWD 配置页自己的 fallback 逻辑塞进 shared。
- 返回工作台动作下沉到本地 `contestAwdConfigRoutes.ts` 后，page model 仍继续持有 mounted 初始化、breadcrumb、checker draft / preview / save owner，职责边界没有被打散。
- `useAwdChallengeSelection.ts` 已经在上一轮去掉 router，本轮再把 page owner 里的 route 读写一起收掉后，AWD 配置页这条 allowlist 的 owner 会更完整。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-awd-config-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-contest-awd-config-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-contest-awd-config-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/contest-awd-config/model/contestAwdConfigRoutes.ts code/frontend/src/features/contest-awd-config/model/index.ts code/frontend/src/features/contest-awd-config/model/useContestAwdConfigPage.ts code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `service` query 当前只是一层简单字符串 contract；如果后续 AWD 配置页需要更多 query 语义，再评估是否要把 parse/normalize builder 抽成本地 helper，而不是现在过早增加抽象。
- 这份 review 是同上下文 self-review；独立 reviewer gate 仍未满足。
