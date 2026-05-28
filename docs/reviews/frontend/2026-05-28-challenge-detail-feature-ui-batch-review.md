# Challenge Detail Feature UI Batch 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-challenge-detail-feature-ui-batch-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/challenge-detail-feature-ui-batch.md`
    - `docs/plan/impl-plan/2026-05-28-challenge-detail-feature-ui-batch-plan.md`
    - `docs/reviews/frontend/2026-05-28-challenge-detail-feature-ui-batch-review.md`
    - `code/frontend/src/features/challenge-detail/index.ts`
    - `code/frontend/src/features/challenge-detail/ui/index.ts`
    - `code/frontend/src/features/challenge-detail/ui/ChallengeWorkspaceShell.vue`
    - `code/frontend/src/features/challenge-detail/ui/ChallengeSolutionsPanel.vue`
    - `code/frontend/src/features/challenge-detail/ui/ChallengeSubmissionRecordsPanel.vue`
    - `code/frontend/src/views/challenges/ChallengeDetail.vue`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/components.d.ts`
    - `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`
    - `code/frontend/src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts`
    - `code/frontend/src/views/challenges/__tests__/challengeDetailSharedShell.test.ts`
    - `code/frontend/src/views/__tests__/pageTabsStyles.test.ts`
    - `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
    - `docs/architecture/features/社区题解与推荐题解设计.md`
    - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 `challenge-detail` 单一 feature 私有 UI 的 owner 迁位和 allowlist 收口。
- Gate verdict：Implemented and re-validated
- Review mode：same-context review
- Independent review gate：未执行；当前回合没有显式 delegation 授权，无法调用独立 reviewer agent。

## Findings

- 无新的未收口 findings。

## Material findings

- 无。

## Senior implementation assessment

- `ChallengeWorkspaceShell.vue`、`ChallengeSolutionsPanel.vue`、`ChallengeSubmissionRecordsPanel.vue` 已整体迁入 `features/challenge-detail/ui`，`ChallengeDetail.vue` 也改为通过 `@/features/challenge-detail` public API 组合 workspace shell。
- `architectureAllowlist.ts` 中 `challenge-detail` 对应的 3 条 `componentFeatureImportAllowlist` 已移除，说明这次不只是移动文件，而是把 legacy component 对 feature 的显式结构例外一并收掉。
- raw-source 护栏、`components.d.ts` 与两处当前事实源文档都已同步更新，避免后续测试和文档继续把 owner 指回旧 `components/challenge/*` 路径。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/challenges/__tests__/ChallengeDetail.test.ts src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts src/views/challenges/__tests__/challengeDetailSharedShell.test.ts src/views/__tests__/pageTabsStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这次只收 `challenge-detail` 的三件 feature 私有 UI；`ChallengeQuestionPanel.vue`、`ChallengeWriteupPanel.vue`、`ChallengeActionAside.vue` 仍留在 `components/challenge/*`，后续若确认只服务单一 feature，仍需要继续判断 owner。
- 独立 reviewer gate 未满足；当前文档只记录 same-context review 和已执行验证结果。

## Touched known-debt status

- `challenge-detail` 这组三件套已不再占用 `componentFeatureImportAllowlist`，P1 范围内“单一 feature 私有 UI 落位错误”这条债务继续缩小。
