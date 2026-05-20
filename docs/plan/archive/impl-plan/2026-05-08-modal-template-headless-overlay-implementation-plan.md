# Modal Template Headless Overlay Implementation Plan

## Objective

拆分 `ModalTemplateShell` 中混在一起的 overlay 行为和视觉模板职责，建立可复用的 headless overlay 行为层。标准弹窗模板继续兼容现有 API，业务专用抽屉后续可直接复用行为层而不是继承通用视觉模板。

## Non-goals

- 不重写所有现有弹窗视觉。
- 不迁移业务页面调用方。
- 不删除 `ModalTemplateShell`、`ClassicCenteredModal` 或 `SlideOverDrawer`。
- 不改变现有打开、关闭、Teleport、Escape、backdrop、body scroll lock 行为。

## Inputs

- `feedback/2026-05-08-specialized-drawer-should-not-inherit-modal-template.md`
- `code/frontend/src/components/common/modal-templates/ModalTemplateShell.vue`
- `code/frontend/src/components/common/modal-templates/ClassicCenteredModal.vue`
- `code/frontend/src/components/common/modal-templates/SlideOverDrawer.vue`
- `code/frontend/src/components/common/__tests__/ModalTemplates.test.ts`

## Task Slices

1. 新增 headless overlay 行为 composable
   - 新增 `useOverlayBehavior.ts`
   - 管理 Escape 监听、body scroll lock、backdrop close 入口
   - 保持卸载清理

2. 新增 headless `OverlayPortal.vue`
   - 负责 Teleport、Transition、根节点事件绑定
   - 不绑定任何产品视觉 class
   - 通过 slot 暴露 close

3. 收敛 `ModalTemplateShell.vue`
   - 改为调用 `OverlayPortal`
   - 保留现有 props、emits、class 结构，避免破坏调用方
   - 保留当前 `.modal-template-shell` / `.modal-template-panel` 视觉兼容层

4. 更新测试
   - 检查新 headless 文件存在
   - 检查 `ModalTemplateShell` 不再自己维护 window/body 副作用
   - 检查现有 modal/drawer 行为测试仍通过

## Expected Files

- `code/frontend/src/components/common/modal-templates/useOverlayBehavior.ts`
- `code/frontend/src/components/common/modal-templates/OverlayPortal.vue`
- `code/frontend/src/components/common/modal-templates/ModalTemplateShell.vue`
- `code/frontend/src/components/common/__tests__/ModalTemplates.test.ts`

## Compatibility Impact

现有调用方仍可继续使用 `ModalTemplateShell`、`ClassicCenteredModal` 和 `SlideOverDrawer`。本次只改变内部职责拆分，不改变公开 props 语义。

## Validation

- `npm run test:run -- src/components/common/__tests__/ModalTemplates.test.ts`
- `npm run typecheck`
- `npm run check:theme-tail`
- `npx prettier --check src/components/common/modal-templates/useOverlayBehavior.ts src/components/common/modal-templates/OverlayPortal.vue src/components/common/modal-templates/ModalTemplateShell.vue src/components/common/__tests__/ModalTemplates.test.ts`
- `npx eslint --quiet src/components/common/modal-templates/useOverlayBehavior.ts src/components/common/modal-templates/OverlayPortal.vue src/components/common/modal-templates/ModalTemplateShell.vue src/components/common/__tests__/ModalTemplates.test.ts`

## Review Focus

- `ModalTemplateShell` 是否仍把视觉补丁作为主要扩展方式。
- headless 行为层是否足够独立，业务组件能否直接复用。
- Escape/backdrop/body scroll lock 是否保持兼容。
- 是否引入 Teleport 样式作用域问题。

## Rollback

如果出现兼容问题，可删除新增 headless 文件，并将 `ModalTemplateShell.vue` 恢复为原先直接管理 Teleport 和副作用的实现。
