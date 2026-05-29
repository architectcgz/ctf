# Challenge Manage Presentation Router Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-challenge-manage-presentation-router-owner-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/challenge-manage-presentation-router-owner-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-challenge-manage-presentation-router-owner-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-challenge-manage-presentation-router-owner-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/platform-challenges/model/useChallengeManagePresentation.ts`
  - `code/frontend/src/features/platform-challenges/model/useChallengeManagePage.ts`
  - `code/frontend/src/views/platform/__tests__/ChallengeManage.test.ts`
- Classification check：同意按单条 feature router owner cleanup 处理；`useChallengeManagePresentation.ts` 不属于合理的 route-aware feature owner。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useChallengeManagePage.ts` 继续持有 `useRouter()` 合理，因为它本来就是 page owner。
- `useChallengeManagePresentation.ts` 只保留 presentation / action menu / action sequencing owner 更合理。
- 这条 allowlist 如果删不掉，就会持续给后续 presentation/helper 混入 router 开后门；本轮应当在 touched surface 内收掉。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ChallengeManage.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/challenge-manage-presentation-router-owner-cleanup.md docs/plan/impl-plan/2026-05-29-challenge-manage-presentation-router-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-challenge-manage-presentation-router-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-challenges/model/useChallengeManagePresentation.ts code/frontend/src/features/platform-challenges/model/useChallengeManagePage.ts code/frontend/src/views/platform/__tests__/ChallengeManage.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `featureRouterImportAllowlist` 还有大量条目，需要后续继续按“page owner / non-page owner”逐条判定。
