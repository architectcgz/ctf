## 目标

- 重构 `SlideOverDrawer` 的头部结构，让图标、标题区、关闭按钮的布局由共享壳层显式表达。
- 移除 `NotificationDropdown` 中依赖错误层级和 `!important` 的头部补丁样式。
- 保留通知中心当前的统计、筛选、列表和底部动作设计，并继续使用项目已有 token / 变量体系。

## 非目标

- 不改通知数据流、筛选逻辑、路由跳转或标记已读行为。
- 不重做 `TopNav`、`AppLayout` 或通知中心页面 `NotificationList`。
- 不顺手调整与本次无关的其他抽屉视觉风格。

## 输入资料

- `docs/architecture/frontend/pages/右侧抽屉弹窗/a.vue`
- `code/frontend/src/components/common/modal-templates/SlideOverDrawer.vue`
- `code/frontend/src/components/layout/NotificationDropdown.vue`
- `code/frontend/src/components/common/__tests__/ModalTemplates.test.ts`
- `code/frontend/src/components/layout/__tests__/NotificationDropdown.test.ts`

## 结构判断

- 当前问题不在组件挂载链路，真实渲染入口是 `AppLayout -> TopNav -> NotificationDropdown`。
- 当前偏差在共享壳层结构：关闭按钮是 `header` 的兄弟节点，而不是 `head-row` 的成员。
- 因此通知中心局部覆盖无法优雅地修正头部排布，只能通过共享壳层重构收口。

## 任务切片

### 切片 1：共享抽屉头部重构

- 目标：让 `SlideOverDrawer` 头部具备稳定的左右轨布局，支持图标/标题主区与关闭按钮操作区并列。
- 变更面：
  - `code/frontend/src/components/common/modal-templates/SlideOverDrawer.vue`
  - `code/frontend/src/components/common/__tests__/ModalTemplates.test.ts`
- 方案：
  - 把关闭按钮移入头部顶层布局容器。
  - 通过共享 CSS 变量控制头部列间距、关闭按钮对齐、标题区宽度，而不是局部绝对定位补丁。
  - 保持现有外部 API 不变，避免影响现有抽屉调用点。
- 验证：
  - `npm run test:run -- src/components/common/__tests__/ModalTemplates.test.ts`
- review 关注点：
  - 共享壳层结构是否清晰。
  - 现有调用方是否仍能复用，不需要额外 props 迁移。

### 切片 2：通知中心样式收口

- 目标：删除通知中心针对共享壳层错误层级的修补规则，只保留必要的主题变量和局部布局样式。
- 变更面：
  - `code/frontend/src/components/layout/NotificationDropdown.vue`
  - `code/frontend/src/components/layout/__tests__/NotificationDropdown.test.ts`
- 方案：
  - 移除 `!important` 与错误层级覆盖。
  - 基于新头部结构调整通知中心头部统计、按钮和 pills 的局部间距。
  - 保持按钮轮廓增强与筛选可读性优化。
- 验证：
  - `npm run test:run -- src/components/layout/__tests__/NotificationDropdown.test.ts`
- review 关注点：
  - 通知中心是否重新依赖脆弱选择器。
  - 是否仍然满足新的参考稿视觉目标。

## 预期修改文件

- `code/frontend/src/components/common/modal-templates/SlideOverDrawer.vue`
- `code/frontend/src/components/layout/NotificationDropdown.vue`
- `code/frontend/src/components/common/__tests__/ModalTemplates.test.ts`
- `code/frontend/src/components/layout/__tests__/NotificationDropdown.test.ts`

## 验证计划

1. `npm run test:run -- src/components/common/__tests__/ModalTemplates.test.ts src/components/layout/__tests__/NotificationDropdown.test.ts`
2. `npm run typecheck`

## 回退说明

- 共享壳层改动若引入回归，可回退本次分支的 `SlideOverDrawer.vue` 和对应测试提交。
- 通知中心局部样式仍保持独立，可单独回退 `NotificationDropdown.vue` 的收口修改。
