# Challenge Package Format Router Target Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-challenge-package-format-router-target-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/challenge-package-format-router-target-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-challenge-package-format-router-target-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-challenge-package-format-router-target-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/challenge-package-import/model/useChallengePackageFormatPage.ts`
  - `code/frontend/src/features/challenge-package-import/model/index.ts`
  - `code/frontend/src/features/challenge-package-import/index.ts`
  - `code/frontend/src/views/platform/ChallengePackageFormat.vue`
  - `code/frontend/src/views/platform/__tests__/ChallengePackageFormat.test.ts`
- Classification check：同意按单条 feature router owner cleanup 处理；`useChallengePackageFormatPage.ts` 不应继续为了单次返回动作保留 `vue-router`。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useChallengePackageFormatPage.ts` 之前只是一次性 `router.push()` wrapper，继续保留为 route-aware composable 的收益很低。
- 这类稳定返回目标更适合收口成纯 route target contract，再由 route view 用 `RouterLink` 消费。
- 本轮是直接删除一条 allowlist，而不是把同样的导航再移到别的 wrapper 上。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ChallengePackageFormat.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/challenge-package-format-router-target-cleanup.md docs/plan/impl-plan/2026-05-29-challenge-package-format-router-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-challenge-package-format-router-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/challenge-package-import/model/useChallengePackageFormatPage.ts code/frontend/src/features/challenge-package-import/model/index.ts code/frontend/src/features/challenge-package-import/index.ts code/frontend/src/views/platform/ChallengePackageFormat.vue code/frontend/src/views/platform/__tests__/ChallengePackageFormat.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- challenge package import 线后续仍可能继续处理 import manage / preview page 的 route owner，但不属于这轮范围。
