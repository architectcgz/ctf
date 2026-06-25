# 前端组件体系设计

> 状态：Current
> 事实源：`code/frontend/src/shared/ui/`、`code/frontend/src/shared/model/`、`code/frontend/src/pages/`、`code/frontend/src/__tests__/`、`code/frontend/src/features/**/{model,ui}/`
> 替代：无

## 本文档范围

| 覆盖 | 不覆盖 |
|------|--------|
| `shared/ui/common/` 共享展示原语和通用交互壳 | 单个页面的数据流细节（见 `07-pages-dataflow.md`） |
| `shared/ui/common/modal-templates/` Overlay 行为层、居中弹窗、侧边抽屉和后台工作区模板 | 后端接口语义（见 `docs/contracts/`） |
| `shared/ui/layout/` 应用总布局、侧栏、顶栏、全局通知实时连接、route transition | 页面内部业务查询、目录筛选或详情页状态机 |
| `shared/model/common/` 共享 UI 原语状态（Toast、危险确认、工作区导航、后台面包屑细节） | 业务数据加载、领域流程编排或页面级异步状态机 |
| `features/**/model` 页面行为 owner、请求编排、路由参数解析、分页、导出和局部状态机 | 承担大段模板和展示样式壳 |
| `features/**/ui` 只服务单一 feature 的 workspace、editor、panel 或 page-sized surface | 跨 feature 复用、直接调用非 contract API |
| `widgets/**` 跨 feature 的 page-sized workspace 组合 | 直接持有业务 API owner、导入 `pages/**` |
| `shared/model/navigation/` route-aware transport、query/tab 同步和工作区导航桥接 | 替具体页面决定业务 tab、筛选语义或 API 请求 |

## 定位

本文档只说明前端组件应该放在哪一层、共享原语有哪些、弹窗模板怎么复用，以及布局壳和共享 UI 状态的 owner。

- 覆盖：`shared/ui/common/`、`shared/ui/layout/`、`shared/ui/common/modal-templates/`、`shared/model/common/`、`shared/model/layout/` 和 route view 的边界。
- 不覆盖：单个页面的数据流细节；这些见 `07-pages-dataflow.md`。

## 当前设计

- `code/frontend/src/shared/ui/common/`
  - 负责：共享展示原语和通用交互壳，例如 `AppEmpty.vue`、`AppToast.vue`、`AppSkeleton.vue`、`WorkspaceDataTable.vue`、`DeleteConfirmModal.vue`
  - 不负责：直接发业务请求、直接依赖 store/router，或内置某个页面的专用流程

- `code/frontend/src/shared/ui/common/modal-templates/`
  - 负责：Overlay 行为层、居中弹窗、侧边抽屉和后台工作区模板，例如 `OverlayPortal.vue`、`ModalTemplateShell.vue`、`ClassicCenteredModal.vue`、`SlideOverDrawer.vue`、`AdminSurfaceModal.vue`、`AdminSurfaceDrawer.vue`
  - 不负责：把所有业务弹窗都揉成一套万能模板；业务样式明显不同的 overlay 仍应自有 DOM/CSS，只复用 headless 行为层

- `code/frontend/src/shared/ui/layout/AppLayout.vue`
  - 负责：应用总布局、侧栏、顶栏、全局通知实时连接、route transition 和 backoffice/student 的内容壳切换
  - 不负责：页面自己的业务查询、目录筛选或详情页状态机

- `code/frontend/src/shared/model/common/`、`code/frontend/src/shared/model/layout/`
  - 负责：共享 UI 原语背后的状态 owner，例如 Toast、危险确认、工作区导航和后台面包屑细节
  - 不负责：业务数据加载、领域流程编排或页面级异步状态机

- `code/frontend/src/__tests__`
  - 负责：全局前端架构守卫与跨页面稳定策略测试
  - 不负责：替代具体 page / feature / widget 的行为测试 owner

- `code/frontend/src/features/**/model`
  - 负责：页面行为 owner、请求编排、路由参数解析、分页、导出和局部状态机
  - 不负责：承担大段模板和展示样式壳

