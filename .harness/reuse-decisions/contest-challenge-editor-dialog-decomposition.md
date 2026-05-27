# Reuse Decision

## Change type
component / dialog / feature

## Existing code searched
- code/frontend/src/components/platform/contest/ContestChallengeEditorDialog.vue
- code/frontend/src/components/platform/contest/ContestChallengeOrchestrationPanel.vue
- code/frontend/src/components/platform/awd-service/AWDChallengeEditorDialog.vue
- code/frontend/src/components/platform/awd-service/AwdChallengeLibrarySection.vue
- code/frontend/src/components/platform/user/UserGovernanceOverviewPanel.vue
- code/frontend/src/components/platform/challenge/ChallengeManageDirectoryPanel.vue
- code/frontend/src/components/common/modal-templates/AdminSurfaceModal.vue
- code/frontend/src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts
- code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts
- code/frontend/src/views/__tests__/studentDirectoryTypographyBoundary.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts
- code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts

## Similar implementations found
- `AwdChallengeLibrarySection.vue` 已经证明“目录筛选 + `WorkspaceDataTable` + 分页”可以作为纯展示分区抽出去，父组件继续保留筛选值和动作 owner，这和 `ContestChallengeEditorDialog.vue` 里的 AWD 题目池选择区最接近。
- `UserGovernanceOverviewPanel.vue`、`ChallengeManageDirectoryPanel.vue` 这类目录面板都采用“父层提供筛选值和列表动作，子组件只负责渲染”的边界，说明 `ContestChallengeEditorDialog.vue` 里的 AWD 选择表格也不需要继续和 submit / validation 混在一个文件里。
- `AWDChallengeEditorDialog.vue` 维持了“父对话框拥有 local draft + validate + submit”的模式，这次对 `ContestChallengeEditorDialog.vue` 也应保留相同 owner，而不是把 form / validation 下沉到子组件里。

## Decision
refactor_existing

## Reason
这次不是新增比赛编排能力，而是收口一个同时承担 AWD 题目目录选择、普通题目选择、分值顺序设置和提交校验的超大对话框。最小正确改动是复用现有目录区块拆分模式，把稳定展示区块抽成子组件，同时继续让父对话框保留 `form`、校验、submit、选择状态和对 orchestration panel 的事件桥接，不新增新的 feature/composable owner。

## Files to modify
- .harness/reuse-decisions/contest-challenge-editor-dialog-decomposition.md
- docs/plan/impl-plan/2026-05-27-contest-challenge-editor-dialog-decomposition-implementation-plan.md
- code/frontend/src/components/platform/contest/ContestChallengeEditorDialog.vue
- code/frontend/src/components/platform/contest/ContestAwdChallengeSelectorSection.vue
- code/frontend/src/components/platform/contest/ContestChallengeSettingsSection.vue
- code/frontend/src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts
- code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts
- code/frontend/src/views/__tests__/studentDirectoryTypographyBoundary.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts
- code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts

## After implementation
- 如果这次形成了稳定的“父对话框持有 form / validation，子组件承接目录 section”的模式，再考虑补到本地 `.harness/reuse-index/`。
- 如果只是 `ContestChallengeEditorDialog.vue` 的局部收口，不额外登记长期 reuse 索引。
