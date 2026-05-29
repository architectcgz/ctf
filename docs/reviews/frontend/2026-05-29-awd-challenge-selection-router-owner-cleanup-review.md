# AWD Challenge Selection Router Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-awd-challenge-selection-router-owner-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/awd-challenge-selection-router-owner-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-awd-challenge-selection-router-owner-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-awd-challenge-selection-router-owner-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/contest-awd-config/model/useAwdChallengeSelection.ts`
  - `code/frontend/src/features/contest-awd-config/model/useContestAwdConfigPage.ts`
  - `code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts`
- Classification check：预期按单条 feature router owner cleanup 处理；`useAwdChallengeSelection.ts` 不属于合理的 route-aware feature owner。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useContestAwdConfigPage.ts` 继续持有 `useRoute()` / `useRouter()` 合理，因为它本来就是 page owner。
- `useAwdChallengeSelection.ts` 应该只保留 service selection / reconcile / sorting owner，不应直接认识 `Router`。
- 这条 allowlist 如果继续保留，会给 selection/helper 层混入 router 开后门；本轮应当在 touched surface 内收掉。
- 当前实现已经把 `service` query 的读写回收到 page owner，selection helper 只保留 fallback / query reconcile / sorting 判断，服务切换与缺省回填行为保持不变。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/awd-challenge-selection-router-owner-cleanup.md docs/plan/impl-plan/2026-05-29-awd-challenge-selection-router-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-awd-challenge-selection-router-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/contest-awd-config/model/useAwdChallengeSelection.ts code/frontend/src/features/contest-awd-config/model/useContestAwdConfigPage.ts code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `featureRouterImportAllowlist` 还有大量条目，需要后续继续按“page owner / non-page owner”逐条判定。