- `code/frontend/src/features/**/ui`
  - 负责：只服务单一 feature 的 workspace、editor、panel 或 page-sized surface，直接消费同 feature 的 model contract；当页面已经完全 feature-owned 时，也可以直接作为路由入口
  - 不负责：跨 feature 复用、直接调用非 contract API，或替 route view 接管路由 owner

- `code/frontend/src/features/platform/*`
  - 负责：管理员端 `/platform/*` 路由下按能力分组的 feature 命名空间，例如 `overview`、`user-management`、`class-management`、`student-management`、`instance-management`、`challenges`、`challenge-detail`、`contests`、`awd-challenges`
  - 不负责：继续维持 `platform-*` 扁平 feature 命名或 page shell 与 feature shell 双份 owner

- `code/frontend/src/widgets/**`
  - 负责：跨 feature 的 page-sized workspace 组合，例如 `contest-detail-workspace`、`awd-review-workspace`、`challenge-detail-workspace`、`notification-*` 和 `scoreboard-detail-workspace`
  - 不负责：直接持有业务 API owner、导入 `pages/**`，或把单一 feature 的内部 UI 提升成跨 feature 共享层

边界：

- `widgets/**` 承接跨 feature 的页面级工作台组合，不直接拥有 API owner
- 若组件只服务单一 feature 且直接消费同 feature model 的 UI，应落在 `features/**/ui` 而不是 `widgets/**`
- 若组件是共享原语、布局壳或跨业务复用的展示块，应落在 `shared/ui/**` 而不是 `widgets/**`
- `widgets/**` 通常由多个 `features/**/ui`、`entities/**/ui`、`shared/ui/**` 组合而成，自己不承担 API 调用和业务状态机

- `code/frontend/src/shared/model/navigation/`、`code/frontend/src/shared/model/layout/useWorkspaceShellNavigation.ts`
  - 负责：route-aware transport、query/tab 同步和工作区导航桥接，供 feature model / layout view state 组合
  - 不负责：替具体页面决定业务 tab、筛选语义或 API 请求

边界：

- `shared/model/navigation/` 承接 route-aware 能力，例如 `useRouteNavigationTransport` 把 router push / replace 抽成可注入 transport，`useRouteQueryTransport` 读写 route query，`useRouteQueryTabs` 将 route query 映射为 tab 状态，`useUrlSyncedTabs` 为 feature model 提供 URL 同步 tab 状态
- route page 不直接持有 `useRouteQueryTabs()`；`routePageArchitectureBoundary.test.ts` 会拦截普通 route page 对 route hooks 的直接使用
- navigation composables 不负责业务 tab 的名称、权限和请求；这些由 feature model 或 widget owner 决定
- `useWorkspaceShellNavigation` 为 layout shell 提供工作区导航状态桥接，例如面包屑、tab 状态、返回按钮等，但不替页面决定具体导航目标和业务语义

## 1. 组件层次

当前前端组件按下面的 owner 分层：

| 层级 | 当前位置 | 当前职责 |
| --- | --- | --- |
| 共享原语 | `shared/ui/common/` | 空状态、Toast、Skeleton、目录表格、删除确认、通用局部承载器 |
| Overlay 模板 | `shared/ui/common/modal-templates/` | Teleport、滚动锁、Escape/backdrop 关闭、经典弹窗和抽屉模板 |
| 布局壳 | `shared/ui/layout/` | `AppLayout`、`Sidebar`、`TopNav` 等全局承载 |
| Widgets | `widgets/*/` | 跨 feature 的 page-sized workspace 组合，不直接拥有 API |
| Feature UI | `features/*/ui/`、`features/platform/*/ui/` | 只服务单一 feature 的工作区、编辑器、目录面板与 page-sized surface |
| 页面入口 | `pages/**` | 运行时 route entry、页面结构组合与最外层事件桥接 |
| 全局前端守卫 | `src/__tests__` | 架构、route page 与跨页面稳定策略测试 |

判断原则：

