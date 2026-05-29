# Class Workspace Section Router Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-class-workspace-section-router-owner-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/class-workspace-section-router-owner-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-class-workspace-section-router-owner-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-class-workspace-section-router-owner-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/class-workspace-redirect/model/useClassWorkspaceSection.ts`
  - `code/frontend/src/views/platform/PlatformClassWorkspaceSection.vue`
  - `code/frontend/src/views/teacher/TeacherClassWorkspaceSection.vue`
  - `code/frontend/src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts`
- Classification check：预期按单一 feature 下游 router helper cleanup 处理；`useClassWorkspaceSection.ts` 不是合理的 route-aware owner。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `useClassWorkspaceSection.ts` 应该只保留 alias route 到 canonical workspace target 的解析 contract，不应继续 import `vue-router`。
- `PlatformClassWorkspaceSection.vue` 与 `TeacherClassWorkspaceSection.vue` 仍应保持薄 route shell，不应为了收掉 feature allowlist 反向长出新的 navigation owner。
- `useClassStudentsPage.ts` 已经是 class students workspace 的既有 page owner，由它承接 alias redirect 更符合当前 guardrail。
- 该条 allowlist 如果继续保留，会继续冻结“feature helper 可以直接导航”的例外；本轮应当在 touched surface 内收掉。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/class-workspace-section-router-owner-cleanup.md docs/plan/impl-plan/2026-05-29-class-workspace-section-router-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-class-workspace-section-router-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/class-workspace-redirect/model/useClassWorkspaceSection.ts code/frontend/src/views/platform/PlatformClassWorkspaceSection.vue code/frontend/src/views/teacher/TeacherClassWorkspaceSection.vue code/frontend/src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `featureRouterImportAllowlist` 还有多条条目，需要后续继续按 page owner / non-page owner 逐条判定。
