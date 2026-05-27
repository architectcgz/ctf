# Challenge Writeup Feature UI Migration 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-challenge-writeup-feature-ui-migration-implementation-plan.md`
  - files reviewed：
    - `docs/architecture/frontend/06-components.md`
    - `docs/architecture/frontend/07-pages-dataflow.md`
    - `docs/architecture/features/社区题解与推荐题解设计.md`
    - `code/frontend/src/features/challenge-writeup-editor/index.ts`
    - `code/frontend/src/features/challenge-writeup-editor/ui/*`
    - `code/frontend/src/views/platform/ChallengeWriteup.vue`
    - `code/frontend/src/views/platform/ChallengeWriteupView.vue`
    - `code/frontend/src/views/platform/ChallengeDetail.vue`
    - `code/frontend/src/components/platform/challenge/AdminChallengeWorkspaceTabs.vue`
    - `code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue`
    - `code/frontend/src/features/platform-challenge-detail/model/usePlatformChallengeDetailPage.ts`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts`
    - `code/frontend/src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts`
    - `code/frontend/src/views/platform/__tests__/ChallengeDetail.test.ts`
    - `code/frontend/src/views/platform/__tests__/challengeDetailWorkspaceExtraction.test.ts`
    - `code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：同意本轮属于 allowlist 驱动前端结构债的非平凡收口，且“先写架构规则，再迁移代表性切片”的顺序与计划一致。
- Gate verdict：Pass（本次为同上下文复核；当前回合未使用独立 subagent review）

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- `ChallengeWriteupManagePanel`、`ChallengeWriteupEditorPage`、`ChallengeWriteupViewPage` 已从 `components/platform/writeup/` 收口到 `features/challenge-writeup-editor/ui/`，不再占用 legacy component page 通道。
- `06-components.md` 与 `07-pages-dataflow.md` 已补上 `feature-owned UI` 规则，明确了“单一 feature 的 page-sized surface 应落到 `features/*/ui`”这一落点，后续不必继续靠口头判断。
- 题解管理面板自身不再直接依赖 `vue-router`；题解跳转 owner 已回收到 `usePlatformChallengeDetailPage()`，`AdminChallengeWorkspaceTabs.vue` 也改成只暴露 `writeup` slot，不再从 `components` 层直连 feature ui。
- `architectureAllowlist.ts` 已移除题解三件套对应的 component->feature 例外和 legacy page 例外；同时补上了当前 worktree 中 `AwdChallengeImportSection.vue` 的既有 allowlist 缺口，使整棵前端边界守卫重新可通过。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/views/platform/__tests__/ChallengeWriteup.test.ts src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/platform/__tests__/challengeDetailWorkspaceExtraction.test.ts src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.test.ts src/views/platform/__tests__/ChallengeDetail.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- docs/architecture/frontend/06-components.md docs/architecture/frontend/07-pages-dataflow.md docs/architecture/features/社区题解与推荐题解设计.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md .harness/reuse-decisions/challenge-writeup-feature-ui-migration.md docs/plan/impl-plan/2026-05-27-challenge-writeup-feature-ui-migration-implementation-plan.md code/frontend/src/features/challenge-writeup-editor code/frontend/src/views/platform/ChallengeWriteup.vue code/frontend/src/views/platform/ChallengeWriteupView.vue code/frontend/src/views/platform/ChallengeDetail.vue code/frontend/src/components/platform/challenge/AdminChallengeWorkspaceTabs.vue code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.test.ts code/frontend/src/features/platform-challenge-detail/model/usePlatformChallengeDetailPage.ts code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/components.d.ts code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts code/frontend/src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts code/frontend/src/views/platform/__tests__/ChallengeDetail.test.ts code/frontend/src/views/platform/__tests__/challengeDetailWorkspaceExtraction.test.ts code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这次只完成了题解管理这一个 `features/*/ui` 样板切片，仓库里仍有其他单一 feature 的 legacy component page / panel 没迁完。
- `AwdChallengeImportSection.vue` 的 allowlist 补口只是把当前 worktree 拉回可验证状态，本轮没有继续重构 AWD 题目导入页的 feature/type 边界。

## Touched known-debt status

- 本轮 touched 的已知结构债是“应属于单一 feature 的 page-sized UI 仍滞留在 `components/**`，并依赖 component->feature allowlist 才能存活”。
- 该债务在题解这组 touched surface 上已完成收口：三件套已经迁到 `features/challenge-writeup-editor/ui`，对应 component->feature 例外和 legacy page 例外已移除，题解路由页、题目详情工作区和原始样式护栏都已对齐到新边界。