- 能在多个页面复用且不绑业务 owner 的，进 `shared/ui/common/`
- 只解决 overlay 行为和模板骨架的，进 `shared/ui/common/modal-templates/`
- 只服务单一 feature，且直接消费同 feature model 的 UI，进 `features/*/ui/`
- 需要把多个 feature / entity / shared UI 组合成一个页面级工作台，且不会直接拥有 API owner 的，进 `widgets/*/`
- 强业务语义的展示组件默认跟随 feature UI / feature model，不回流到新的 `components/**` 历史目录
- 页面数据编排和路由交互不放进共享组件

## 2. 当前共享原语

### 2.0 共享组件清单

当前 `shared/ui/common/` 下的组件清单：

#### 2.0.1 展示原语

| 组件 | 用途 | 代码位置 |
|------|------|---------|
| `AppEmpty.vue` | 统一空状态壳和图标映射 | `shared/ui/common/` |
| `AppLoading.vue` | 全局或局部加载状态 | `shared/ui/common/` |
| `AppSkeleton.vue` | 页面或局部数据加载骨架 | `shared/ui/common/` |
| `AppCard.vue` | 通用卡片容器 | `shared/ui/common/` |
| `SectionCard.vue` | 页面内分区卡片 | `shared/ui/common/` |
| `MetricCard.vue` | 指标展示卡片 | `shared/ui/common/` |
| `PageHeader.vue` | 页面顶部标题区 | `shared/ui/common/` |
| `SkillRadar.vue` | 技能雷达图 | `shared/ui/common/` |

#### 2.0.2 交互原语

| 组件 | 用途 | 代码位置 |
|------|------|---------|
| `AppToast.vue` | 全局 toast 渲染，消费 `useToast()` 状态 | `shared/ui/common/` |
| `AppDestructiveConfirm.vue` | 危险确认弹窗壳 | `shared/ui/common/` |
| `DeleteConfirmModal.vue` | 删除确认模态框 | `shared/ui/common/` |
| `CActionMenu.vue` | 通用操作菜单 | `shared/ui/common/menus/` |

#### 2.0.3 目录与数据表格

| 组件 | 用途 | 代码位置 |
|------|------|---------|
| `WorkspaceDataTable.vue` | 工作区目录表格骨架、列配置、插槽型单元格渲染 | `shared/ui/common/` |
| `WorkspaceDirectoryToolbar.vue` | 工作区目录工具栏 | `shared/ui/common/` |
| `WorkspaceDirectoryPagination.vue` | 工作区目录分页 | `shared/ui/common/` |
| `PagePaginationControls.vue` | 通用分页控件 | `shared/ui/common/` |

#### 2.0.4 Overlay 模板

| 组件 | 用途 | 代码位置 |
|------|------|---------|
| `OverlayPortal.vue` | `Teleport` 到 `body`、transition、backdrop 点击关闭 | `shared/ui/common/modal-templates/` |
| `ModalTemplateShell.vue` | OverlayPortal 包装层，注入 aria/role/panelClass/frosted | `shared/ui/common/modal-templates/` |
| `ClassicCenteredModal.vue` | 居中弹窗结构模板 | `shared/ui/common/modal-templates/` |
| `SlideOverDrawer.vue` | 右侧抽屉结构模板 | `shared/ui/common/modal-templates/` |
| `AdminSurfaceModal.vue` | 后台工作区常规居中弹窗 | `shared/ui/common/modal-templates/` |
| `AdminSurfaceDrawer.vue` | 后台工作区常规侧边抽屉 | `shared/ui/common/modal-templates/` |
| `MinimalFloatingModal.vue` | 轻量浮动模态框 | `shared/ui/common/modal-templates/` |
| `CFocusedInputDialog.vue` | 聚焦输入对话框 | `shared/ui/common/modal-templates/` |
| `CImmersiveConfirmDialog.vue` | 沉浸式确认对话框 | `shared/ui/common/modal-templates/` |
| `CLightActionPopover.vue` | 轻量操作弹出层 | `shared/ui/common/modal-templates/` |
| `CContextTooltip.vue` | 上下文提示浮层 | `shared/ui/common/modal-templates/` |

#### 2.0.5 组件使用规范

**复用原则**：

