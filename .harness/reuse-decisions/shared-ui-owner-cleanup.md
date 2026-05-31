# Reuse Decision

## Change type
component / composition / layout / docs / test

## Existing code searched
- `code/frontend/src/components/common/*`
- `code/frontend/src/components/common/modal-templates/*`
- `code/frontend/src/components/layout/*`
- `code/frontend/src/components/layout/sidebar/*`
- `code/frontend/src/components/layout/topnav/*`
- `code/frontend/src/components/layout/notification-drawer/*`
- `code/frontend/src/components/navigation/*`
- `code/frontend/src/components/charts/*`
- `code/frontend/src/components/errors/*`
- `code/frontend/src/widgets/layout-shell/*`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `code/frontend/src/__tests__/frontendArchitecturePolicy.ts`
- `code/frontend/scripts/frontend-architecture-policy.json`
- `code/frontend/vite.config.ts`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/06-components.md`

## Similar implementations found
- `features/*/ui` 与 `features/*/model` 已经在前几轮 owner cleanup 中承接“可渲染表面”和“状态 / 流程 owner”的分层。
- `widgets/layout-shell/model/*` 已经把 layout 相关 bridge 独立成 model 文件，说明 layout 运行时桥接本身适合从纯 UI 壳里拆开。
- `entities/training-timeline/ui` 的迁移已经表明：目录移动不能只改组件路径，还要同步 raw-source 测试、架构守卫和自动生成类型入口。

## Decision
refactor_existing

## Reason
- `components/common`、`components/layout`、`components/navigation`、`components/charts`、`components/errors` 已经不再是“历史业务组件目录”，而是长期共享层；继续挂在 `components/*` 会把“共享 UI / 共享 helper / 共享契约”混在一起。
- 这些文件不应该整体硬塞进 `shared/ui`。可渲染组件和样式壳进入 `shared/ui`；纯类型、导航契约、overlay 行为、chart helper、layout bridge / view-state 等非 UI 文件应分别进入 `shared/model` 或 `shared/lib`。
- `components/layout/*` 当前还依赖 `widgets/layout-shell/*` bridge；如果只搬目录、不一起收口这条反向依赖，会形成新的 `shared -> widgets` 非法依赖。

## Files to modify
- `.harness/reuse-decisions/shared-ui-owner-cleanup.md`
- `docs/plan/impl-plan/2026-05-30-shared-ui-owner-cleanup-plan.md`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/06-components.md`
- `code/frontend/scripts/frontend-architecture-policy.json`
- `code/frontend/src/__tests__/frontendArchitecturePolicy.ts`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `code/frontend/vite.config.ts`
- `code/frontend/src/components.d.ts`
- `code/frontend/src/shared/**`
- `code/frontend/src/shared/ui/common/DeleteConfirmModal.vue`
- `code/frontend/src/shared/ui/common/WorkspaceDataTable.vue`
- `code/frontend/src/shared/ui/common/__tests__/WorkspaceDataTable.test.ts`
- `code/frontend/src/shared/ui/common/modal-templates/AdminSurfaceDrawer.vue`
- `code/frontend/src/shared/ui/common/modal-templates/AdminSurfaceModal.vue`
- `code/frontend/src/shared/ui/common/modal-templates/ClassicCenteredModal.vue`
- `code/frontend/src/shared/ui/common/modal-templates/MinimalFloatingModal.vue`
- `code/frontend/src/shared/ui/common/modal-templates/ModalTemplateShell.vue`
- `code/frontend/src/shared/ui/common/modal-templates/OverlayPortal.vue`
- `code/frontend/src/shared/ui/common/modal-templates/SlideOverDrawer.vue`
- `code/frontend/src/shared/ui/layout/AppLayout.vue`
- `code/frontend/src/shared/ui/layout/NotificationDrawer.vue`
- `code/frontend/src/shared/ui/layout/notification-drawer/NotificationDrawerBody.vue`
- `code/frontend/src/shared/ui/layout/notification-drawer/NotificationDrawerFooter.vue`
- `code/frontend/src/shared/ui/layout/notification-drawer/NotificationDrawerHeader.vue`
- `code/frontend/src/shared/ui/layout/notification-drawer/NotificationDrawerSummary.vue`
- `code/frontend/src/shared/ui/layout/notification-drawer/NotificationDrawerTabs.vue`
- `code/frontend/src/widgets/layout-shell/**`
- `code/frontend/src/pages/**`
- `code/frontend/src/features/**`
- `code/frontend/src/entities/**`
- `code/frontend/src/widgets/**`
- `code/frontend/src/components/common/__tests__/*`
- `code/frontend/src/components/layout/__tests__/*`
- `code/frontend/src/components/charts/__tests__/*`
- `code/frontend/src/pages/**/__tests__/*`
- `code/frontend/src/__tests__/*`

## After implementation
- `components/common`、`components/layout`、`components/navigation`、`components/charts`、`components/errors` 的长期共享职责将迁入 `shared/*`。
- `shared/ui` 只保留可渲染组件与样式壳；非 UI helper / contract / type / composition 分别落到 `shared/model` 或 `shared/lib`。
- layout 相关 bridge 不再通过 `widgets/layout-shell` 反向挂在共享布局壳上。
