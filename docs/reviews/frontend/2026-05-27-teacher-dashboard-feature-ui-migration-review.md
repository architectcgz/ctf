# Teacher Dashboard Feature UI Migration 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-teacher-dashboard-feature-ui-migration-implementation-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/teacher-dashboard-feature-ui-migration.md`
    - `docs/plan/impl-plan/2026-05-27-teacher-dashboard-feature-ui-migration-implementation-plan.md`
    - `code/frontend/src/features/teacher-dashboard/index.ts`
    - `code/frontend/src/features/teacher-dashboard/ui/*`
    - `code/frontend/src/views/teacher/TeacherDashboard.vue`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts`
    - `code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts`
    - `code/frontend/src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts`
    - `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
    - `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 allowlist 驱动的前端结构债收口，继续优先处理低风险 `feature-owned UI` 候选。
- Gate verdict：Pass（本次为同上下文复核；当前回合未使用独立 subagent review）

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- `TeacherDashboardPage.vue` 已从 `components/teacher/dashboard/` 收口到 `features/teacher-dashboard/ui/`，route view 不再依赖 legacy component page。
- `views/teacher/TeacherDashboard.vue` 现在直接从 `features/teacher-dashboard` public API 组合 `useDashboardPage()` 与 `TeacherDashboardPage`，继续保持 route page 薄壳。
- 教师总览的 portrait / insight / trend / review / intervention 子面板仍保留在 `components/teacher/dashboard/` 作为稳定分区，本轮没有把 page shell owner 迁移和更深层 teacher 页面重组混在一起。
- `architectureAllowlist.ts` 已移除教师总览对应的一条 `componentFeatureImportAllowlist` 和一条 `legacyComponentPageAllowlist`，相关 raw-source 测试与 `components.d.ts` 也已对齐到新路径。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherDashboard.test.ts src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/teacher-dashboard-feature-ui-migration.md docs/plan/impl-plan/2026-05-27-teacher-dashboard-feature-ui-migration-implementation-plan.md docs/reviews/frontend/2026-05-27-teacher-dashboard-feature-ui-migration-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/features/teacher-dashboard code/frontend/src/views/teacher/TeacherDashboard.vue code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/components.d.ts code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts code/frontend/src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts code/frontend/src/views/__tests__/workspaceShellStyles.test.ts code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 本轮只迁教师总览 page shell，不处理其它 teacher dashboard / class 页面。
- 子分区仍在 `components/teacher/dashboard/`，后续若要继续收口，需要单独判断是否值得。

## Touched known-debt status

- 本轮 touched 的已知结构债是“应属于单一 feature 的 page-sized UI 仍滞留在 `components/**`，并依赖 allowlist 才能存活”。
- 该债务在教师总览这组 touched surface 上已完成收口：page shell 已迁到 `features/teacher-dashboard/ui`，对应 component->feature 例外和 legacy page 例外已移除，route view 与 raw-source guardrail 已同步到新边界。