- **展示原语**：跨页面、跨 feature 复用的基础展示组件
- **交互原语**：全局反馈、确认流程的统一交互
- **目录与表格**：后台工作区目录页的标准结构
- **Overlay 模板**：弹窗、抽屉、浮层的行为层和结构模板

**不适合放在 `shared/ui/common/` 的内容**：

- 强业务语义的展示块（应放 `entities/**/ui` 或 `features/**/ui`）
- 只服务单一 feature 的 UI（应放 `features/**/ui`）
- 页面级工作台组合（应放 `widgets/**`）
- 页面数据编排和路由交互（应放 `features/**/model`）

#### 2.0.6 设计 Token 映射

共享组件的视觉风格通过以下 CSS 文件控制：

| CSS 文件 | 用途 |
|---------|------|
| `assets/styles/theme.css` | 全局主题变量（颜色、间距、字体） |
| `assets/styles/surface-shell-background.css` | 页面壳与背景 |
| `assets/styles/workspace-shell.css` | 工作区壳与按钮原语 |
| `assets/styles/journal-soft-surfaces.css` | 学生侧 soft journal 风格 |
| `assets/styles/journal-admin-shell.css` | 管理员侧工作区风格 |
| `assets/styles/journal-user-shell.css` | 用户侧壳风格 |

**设计 Token 使用规范**：

- 共享组件内部优先使用 `theme.css` 中的 CSS 变量
- 不在组件内硬编码颜色值或间距
- 页面级样式调整通过覆盖 CSS 变量，而非覆盖组件类
- 夜间模式适配通过 `[data-theme='dark']` 选择器统一处理

### 2.1 通用按钮采用边界

`workspace-shell.css` 中的 `ui-btn` 是通用工作区按钮原语。页面如果只是需要常规 primary、secondary、ghost、danger 动作，应直接组合 `ui-btn` 与对应变体类；页面局部只允许通过 `--ui-btn-*` 变量调整尺寸、边框强度或主题语义。

按钮迁移清单只保留在 Git 历史里，用于追溯当时的迁移顺序和候选范围；当前共享原语边界以本节为准。

边界：

- 错误页、空状态、工作区恢复动作、列表工具栏和普通表单动作默认使用 `ui-btn`。
- 页面不得新增 `xxx-action-primary / xxx-action-secondary` 这类平行按钮体系来重写 `border / background / hover / focus`。
- 若某个页面需要专用按钮外观，应先判断它是否应该成为共享变体；不能只在页面 scoped style 中复制一套按钮状态。
- 夜间模式 hover 边框应保持柔和，优先通过页面壳的 `--ui-btn-*-hover-border` 变量降低强度，而不是覆盖 `.ui-btn:hover`。

### 2.1 Student journal 按钮变体

`code/frontend/src/assets/styles/journal-soft-surfaces.css` 负责学生侧 soft journal 页面按钮的基础变体。

当前变体：

| 类名 | 当前负责 |
| --- | --- |
| `journal-btn-primary` | 学生侧主要动作按钮 |
| `journal-btn-secondary` | 学生侧普通行动按钮，保留边框并通过共享 token 适配 hover、focus 和夜间模式 |
| `journal-btn-outline` | 学生侧弱化入口或辅助动作按钮 |

边界：

- 页面组件只选择按钮语义类，不在 scoped style 中重写按钮的 `border / background / color / hover / focus`。
- 若页面需要表达业务状态，优先组合状态类与共享按钮变体；状态类不重新实现一套按钮主题。
- 夜间模式由 `journal-soft-surfaces.css` 的 token 和共享 selector 统一处理，不在页面里新增 `:global([data-theme='dark'])` 按钮补丁。
- 所有学生侧 soft journal 按钮在 light / dark 下都必须保留可见边框。按钮边框默认以 `--journal-control-border` 为参照混合，不使用纯 `transparent` 作为唯一边界参照。
- 夜间模式下按钮 hover 只做低强度边框增强，避免把 accent 直接打满到按钮边框。primary 可以略强于 secondary / outline，但 secondary / outline 默认只轻微提亮，方便定位且不刺眼。

### 2.2 目录与空状态

