# Platform Challenge Route Page Convergence 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-platform-challenge-route-page-convergence-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/platform-challenge-route-page-convergence.md`
  - `docs/plan/impl-plan/2026-05-29-platform-challenge-route-page-convergence-plan.md`
  - `docs/reviews/frontend/2026-05-29-platform-challenge-route-page-convergence-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/platform-challenges/model/usePlatformChallengeRoutePage.ts`
  - `code/frontend/src/features/platform-challenges/model/useChallengeTopologyStudioRoutePage.ts`
  - `code/frontend/src/features/platform-challenges/model/useChallengeWriteupPage.ts`
  - `code/frontend/src/features/platform-challenges/model/useChallengeWriteupViewPage.ts`
  - `code/frontend/src/features/platform-challenges/model/index.ts`
  - `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
  - `code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts`
- Classification check：同意按 `platform-challenges` feature 内重复 route wrapper owner convergence 处理；旧三个 wrapper 已属于过薄重复 page helper，不应继续各自持有 `vue-router`。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `usePlatformChallengeRoutePage.ts` 现在是这组 platform challenge route-aware workflow 的唯一 router owner，统一承接 `challengeId`、返回详情页导航，以及查看态的“去题解编辑页”跳转。
- `useChallengeTopologyStudioRoutePage.ts`、`useChallengeWriteupPage.ts`、`useChallengeWriteupViewPage.ts` 已降为纯委托层，对外 public API 名称保持不变，route view 调用面没有扩散。
- `architectureAllowlist.ts` 已把这一组 3 条旧例外收敛成 1 条新 owner，allowlist 净减少 2 条。
- `ChallengeTopologyStudio.test.ts` 与 `ChallengeWriteup.test.ts` 已新增 raw-source 护栏，继续约束“route view 只组合、不直接拿 router”以及“旧 wrapper 不再直接 import vue-router”。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/platform/__tests__/ChallengeWriteup.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-challenge-route-page-convergence.md docs/plan/impl-plan/2026-05-29-platform-challenge-route-page-convergence-plan.md docs/reviews/frontend/2026-05-29-platform-challenge-route-page-convergence-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-challenges/model/usePlatformChallengeRoutePage.ts code/frontend/src/features/platform-challenges/model/useChallengeTopologyStudioRoutePage.ts code/frontend/src/features/platform-challenges/model/useChallengeWriteupPage.ts code/frontend/src/features/platform-challenges/model/useChallengeWriteupViewPage.ts code/frontend/src/features/platform-challenges/model/index.ts code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `usePlatformChallengeDetailRoutePage.ts` 仍保留在 `featureRouterImportAllowlist`，但它承接的是 detail route view 的实际 query-tab owner，这轮不一并处理。
