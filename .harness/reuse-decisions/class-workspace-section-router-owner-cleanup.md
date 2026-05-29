# Reuse Decision

## Change type
frontend refactor / feature router helper cleanup

## Existing code searched
- code/frontend/src/features/class-workspace-redirect/model/useClassWorkspaceSection.ts
- code/frontend/src/views/platform/PlatformClassWorkspaceSection.vue
- code/frontend/src/views/teacher/TeacherClassWorkspaceSection.vue
- code/frontend/src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `useStudentAnalysisNavigation.ts` 已改为消费 callback contract，不再直接依赖 `vue-router`。
- `useNotificationDrawer.ts` 已改为消费导航 callback，具体路由动作回到 layout shell bridge owner。
- `useClassWorkspaceSection.ts` 当前不是 page owner，而是 alias route 到 canonical workspace 的解析 helper，更接近上述“helper 不直接拿 router”的模式。

## Decision
refactor_existing

## Reason
`features/class-workspace-redirect/model/useClassWorkspaceSection.ts -> vue-router` 不属于合理长期例外。它当前只负责把 legacy class workspace alias route 映射回 canonical workspace `panel` query，并不应该同时持有 `useRoute()`、`useRouter()` 和 `router.replace()`。

最小正确改动是：

- 让 `useClassWorkspaceSection()` 只消费本地 route-like contract，产出 canonical workspace target
- 让既有 page owner `useClassStudentsPage.ts` 承接 alias redirect 的 `router.replace()`，避免把 route owner 漂回 view
- 删除对应 allowlist 条目，并补 raw-source 护栏

本轮不做：

- 不重构 `PlatformClassStudents.vue` / `TeacherClassStudents.vue` 的工作区逻辑
- 不调整 class workspace 的 panel 语义、query 结构或 teacher / platform route 命名
- 不处理 `featureRouterImportAllowlist` 其它条目

## Files to modify
- .harness/reuse-decisions/class-workspace-section-router-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-class-workspace-section-router-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-class-workspace-section-router-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/class-workspace-redirect/model/useClassWorkspaceSection.ts
- code/frontend/src/views/platform/PlatformClassWorkspaceSection.vue
- code/frontend/src/views/teacher/TeacherClassWorkspaceSection.vue
- code/frontend/src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts

## After implementation
- `useClassWorkspaceSection.ts` 不再 import `vue-router`
- class workspace alias route 的 canonical redirect target 由 helper 计算，但真正的 `router.replace()` 回到 `useClassStudentsPage.ts`
- `featureRouterImportAllowlist` 再减少一条