| 文件 | 当前负责 |
| --- | --- |
| `shared/ui/common/WorkspaceDataTable.vue` | 工作区目录表格骨架、列配置、插槽型单元格渲染 |
| `shared/ui/common/AppEmpty.vue` | 统一空状态壳和图标映射 |
| `shared/ui/common/AppSkeleton.vue` | 页面或局部数据加载骨架 |

说明：

- `WorkspaceDataTable` 当前只吃 `columns / rows / rowKey` 和具名单元格插槽，不持有分页、排序和请求。
- 空状态和加载状态继续作为显示原语，不自己决定“什么时候算空”。

### 2.3 全局反馈与危险操作

| 文件 | 当前负责 |
| --- | --- |
| `shared/ui/common/AppToast.vue` | 全局 toast 渲染，消费 `useToast()` 的状态 |
| `shared/ui/common/AppDestructiveConfirm.vue` | 危险确认弹窗壳，消费 `useDestructiveConfirmState()` 的状态 |

边界：

- `AppToast` 负责视觉呈现，不负责业务动作触发。
- `AppDestructiveConfirm` 可以给危险动作提供统一交互，但不直接删除资源；真正的删除逻辑仍在 feature model。

## 3. Overlay 模板体系

当前 overlay 体系分三层：

### 3.1 行为层

| 文件 | 当前负责 |
| --- | --- |
| `modal-templates/OverlayPortal.vue` | `Teleport` 到 `body`、transition、backdrop 点击关闭 |
| `modal-templates/useOverlayBehavior.ts` | overlay 栈、仅栈顶响应 Escape、body scroll lock |
| `modal-templates/ModalTemplateShell.vue` | OverlayPortal 包装层，注入 aria/role/panelClass/frosted 等公共壳能力 |

### 3.2 结构模板

| 文件 | 当前负责 |
| --- | --- |
| `ClassicCenteredModal.vue` | 居中弹窗结构 |
| `SlideOverDrawer.vue` | 右侧抽屉结构 |

### 3.3 语义模板

| 文件 | 当前负责 |
| --- | --- |
| `AdminSurfaceModal.vue` | 后台工作区常规居中弹窗 |
| `AdminSurfaceDrawer.vue` | 后台工作区常规侧边抽屉 |

当前复用规则：

- 新增后台弹窗或抽屉时，优先复用 `AdminSurfaceModal` / `AdminSurfaceDrawer`
- 只想复用行为，不想复用视觉骨架时，直接基于 `OverlayPortal` / `ModalTemplateShell` 组合
- 不再要求业务 overlay 全部走一个“万能模板”

## 4. 布局壳

`code/frontend/src/shared/ui/layout/AppLayout.vue` 当前是全局路由承载壳。

负责内容：

- 渲染 `Sidebar` 与 `TopNav`
- 通过 `RouterView` 承载所有路由页面
- 根据 `route.meta.contentLayout` 切换默认壳与 `bleed` 全宽布局
- 对 backoffice 页面隐藏 route-level `.workspace-topbar`
- 在 `onMounted()` 启动 `useNotificationRealtime()`

不负责内容：

- 不把各个页面的数据加载前置到 layout
- 不在 layout 内维护每个页面的查询条件或详情选中状态

### 4.1 Navigation composables

| 文件 | 当前负责 |
| --- | --- |
| `shared/model/navigation/useRouteNavigationTransport.ts` | 把 router push / replace 抽成可注入 transport，降低 feature model 对 router concrete 的依赖 |
| `shared/model/navigation/useRouteQueryTransport.ts` | 读写 route query 的 transport owner |
| `shared/model/navigation/useRouteQueryTabs.ts` | 将 route query 映射为 tab 状态 |
| `shared/model/navigation/useUrlSyncedTabs.ts` | 为 feature model 提供 URL 同步 tab 状态 |
| `shared/model/layout/useWorkspaceShellNavigation.ts` | layout shell 消费的工作区导航状态桥接 |

边界：

- route page 不直接持有 `useRouteQueryTabs()`；`routePageArchitectureBoundary.test.ts` 会拦截普通 route page 对 route hooks 的直接使用。
- navigation composables 不负责业务 tab 的名称、权限和请求；这些由 feature model 或 widget owner 决定。

