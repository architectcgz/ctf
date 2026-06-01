# Utility Contract Boundary Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-utility-contract-boundary-normalization-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/utility-contract-boundary-normalization.md`
  - `docs/plan/impl-plan/2026-05-29-utility-contract-boundary-normalization-plan.md`
  - `docs/reviews/frontend/2026-05-29-utility-contract-boundary-normalization-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/utils/contest.ts`
  - `code/frontend/src/utils/platformContestAwdChallengeLinks.ts`
  - `code/frontend/src/utils/skillProfile.ts`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
- Classification check：同意按 utility contract owner 收口处理，风险主要在本地最小类型是否与现有 consumer 保持结构兼容。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `utils/contest.ts` 已改为本地 contest status / mode 类型，不再直接依赖 `@/api/contracts`。
- `utils/skillProfile.ts` 已改为本地 skill profile / recommendation contract，API wrapper 和 feature consumer 继续通过结构兼容对象使用，不需要改 import 路径。
- `utils/platformContestAwdChallengeLinks.ts` 已改为本地 AWD service / challenge link contract，mapper 仍保持现有输出字段集合，能继续支撑 contest workbench 和 AWD admin 这两条消费链。
- `utilityBoundaryImportAllowlist` 已清空；当前 allowlist A 里 shared/common/utility 这几组 API contract 例外已经全部收口。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/utility-contract-boundary-normalization.md docs/plan/impl-plan/2026-05-29-utility-contract-boundary-normalization-plan.md docs/reviews/frontend/2026-05-29-utility-contract-boundary-normalization-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/utils/contest.ts code/frontend/src/utils/platformContestAwdChallengeLinks.ts code/frontend/src/utils/skillProfile.ts code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts code/frontend/src/__tests__/architectureAllowlist.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `featureRouterImportAllowlist` 和 `composableMultiBoundaryAllowlist` 还在，它们是否属于真实历史债需要单独逐条判断，不适合和这轮 utility contract 收口混在一起。
