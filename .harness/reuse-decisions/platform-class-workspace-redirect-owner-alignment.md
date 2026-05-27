# Reuse Decision

## Change type
+feature / page / test / docs

## Existing code searched
- `code/frontend/src/features/class-workspace-redirect/model/useClassWorkspaceSection.ts`
- `code/frontend/src/views/platform/PlatformClassWorkspaceSection.vue`
- `code/frontend/src/views/teacher/TeacherClassWorkspaceSection.vue`
- `code/frontend/src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts`
- `code/frontend/src/utils/teachingWorkspaceRouting.ts`
- `code/frontend/src/router/__tests__/sharedRoutes.test.ts`
- `docs/plan/impl-plan/2026-05-24-platform-route-owner-decoupling-implementation-plan.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-admin-awd-review-api-owner-alignment-review.md`

## Similar implementations found
- `teachingWorkspaceRouting.ts` 已经承载 platform / teacher 共享教学工作台的 role-aware canonical route 解析，说明 route 名选择应收口到独立 routing util，而不是散在具体 page workflow 里。
- 当前 `class-workspace-redirect` 已经是 teacher / platform 共用的 redirect owner，最小改动是继续复用这个 feature，把“panel 解析”和“canonical route owner”拆开，而不是恢复 platform / teacher 两套平行 redirect feature。

## Decision
refactor_existing

## Reason
- 当前 `useClassWorkspaceSection` 同时硬编码 teacher / platform 的 alias route 名和 canonical target route 名，虽然功能正确，但 feature 内部仍混着两套命名空间的 target owner，成为 backlog 里剩余的 redirect 命名残留。
- 最小正确方案不是新建更多 role-specific feature，而是继续复用现有 `class-workspace-redirect`，把它收口为“只负责从 legacy route 解析 panel + 接收调用方给出的 canonical route name”。
- 这样平台 route view 只显式声明自己的 canonical target，teacher route view 只显式声明 teacher target；共享 feature 不再自己决定 teacher/platform 最终落点。

## Files to modify
- `.harness/reuse-decisions/platform-class-workspace-redirect-owner-alignment.md`
- `docs/plan/impl-plan/2026-05-27-platform-class-workspace-redirect-owner-alignment-implementation-plan.md`
- `code/frontend/src/features/class-workspace-redirect/model/useClassWorkspaceSection.ts`
- `code/frontend/src/views/platform/PlatformClassWorkspaceSection.vue`
- `code/frontend/src/views/teacher/TeacherClassWorkspaceSection.vue`
- `code/frontend/src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-platform-class-workspace-redirect-owner-alignment-review.md`

## After implementation
- 如果这类 alias route owner 收口模式稳定，后续其他共享 redirect feature 也应保持“共享 feature 只解析状态，最终 canonical route 由 route view 或 routing util 显式注入”的边界。