## 5. 组件边界与迁移方向

历史 `components/` 目录已经清空；原先挂在这里的共享原语、布局壳和 feature-owned 大块 UI 已分别迁到 `shared/ui/**`、`shared/model/**`、`features/**/ui`。

后续约束：

- 共享原语继续沉到 `shared/ui/common/`
- 共享 UI 状态继续沉到 `shared/model/common/` 和 `shared/model/layout/`
- 页面行为优先下沉到 `features/**/model`
- 只服务单一 feature 的大块 UI 优先收进 `features/**/ui`
- route view 保持薄壳
- 不再新建回 `components/**` 作为长期 owner 层

### 5.1 业务组件分层规则（Entities vs Features 判断）

**判断原则**：

| 判断维度 | Entities | Features |
|---------|----------|----------|
| 主要回答的问题 | "这个业务对象是什么、如何稳定展示" | "用户在这里要完成什么动作" |
| 典型内容 | 卡片、标签、状态映射、轻量类型 | 上传、提交、筛选、编辑器、工作区 |
| 依赖方向 | 不依赖 features、pages、route state | 可以依赖 entities、shared、api |
| 复用场景 | 多个 feature 都会用到该对象的展示 | 只服务单一功能或工作区 |

**Entities 示例**：
- `entities/challenge/ui/ChallengeCard.vue` - 题目卡片展示
- `entities/contest/model/contestStatus.ts` - 竞赛状态映射
- `entities/user/ui/UserAvatar.vue` - 用户头像组件
- `entities/team/ui/TeamBadge.vue` - 队伍徽章

**Features 示例**：
- `features/challenge-detail/ui/ChallengeSubmitPanel.vue` - 题目提交面板
- `features/contest-create/ui/ContestForm.vue` - 竞赛创建表单
- `features/challenge-filter/ui/ChallengeFilterBar.vue` - 题目筛选栏
- `features/writeup-editor/ui/WriteupEditor.vue` - 题解编辑器

**决策树**：

```
这个组件主要做什么？
    ↓
展示业务对象的稳定属性 → Entities
    例：显示题目类型、难度、标签
    ↓
用户动作流程或工作区 → Features
    例：提交 Flag、筛选题目、编辑题解
    ↓
跨业务对象的通用 UI → Shared
    例：空状态、Toast、表格骨架
```

**边界守卫**：
- `entities/*` 不能 import `features/*` 或 `pages/*`
- `entities/*` 不能依赖 route state 或异步工作流
- 业务语义明确的展示块不应放在 `shared/*`

### 5.2 Workspace 组件设计（Widgets 与 Pages 关系）

**定位**：`widgets/*` 负责跨 feature 页面区块组合，承接工作区级完整内容区。

#### 与 Features 协作模式

**单一能力 Surface（Features）**：
- 职责：只服务单一功能的 UI 块
- 位置：`features/**/ui/`
- 示例：
  - `features/challenge-submit/ui/SubmitForm.vue`
  - `features/challenge-hints/ui/HintsList.vue`
  - `features/instance-control/ui/InstancePanel.vue`

**工作区级组合（Widgets）**：
- 职责：组合多个 features 和 entities 成完整工作区
- 位置：`widgets/*-workspace/`
- 示例：
  - `widgets/challenge-detail-workspace/` - 组合题面 + 提交 + 实例 + 题解
  - `widgets/contest-detail-workspace/` - 组合概览 + 公告 + 排行榜 + 题目
  - `widgets/awd-review-workspace/` - 组合复盘数据 + 攻击记录 + 轮次详情

**协作流程**：

```
Pages (路由入口)
    ↓ 组合
Widgets (工作区级组合)
    ↓ 组合
Features (单一能力) + Entities (业务对象展示)
    ↓ 消费
Shared (通用 UI 原语)
```

#### Workspace 命名约定

**命名格式**：
- 基本格式：`<context>-<entity>-workspace`
- 简化格式：`<context>-workspace`（当 context 已明确表达实体）

**当前命名实例**：

