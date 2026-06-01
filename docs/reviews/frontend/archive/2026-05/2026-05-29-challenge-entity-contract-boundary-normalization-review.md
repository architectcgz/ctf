# Challenge Entity Contract Boundary Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-challenge-entity-contract-boundary-normalization-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/challenge-entity-contract-boundary-normalization.md`
  - `docs/plan/impl-plan/2026-05-29-challenge-entity-contract-boundary-normalization-plan.md`
  - `docs/reviews/frontend/2026-05-29-challenge-entity-contract-boundary-normalization-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/entities/challenge/model/*`
  - `code/frontend/src/entities/challenge/ui/*`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
- Classification check：同意按 entity 展示层 contract owner 收口处理，风险主要在本地类型定义是否过窄，以及 allowlist / 边界测试同步。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `entities/challenge/model/presentation.ts` 已不再直接依赖 `@/api/contracts`，题目实体的 category / difficulty / status / instance sharing 展示语义已经回收到 entity 本地类型。
- `ChallengeCategoryDifficultyPills.vue`、`ChallengeCategoryPill.vue`、`ChallengeCategoryText.vue`、`ChallengeDifficultyText.vue`、`ChallengeDirectoryRow.vue`、`ChallengeMetaStrip.vue`、`ChallengeProfileMetaGrid.vue`、`ChallengeProfileSummaryStrip.vue` 已改为使用 entity 自己的最小字段接口，不再直接吃 API DTO 类型。
- 这刀没有把 feature / page 消费面扩大成新的 adapter 链；外部仍直接传结构兼容对象，改动面停留在 entity 内。
- `commonForbiddenImportAllowlist` 目前只剩 `components/common/InstancePanel.vue -> @/api/contracts`，challenge entity 这组历史例外已经收口。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/challenges/__tests__/ChallengeList.test.ts src/views/platform/__tests__/challengeDetailPanelExtraction.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/__tests__/studentUserSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `InstancePanel.vue` 这条 `commonForbiddenImportAllowlist` 还在，后续如果继续清 shared/common 边界，应单独判断它属于“本地展示类型可继续下沉”还是“common 组件合理持有基础契约”。
