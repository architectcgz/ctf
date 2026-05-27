# Teacher Instance Feature UI Migration 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-teacher-instance-feature-ui-migration-implementation-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/teacher-instance-feature-ui-migration.md`
    - `docs/plan/impl-plan/2026-05-27-teacher-instance-feature-ui-migration-implementation-plan.md`
    - `docs/reviews/frontend/2026-05-27-teacher-instance-feature-ui-migration-review.md`
    - `code/frontend/src/features/teacher-instances/**/*`
    - `code/frontend/src/views/teacher/InstanceManagement.vue`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/components.d.ts`
    - `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
    - `code/frontend/src/views/teacher/__tests__/teacherRootShellCleanup.test.ts`
    - `code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts`
    - `code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts`
    - `code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts`
    - `code/frontend/src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts`
    - `code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts`
    - `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
    - `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
    - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 allowlist 驱动的前端结构债收口，这次只处理教师实例 page shell 迁移，不引入新的 route / API owner。
- Gate verdict：Self-check pass，独立 reviewer gate 待补

## Findings

- 未发现阻塞性问题。

## Material findings

- None.

## Senior implementation assessment

- `TeacherInstanceManagementPage.vue` 已迁入 `features/teacher-instances/ui/`，`views/teacher/InstanceManagement.vue` 改为通过 `features/teacher-instances` public API 组合 page shell 与 page model，route view 没有继续直连 legacy component path。
- `useInstanceManagementPage.ts` 和 `useInstances.ts` 继续持有初始化、筛选、防抖、销毁确认、分页和 dashboard 跳转 owner，本轮没有把这些流程重新吸回 UI。
- `legacyComponentPageAllowlist` 已移除教师实例页，对应的教师端 surface / shell / raw-source 测试路径也已切到新 owner。
- 当前记录只覆盖实现者 self-check；如果要完全满足 development-pipeline 的独立 review gate，还需要在用户显式授权 delegation 后补一轮 reviewer agent 复核。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/InstanceManagement.test.ts src/views/teacher/__tests__/teacherRootShellCleanup.test.ts src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts src/views/teacher/__tests__/teacherSurface.test.ts src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/teacher-instance-feature-ui-migration.md docs/plan/impl-plan/2026-05-27-teacher-instance-feature-ui-migration-implementation-plan.md docs/reviews/frontend/2026-05-27-teacher-instance-feature-ui-migration-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/features/teacher-instances code/frontend/src/views/teacher/InstanceManagement.vue code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/components.d.ts code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts code/frontend/src/views/teacher/__tests__/teacherRootShellCleanup.test.ts code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts code/frontend/src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts code/frontend/src/views/__tests__/workspaceShellStyles.test.ts code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `TeacherInstanceHeroPanel.vue` 与 `TeacherInstanceDirectorySection.vue` 仍保留在 `components/teacher/instance-management/`，本轮只收口 page-sized shell。

## Touched known-debt status

- `TeacherInstanceManagementPage.vue` 这条 `legacy component page` 结构债已在 touched surface 内收口。
