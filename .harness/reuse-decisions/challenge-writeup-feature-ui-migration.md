# Reuse Decision

## Change type
frontend architecture / feature ui / migration

## Existing code searched
- code/frontend/src/components/platform/writeup/ChallengeWriteupManagePanel.vue
- code/frontend/src/components/platform/writeup/ChallengeWriteupEditorPage.vue
- code/frontend/src/components/platform/writeup/ChallengeWriteupViewPage.vue
- code/frontend/src/components/platform/challenge/AdminChallengeWorkspaceTabs.vue
- code/frontend/src/views/platform/ChallengeWriteup.vue
- code/frontend/src/views/platform/ChallengeWriteupView.vue
- code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupEditorPage.ts
- code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupManagement.ts
- code/frontend/src/features/platform-challenge-detail/ui/index.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts
- code/frontend/src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts
- docs/architecture/frontend/06-components.md
- docs/architecture/frontend/07-pages-dataflow.md

## Similar implementations found
- `features/platform-challenge-detail/ui/*` 已经证明“只服务单个 feature、直接消费同 feature model contract 的 UI”可以收进 `features/*/ui`，而不是继续挂在 `components/*`。
- `ChallengeWriteupManagePanel.vue`、`ChallengeWriteupEditorPage.vue`、`ChallengeWriteupViewPage.vue` 当前都直接依赖 `features/challenge-writeup-editor`，说明它们早已不是中立业务组件，而是题解 feature 自己的 UI 面。
- `views/platform/ChallengeWriteup.vue` 与 `views/platform/ChallengeWriteupView.vue` 目前已经足够薄，只负责 route page 组合和 feature page model，不需要再保留一层 `components/platform/writeup/*` 作为中间桥。

## Decision
refactor_existing

## Reason
这次不是新增题解能力，而是收口 allowlist 驱动的边界例外。最小正确改动是把题解管理三件套迁到 `features/challenge-writeup-editor/ui`，并把“feature-owned UI” 的判定规则写回前端架构事实源；这样 route view、feature model、feature ui、common components 的边界就能在一个真实切片里闭环，而不是继续靠 `components/*Page.vue -> @/features/*` 例外维持现状。

## Files to modify
- .harness/reuse-decisions/challenge-writeup-feature-ui-migration.md
- docs/plan/impl-plan/2026-05-27-challenge-writeup-feature-ui-migration-implementation-plan.md
- docs/architecture/frontend/06-components.md
- docs/architecture/frontend/07-pages-dataflow.md
- code/frontend/src/features/challenge-writeup-editor/index.ts
- code/frontend/src/features/challenge-writeup-editor/ui/*
- code/frontend/src/features/platform-challenge-detail/model/usePlatformChallengeDetailPage.ts
- code/frontend/src/views/platform/ChallengeWriteup.vue
- code/frontend/src/views/platform/ChallengeWriteupView.vue
- code/frontend/src/views/platform/ChallengeDetail.vue
- code/frontend/src/components/platform/challenge/AdminChallengeWorkspaceTabs.vue
- code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue
- code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components.d.ts
- code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts
- code/frontend/src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts
- code/frontend/src/views/platform/__tests__/ChallengeDetail.test.ts
- code/frontend/src/views/platform/__tests__/challengeDetailWorkspaceExtraction.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- 如果这次形成了稳定的 `features/*/ui` 收口模式，后续继续沿同样方式处理其他直接依赖 feature model 的 legacy component page / panel。
- 如果只是题解这组本地收口，不额外登记长期 reuse 索引。
