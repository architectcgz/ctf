# 页面设计：网络拓扑编排器 (Topology Editor)

> 继承：design-system/MASTER.md | 角色：管理员 | 任务：T19
> 位置：管理后台 > 靶场管理 > 创建/编辑靶场 > 拓扑编排 Tab

---

## 1. 入口

靶场创建/编辑 Drawer 的分步表单中新增第 2 步"网络拓扑"（原流程：基本信息 → 镜像配置 → Flag → 资源限制 → 提示）。

调整后步骤：基本信息 → **网络拓扑** → Flag → 资源限制 → 提示

单容器靶场可跳过此步（默认单节点拓扑）。

点击"配置拓扑"按钮后，打开全屏 overlay 编排器（`fixed inset-0 z-50 bg-base`），编辑完成后点击"保存并返回"回到 Drawer 流程。Drawer 中显示拓扑摘要（节点数、网络数、缩略图预览）。

---

## 2. 拓扑编排器主布局

```
┌──────────────────────────────────────────────────────────┐
│  "网络拓扑编排"                                            │
│  [从模板加载▼]  [保存为模板]  [自动布局]  [重置]            │
├────────────────────────────────────────┬─────────────────┤
│                                        │ 节点属性面板     │
│  Canvas 画布区域                        │ (w-72)          │
│  (flex-1, bg-base, 网格背景)            │                 │
│                                        │ ── 选中节点 ──  │
│       ┌─────┐                          │ 名称: [web-01]  │
│       │ Web │ ←── 选中态               │ 镜像: [▼选择]   │
│       │ :80 │     border-primary       │ 端口: [80,443]  │
│       └──┬──┘                          │ 资源限制:       │
│          │ 连线 (stroke-primary/50)    │  CPU: [1核▼]    │
│       ┌──┴──┐                          │  内存: [512M▼]  │
│       │ DB  │                          │ 网络层: [▼]     │
│       │:3306│                          │                 │
│       └──┬──┘                          │ ── 网络连接 ──  │
│          │                             │ ☑ web-01 → db   │
│       ┌──┴──┐                          │ ☑ db → internal │
│       │内网  │                          │ ☐ web → internal│
│       │:22  │                          │                 │
│       └─────┘                          │ [删除节点]       │
│                                        │                 │
│  ── 工具栏 (底部浮动) ──                │                 │
│  [+ 添加节点] [+ 添加网络] [缩放] [适应] │                 │
├────────────────────────────────────────┴─────────────────┤
│  资源预估: 3 容器 | CPU 3核 | 内存 1.5GB | 2 网络         │
│  ⚠ 提示: 当前拓扑需要较多资源，建议控制在 5 节点以内       │
└──────────────────────────────────────────────────────────┘
```

---

## 3. 画布区域

### 背景
- 网格点阵背景：`radial-gradient` 点阵，颜色 `#21262d`，间距 20px
- 支持拖拽平移和滚轮缩放

### 节点样式

```
默认节点:
  bg-surface border border-border rounded-lg px-3 py-2
  min-w-[80px] text-center cursor-grab

  [图标 20px text-muted]        -- Lucide: Server/Database/Globe/Monitor
  [名称 text-sm text-primary]   -- "web-01"
  [端口 text-xs text-muted font-mono]  -- ":80"

选中节点:
  border-primary bg-primary/5 shadow-[0_0_0_1px_#0891b2]

hover 节点:
  border-primary/50

网络层标签 (节点左上角):
  text-[10px] px-1 rounded bg-{层级色}/10 text-{层级色}
  外网: cyan    中间层: amber    内网: violet
```

### 连线样式

```
默认连线:
  stroke: #30363d  stroke-width: 2  stroke-dasharray: none

选中连线:
  stroke: #0891b2  stroke-width: 2

不可达连线 (手动断开):
  stroke: #ef4444/30  stroke-width: 1  stroke-dasharray: 4,4
```

### 网络层可视化

```
三层网络用半透明背景区域区分:

外网层 (DMZ):     bg-cyan-500/3   border border-dashed border-cyan-500/15
中间层 (Service): bg-amber-500/3  border border-dashed border-amber-500/15
内网层 (Internal):bg-violet-500/3 border border-dashed border-violet-500/15

层标签: 左上角 text-xs text-{层级色}/60
```

---

## 4. 节点属性面板

- 位置：画布右侧固定面板，`w-72 bg-surface border-l border-border`
- 未选中节点时显示拓扑概览（节点数、网络数、资源汇总）
- 选中节点时显示该节点的配置表单

### 表单字段
- 节点名称：文本输入，`text-sm`
- 关联镜像：下拉选择（从已有镜像列表）
- 暴露端口：多值输入（逗号分隔，`font-mono`）
- 资源限制：CPU / 内存下拉选择
- 所属网络层：单选（外网 / 中间层 / 内网）
- 网络连接：checkbox 列表，控制与其他节点的可达性

---

## 5. 底部资源预估栏

```
bg-elevated border-t border-border px-4 py-3 text-sm

正常: text-secondary
  "3 容器 | CPU 3核 | 内存 1.5GB | 2 网络"

警告 (节点 > 5 或资源超阈值):
  bg-warning/5 border-warning/20
  [AlertTriangle 16px text-warning] + 警告文案
```

---

## 6. 交互说明

| 操作 | 行为 |
|------|------|
| 点击画布空白 | 取消选中 |
| 点击节点 | 选中，右侧显示属性 |
| 拖拽节点 | 移动位置 |
| 从节点端口拖出 | 创建连线到目标节点 |
| 双击连线 | 选中连线，可删除 |
| 滚轮 | 缩放画布 |
| 空格+拖拽 | 平移画布 |
| Delete 键 | 删除选中的节点/连线 |

### 删除节点确认

```
Dialog:
  "确认删除节点 web-01？"
  "关联的 2 条网络连接将同时删除"
  [取消] [确认删除(danger)]
```

---

## 7. 技术实现建议

- 画布渲染：推荐 Vue Flow（基于 D3）或自定义 Canvas/SVG
- 拓扑数据结构：JSONB，包含 nodes[] 和 edges[]
- 节点拖拽：HTML5 Drag API 或 pointer events
