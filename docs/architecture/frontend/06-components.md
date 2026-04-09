# 前端组件体系设计

> 对应：design-system/MASTER.md 组件规范

---

## 1. 组件分层

```
components/
├── common/      # 基础 UI 原子组件（无业务逻辑）
├── layout/      # 布局组件（AppLayout, TopNav, Sidebar）
└── charts/      # ECharts 封装组件

views/           # 页面级组件（路由对应）
└── 各页面内可拆分局部子组件（不提升到 components/）
```

原则：
- `common/` 组件不依赖任何 Store 或 API，仅通过 props/emits 通信
- 组件统一使用 `<script setup lang="ts">`，Props/Emits 必须有明确类型
- **Element Plus 优先直用**：页面中可直接使用 `El*` 组件；`App*` 组件仅在需要统一业务行为/视觉 token 时封装
- 页面内复用 ≤ 2 次的子组件放在同目录下，不提升到全局
- 图表组件统一封装 ECharts 初始化/销毁/resize 逻辑

---

## 2. 基础组件清单

> 说明：以下 `App*` 组件为“统一业务行为/样式 token 的薄封装”。基础表单/表格/弹窗优先使用 Element Plus 的 `El*` 组件，避免重复造轮子。

### 2.1 AppButton（可选封装）

| Prop | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| variant | `'primary' \| 'secondary' \| 'danger' \| 'ghost'` | `'primary'` | 按钮样式 |
| size | `'sm' \| 'md' \| 'lg'` | `'md'` | 尺寸 |
| loading | boolean | false | 加载态（spinner + 禁用） |
| disabled | boolean | false | 禁用态 |
| icon | string | - | Lucide 图标名（按钮前置图标） |

实现建议：内部使用 `ElButton`，统一 `loading/disabled` 行为与主题色映射。

### 2.2 AppCard

| Prop | 类型 | 说明 |
|------|------|------|
| hoverable | boolean | 是否有 hover 边框效果 |
| highlight | `'success' \| 'primary' \| null` | 高亮边框色 |
| padding | `'sm' \| 'md' \| 'lg'` | 内边距 |

Slots: `default`, `header`, `footer`

强制约定（数值展示卡片）：
- 数值展示型卡片（MetricCard、summary item、dashboard KPI）默认必须带说明性文字（`hint`/`helper`），不能只显示“指标名 + 数值”。
- 说明文字至少要覆盖统计口径、时间范围或状态含义中的一项，避免歧义。

### 2.3 AppInput

| Prop | 类型 | 说明 |
|------|------|------|
| modelValue | string | v-model 绑定 |
| type | `'text' \| 'password' \| 'flag'` | flag 类型自动应用 font-mono + 主色边框 |
| placeholder | string | 占位文字 |
| error | string | 错误提示文字 |
| disabled | boolean | 禁用态 |
| prefix | string | 前置图标名 |
| suffix | string | 后置图标名 |

实现建议：普通输入优先用 `ElInput`；`type='flag'` 可做 `AppFlagInput`（更贴合业务），而不是强行扩展通用 Input。

### 2.4 AppTable

| Prop | 类型 | 说明 |
|------|------|------|
| columns | `Column[]` | 列定义 `{ key, title, width?, align?, render? }` |
| data | `any[]` | 表格数据 |
| loading | boolean | 加载态（skeleton 行） |
| selectable | boolean | 是否显示复选框列 |
| highlightTop | number | 高亮前 N 行（排行榜用） |

Events: `@select`, `@row-click`

实现建议：管理端表格优先使用 `ElTable`；排行榜等需要“高亮/动画/冻结状态”的场景才考虑封装自定义表格展示组件。

### 2.5 AppPagination

| Prop | 类型 | 说明 |
|------|------|------|
| total | number | 总条数 |
| page | number | 当前页 |
| pageSize | number | 每页条数 |

Events: `@change(page)`, `@size-change(size)`

### 2.6 AppTag

| Prop | 类型 | 说明 |
|------|------|------|
| color | string | 标签颜色（支持分类色标/难度色标/自定义） |
| size | `'sm' \| 'md'` | 尺寸 |

### 2.7 AppDialog / AppDrawer

| Prop | 类型 | 说明 |
|------|------|------|
| modelValue | boolean | v-model 控制显隐 |
| title | string | 标题 |
| width | string | 宽度（Drawer 默认 `480px`，Dialog 默认 `480px`） |

Slots: `default`, `footer`

### 2.8 AppToast

全局单例，通过 `useToast()` composable 调用，不直接使用组件。

### 2.9 AppEmpty

| Prop | 类型 | 说明 |
|------|------|------|
| icon | string | Lucide 图标名 |
| title | string | 主文案 |
| description | string | 副文案 |

Slots: `action`（引导按钮）

### 2.10 AppSkeleton

| Prop | 类型 | 说明 |
|------|------|------|
| rows | number | 骨架行数 |
| type | `'text' \| 'card' \| 'table'` | 骨架形态 |
