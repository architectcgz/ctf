# 前端组件体系设计

> 状态：Current
> 事实源：`code/frontend/src/components/`、`code/frontend/src/views/`、`code/frontend/src/features/**/model/`
> 替代：无

## 定位

本文档只说明前端组件应该放在哪一层、共享原语有哪些、弹窗模板怎么复用，以及布局壳的 owner。

- 覆盖：`components/common/`、`components/layout/`、`components/common/modal-templates/`、遗留业务组件目录和 route view 的边界。
- 不覆盖：单个页面的数据流细节；这些见 `07-pages-dataflow.md`。

## 当前设计

- `code/frontend/src/components/common/`
  - 负责：共享展示原语和通用交互壳，例如 `AppEmpty.vue`、`AppToast.vue`、`AppSkeleton.vue`、`WorkspaceDataTable.vue`、`DeleteConfirmModal.vue`
  - 不负责：直接发业务请求、直接依赖 store/router，或内置某个页面的专用流程

- `code/frontend/src/components/common/modal-templates/`
  - 负责：Overlay 行为层、居中弹窗、侧边抽屉和后台工作区模板，例如 `OverlayPortal.vue`、`ModalTemplateShell.vue`、`ClassicCenteredModal.vue`、`SlideOverDrawer.vue`、`AdminSurfaceModal.vue`、`AdminSurfaceDrawer.vue`
  - 不负责：把所有业务弹窗都揉成一套万能模板；业务样式明显不同的 overlay 仍应自有 DOM/CSS，只复用 headless 行为层

- `code/frontend/src/components/layout/AppLayout.vue`
  - 负责：应用总布局、侧栏、顶栏、全局通知实时连接、route transition 和 backoffice/student 的内容壳切换
  - 不负责：页面自己的业务查询、目录筛选或详情页状态机

- `code/frontend/src/views/**` 与 `code/frontend/src/features/**/model`
  - 负责：view 作为路由入口，feature model 作为页面行为 owner；复杂页面继续优先拆到 feature model 或局部业务组件
  - 不负责：把页面控制逻辑继续堆回共享原语层

## 1. 组件层次

当前前端组件按下面的 owner 分层：

| 层级 | 当前位置 | 当前职责 |
| --- | --- | --- |
| 共享原语 | `components/common/` | 空状态、Toast、Skeleton、目录表格、删除确认、通用局部承载器 |
| Overlay 模板 | `components/common/modal-templates/` | Teleport、滚动锁、Escape/backdrop 关闭、经典弹窗和抽屉模板 |
| 布局壳 | `components/layout/` | `AppLayout`、`Sidebar`、`TopNav` 等全局承载 |
| 业务展示组件 | `components/teacher/`、`components/platform/`、`components/contests/`、`components/scoreboard/` 等 | 领域相关展示和局部桥接，逐步向 feature owner 过渡 |
| 路由 view | `views/**` | 页面壳与 feature 组合入口 |

判断原则：

- 能在多个页面复用且不绑业务 owner 的，进 `components/common/`
- 只解决 overlay 行为和模板骨架的，进 `components/common/modal-templates/`
- 强业务语义的展示组件保留在业务目录，不伪装成“通用组件”
- 页面数据编排和路由交互不放进共享组件

## 2. 当前共享原语

### 2.0 通用按钮采用边界

`workspace-shell.css` 中的 `ui-btn` 是通用工作区按钮原语。页面如果只是需要常规 primary、secondary、ghost、danger 动作，应直接组合 `ui-btn` 与对应变体类；页面局部只允许通过 `--ui-btn-*` 变量调整尺寸、边框强度或主题语义。

当前全仓按钮迁移清单见 `docs/todos/2026-05-10-frontend-button-primitive-migration-audit.md`。该清单是后续迁移顺序和候选范围，不替代本节的共享原语边界。

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
| `components/common/WorkspaceDataTable.vue` | 工作区目录表格骨架、列配置、插槽型单元格渲染 |
| `components/common/AppEmpty.vue` | 统一空状态壳和图标映射 |
| `components/common/AppSkeleton.vue` | 页面或局部数据加载骨架 |

说明：

- `WorkspaceDataTable` 当前只吃 `columns / rows / rowKey` 和具名单元格插槽，不持有分页、排序和请求。
- 空状态和加载状态继续作为显示原语，不自己决定“什么时候算空”。

### 2.3 全局反馈与危险操作

| 文件 | 当前负责 |
| --- | --- |
| `components/common/AppToast.vue` | 全局 toast 渲染，消费 `useToast()` 的状态 |
| `components/common/DeleteConfirmModal.vue` | 危险确认弹窗，基于 `ModalTemplateShell` 组合默认文案和动作 |

边界：

- `AppToast` 负责视觉呈现，不负责业务动作触发。
- `DeleteConfirmModal` 可以给危险动作提供统一交互，但不直接删除资源；真正的删除逻辑仍在 feature model。

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

`code/frontend/src/components/layout/AppLayout.vue` 当前是全局路由承载壳。

负责内容：

- 渲染 `Sidebar` 与 `TopNav`
- 通过 `RouterView` 承载所有路由页面
- 根据 `route.meta.contentLayout` 切换默认壳与 `bleed` 全宽布局
- 对 backoffice 页面隐藏 route-level `.workspace-topbar`
- 在 `onMounted()` 启动 `useNotificationRealtime()`

不负责内容：

- 不把各个页面的数据加载前置到 layout
- 不在 layout 内维护每个页面的查询条件或详情选中状态

## 5. 组件边界与迁移方向

当前代码里仍存在 `components/teacher/`、`components/platform/`、`components/contests/` 等业务目录，这是现状事实，不是新的共享组件入口。

后续约束：

- 共享原语继续沉到 `components/common/`
- 页面行为优先下沉到 `features/**/model`
- route view 保持薄壳
- 业务展示组件可以继续存在，但不要再向其内堆新的页面级状态机

## 6. Guardrail

- 共享弹窗模板存在性、关闭行为和 fallthrough 约束：`code/frontend/src/components/common/__tests__/ModalTemplates.test.ts`
- 目录表格应继续使用主题 token：`code/frontend/src/components/common/__tests__/WorkspaceDataTable.test.ts`
- 空状态表面样式：`code/frontend/src/components/common/__tests__/AppEmptySurface.test.ts`
- Toast 样式与交互：`code/frontend/src/components/common/__tests__/AppToast.test.ts`
- 布局壳与 backoffice/student 内容壳切换：`code/frontend/src/components/layout/__tests__/AppLayout.test.ts`
- 分层约束：`code/frontend/src/__tests__/architectureBoundaries.test.ts`
