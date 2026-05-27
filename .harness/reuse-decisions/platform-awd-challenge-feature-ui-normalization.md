# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/features/platform-awd-challenges/ui/AWDChallengeLibraryPage.vue
- code/frontend/src/views/platform/AWDChallengeLibrary.vue
- code/frontend/src/components/platform/awd-service/AWDChallengeEditorDialog.vue
- code/frontend/src/components/platform/awd-service/AwdChallengeImportSection.vue
- code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeEditorDialog.test.ts
- code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- docs/architecture/frontend/06-components.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `features/challenge-writeup-editor/ui/*` 已经承接“单一 feature 的 manage / editor / view surface”，不再让对应 page-sized UI 继续挂在 `components/**`。
- `features/platform-awd-challenges/ui/AWDChallengeLibraryPage.vue` 已经是 `platform-awd-challenges` 的 page shell，但其下游编辑对话框与导入区块仍留在 `components/platform/awd-service`。
- 现有架构文档已经把“只服务单一 feature 的 workspace、editor、panel 或 page-sized surface 应进 `features/*/ui/`”定为事实源，不需要再创造新的落点规则。

## Decision
refactor_existing

## Reason
这组 `AWDChallengeEditorDialog.vue` / `AwdChallengeImportSection.vue` 只被 `platform-awd-challenges` feature 消费，并且直接依赖该 feature 暴露的 contract。继续把它们留在 `components/platform/awd-service` 只会让 `componentFeatureImportAllowlist` 继续冻结历史路径。最小正确改动是把这两块 UI 迁入 `features/platform-awd-challenges/ui`，由 feature public API 暴露，保留现有 page model、导入流程和保存行为 owner 不变。

## Files to modify
- .harness/reuse-decisions/platform-awd-challenge-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-27-platform-awd-challenge-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-27-platform-awd-challenge-feature-ui-normalization-review.md
- code/frontend/src/features/platform-awd-challenges/index.ts
- code/frontend/src/features/platform-awd-challenges/ui/index.ts
- code/frontend/src/features/platform-awd-challenges/ui/AWDChallengeLibraryPage.vue
- code/frontend/src/features/platform-awd-challenges/ui/AWDChallengeEditorDialog.vue
- code/frontend/src/features/platform-awd-challenges/ui/AwdChallengeImportSection.vue
- code/frontend/src/views/platform/AWDChallengeLibrary.vue
- code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeEditorDialog.test.ts
- code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts
- code/frontend/src/components.d.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts
- code/frontend/src/views/__tests__/workspaceShellStyles.test.ts
- code/frontend/src/views/__tests__/studentDirectoryTypographyBoundary.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `platform-awd-challenges` 这组 feature 会同时持有 page shell、导入区块和编辑对话框的 UI owner。
- `AWDChallengeLibrary.vue` route view 会继续只组合 feature public API，而不再直接引用旧 `components/platform/awd-service` 路径。
- `componentFeatureImportAllowlist` 中这两条 `platform-awd-challenges` 历史例外可以移除。