| 目录名 | 说明 |
|--------|------|
| `challenge-detail-workspace` | 题目详情工作区 |
| `contest-detail-workspace` | 竞赛详情工作区 |
| `contest-list-workspace` | 竞赛列表工作区 |
| `awd-review-workspace` | AWD 复盘工作区 |
| `notification-list-workspace` | 通知列表工作区 |
| `scoreboard-detail-workspace` | 排行榜详情工作区 |

**命名禁止模式**：
- 不使用 `*-page`（page 是路由入口，不是 widget）
- 不使用 `*-container`（过于泛化，无明确语义）
- 不使用 `*-view`（与 route view 混淆）

#### 与 Pages 关系

**Pages 职责**：
- 路由入口，只负责组合 widgets 或 features
- 处理路由参数解析和初始 query 同步
- 不应包含大段业务逻辑或复杂 UI 结构

**Widgets 职责**：
- 工作区级 UI 组合和布局协调
- 跨 feature 的状态桥接（如工作区 tabs 与子区块同步）
- 不应反向依赖 pages 或 route state

**决策规则**：

1. **内容需要组合多个 features** → 创建 widget
   ```typescript
   // pages/challenges/[id]/RoutePage.vue
   <template>
     <ChallengeDetailWorkspace :challenge-id="challengeId" />
   </template>
   ```

2. **内容只需要单一 feature** → 直接使用 feature UI
   ```typescript
   // pages/notifications/index/RoutePage.vue
   <template>
     <NotificationList />
   </template>
   ```

3. **内容需要复杂布局协调** → 创建 widget
   ```typescript
   // widgets/awd-review-workspace/AWDReviewWorkspace.vue
   <template>
     <div class="workspace-shell">
       <WorkspaceToolbar />
       <WorkspaceTabs v-model="activeTab" />
       <AWDReviewOverview v-if="activeTab === 'overview'" />
       <AWDAttackRecords v-else-if="activeTab === 'attacks'" />
     </div>
   </template>
   ```

**反例（不应出现）**：
- Pages 包含大段 UI 结构和业务逻辑
- Widgets 直接导入 `useRoute()` 或依赖路由参数
- Features UI 组件尝试承担工作区级布局协调

### 5.3 `feature-owned UI` 判定规则

以下条件同时成立时，默认直接落到 `features/*/ui/`，不要再新建 `components/**`：

- 组件只服务一个 feature 或一个 feature family
- 组件直接依赖该 feature 的 model/composable，或者只消费该 feature 暴露的 contract
- 组件承担的是该 feature 的 editor / manage / review / workspace 壳，而不是跨 feature 复用的中立展示

以下条件成立时，优先判断是否进入 `shared/ui/**` 或 `shared/model/**`：

- 组件会被多个 feature 或多个 route page 复用
- 组件不绑定单一 feature model，只接收中立 props / emits
- 组件本质上是共享原语、布局壳或跨业务复用的展示块

代表性例子：

- `features/platform/challenge-detail/ui/*`
  - 负责：平台题目详情 feature 自己的 Flag 配置 UI
  - 不负责：变成跨页面复用组件仓
- `features/challenge-writeup-editor/ui/*`
  - 负责：题解管理、编辑、查看这组只服务题解 feature 的 UI 面
  - 不负责：route view、题目详情导航或 API owner

## 6. Guardrail

- 共享弹窗模板存在性、关闭行为和 fallthrough 约束：`code/frontend/src/shared/ui/common/__tests__/ModalTemplates.test.ts`
- 目录表格应继续使用主题 token：`code/frontend/src/shared/ui/common/__tests__/WorkspaceDataTable.test.ts`
- 空状态表面样式：`code/frontend/src/shared/ui/common/__tests__/AppEmptySurface.test.ts`
- Toast 样式与交互：`code/frontend/src/shared/ui/common/__tests__/AppToast.test.ts`
- 布局壳与 backoffice/student 内容壳切换：`code/frontend/src/shared/ui/layout/__tests__/AppLayout.test.ts`
- 分层约束：`code/frontend/src/__tests__/architectureBoundaries.test.ts`
- route page 导航 hook 边界：`code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
- 工作区导航状态：`code/frontend/src/shared/model/layout/__tests__/useWorkspaceShellNavigation.test.ts`
