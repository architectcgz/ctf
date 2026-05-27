# Reuse Decision

## Change type
component / page / feature

## Existing code searched
- code/frontend/src/components/platform/awd-service/AWDChallengeLibraryPage.vue
- code/frontend/src/views/platform/AWDChallengeLibrary.vue
- code/frontend/src/views/platform/AWDChallengeImport.vue
- code/frontend/src/features/platform-awd-challenges/model/useAwdChallengeLibraryPage.ts
- code/frontend/src/features/platform-awd-challenges/model/useAwdChallengeImportPage.ts
- code/frontend/src/features/platform-awd-challenges/model/usePlatformAwdChallenges.ts
- code/frontend/src/components/platform/dashboard/PlatformOverviewPage.vue
- code/frontend/src/components/platform/dashboard/PlatformOverviewHeroPanel.vue
- code/frontend/src/components/platform/dashboard/PlatformOverviewAlertsSection.vue
- code/frontend/src/components/platform/dashboard/PlatformOverviewHotspotsSection.vue
- code/frontend/src/components/platform/user/UserGovernancePage.vue
- code/frontend/src/components/platform/user/UserGovernanceOverviewPanel.vue
- code/frontend/src/components/platform/user/UserGovernanceImportPanel.vue
- code/frontend/src/views/platform/__tests__/AWDChallengeLibrary.test.ts
- code/frontend/src/views/__tests__/workspaceShellStyles.test.ts

## Similar implementations found
- `PlatformOverviewPage.vue` 已经按 hero / alerts / hotspots 拆成稳定展示分区，父页只保留 page shell 和数据装配，这和 `AWDChallengeLibraryPage.vue` 当前的“页头 + library 模式 + import 模式”结构最接近。
- `UserGovernancePage.vue` 已经把 overview 与 import 面板拆成独立子组件，同时保留父页对 `panel` 和选中态的 owner，这说明 `AWDChallengeLibraryPage.vue` 也应采用“父页保留 mode 和顶层事件桥接，子组件收展示分区”的模式。
- 现有 `WorkspaceDirectoryToolbar`、`WorkspaceDataTable`、`WorkspaceDirectoryPagination`、`AppEmpty`、`AppLoading` 已经覆盖目录页和空/加载状态，不需要为 AWD 题目页再发明新的列表基础设施。

## Decision
refactor_existing

## Reason
这次不是新增 AWD 题目库能力，而是收口一个已经由 route view 和 feature composable 拥有真实 page owner 的 legacy component page。最小正确改动是沿用现有 workspace page 分区拆分模式，把 `AWDChallengeLibraryPage.vue` 内部的稳定展示区块拆成子组件，同时继续复用现成的目录组件和 feature composable，而不是再新增一层新的 page model、widget 框架或重复基础表格能力。

## Files to modify
- .harness/reuse-decisions/awd-challenge-library-page-decomposition.md
- docs/plan/impl-plan/2026-05-27-awd-challenge-library-page-decomposition-implementation-plan.md
- code/frontend/src/components/platform/awd-service/AWDChallengeLibraryPage.vue
- code/frontend/src/components/platform/awd-service/AwdChallengeWorkspaceHeader.vue
- code/frontend/src/components/platform/awd-service/AwdChallengeLibrarySection.vue
- code/frontend/src/components/platform/awd-service/AwdChallengeImportSection.vue
- code/frontend/src/views/platform/AWDChallengeLibrary.vue
- code/frontend/src/views/platform/AWDChallengeImport.vue
- code/frontend/src/views/platform/__tests__/AWDChallengeLibrary.test.ts
- code/frontend/src/views/__tests__/workspaceShellStyles.test.ts

## After implementation
- 如果这次拆分形成了稳定的“共享 workspace page 同时承载 library/import 双 mode”的模式，再考虑把线索补到本地 `.harness/reuse-index/`。
- 如果只是当前 AWD 题目页的局部 owner 收口，不额外登记长期 reuse 索引。
