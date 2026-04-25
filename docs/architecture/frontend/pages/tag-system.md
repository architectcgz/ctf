# 页面设计：标签体系管理 (Tag System)

> 继承：../design-system/MASTER.md | 角色：管理员 | 任务：T21
> 技术栈：Element Plus (主) + Tailwind CSS (辅)
> 位置：管理后台 > 靶场管理 > 标签管理 Tab

---

## 技术栈

**Element Plus：** `<ElTabs>`, `<ElButton>`, `<ElDialog>`, `<ElRadioGroup>`, `<ElInput>`, `<ElCheckbox>`, `<ElTag>`
**Tailwind CSS：** 布局、间距

---

## 1. 标签管理页

```
┌──────────────────────────────────────────────────────────┐
│  "标签管理"                              [+ 创建标签]     │
│  [漏洞类型(12)] [技术栈(8)] [知识点(15)]                  │
├──────────────────────────────────────────────────────────┤
│  标签列表 (当前: 漏洞类型)                                │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │ SQL注入           关联 18 道靶场    [编辑] [删除]   │  │
│  ├────────────────────────────────────────────────────┤  │
│  │ XSS               关联 12 道靶场    [编辑] [删除]   │  │
│  ├────────────────────────────────────────────────────┤  │
│  │ 命令注入           关联 8 道靶场     [编辑] [删除]   │  │
│  ├────────────────────────────────────────────────────┤  │
│  │ 文件上传           关联 6 道靶场     [编辑] [删除]   │  │
│  ├────────────────────────────────────────────────────┤  │
│  │ SSRF              关联 3 道靶场     [编辑] [删除]   │  │
│  └────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

### 维度 Tab

```
Tab 样式复用 MASTER 规范:
  默认: text-secondary hover:text-primary
  活跃: text-primary border-b-2 border-primary

Tab 后括号内数字: text-muted text-xs
```

### 标签行

```
bg-surface border-b border-border-subtle px-4 py-3
hover:bg-elevated transition-colors duration-100

标签名: text-sm text-primary font-medium
关联数: text-xs text-muted  "关联 18 道靶场"
操作: text-sm text-secondary hover:text-primary
```

### 删除保护

关联靶场数 > 0 时，删除按钮点击弹出确认：

```
"标签「SQL注入」已关联 18 道靶场，删除后这些靶场将移除该标签。"
[取消] [确认删除(danger)]
```

---

## 2. 创建/编辑标签 Dialog

```
┌─────────────────────────────────────┐
│  "创建标签"                  [关闭]  │
├─────────────────────────────────────┤
│  所属维度                           │
│  (○) 漏洞类型  (○) 技术栈  (○) 知识点│
│                                     │
│  标签名称                           │
│  [________________]                 │
│                                     │
│  描述 (可选)                        │
│  [________________]                 │
│                                     │
│  [取消]              [创建]          │
└─────────────────────────────────────┘
```

- Dialog 宽度：`max-w-sm`
- 编辑模式：维度不可修改（灰显）

---

## 3. 靶场列表筛选增强

在现有靶场列表页（challenges.md）的筛选栏增加多维度标签筛选：

```
原筛选栏:
  [搜索框] [分类▼] [难度▼] [状态▼] [排序▼]

增强后:
  [搜索框] [分类▼] [难度▼] [状态▼] [排序▼]
  标签筛选: [漏洞类型▼] [技术栈▼] [知识点▼]  [清除筛选]
```

### 标签筛选下拉

```
多选下拉 (Checkbox Dropdown):
  max-h-[240px] overflow-y-auto
  bg-elevated border border-border rounded-lg shadow-lg p-1

  每项:
    px-3 py-2 hover:bg-surface rounded
    [☑/☐] [标签名 text-sm] [关联数 text-xs text-muted 右对齐]

  底部:
    border-t border-border-subtle pt-2 mt-1
    [清除] [确认]
```

### 已选标签展示

筛选栏下方显示已选标签 chips：

```
flex flex-wrap gap-2 mt-2

每个 chip:
  bg-primary/10 text-primary text-xs px-2 py-1 rounded-full
  [标签名] [X 关闭]

维度前缀:
  漏洞类型 chip: bg-danger/10 text-danger
  技术栈 chip:   bg-cyan-500/10 text-cyan-400
  知识点 chip:   bg-violet-500/10 text-violet-400
```

筛选逻辑：同维度内 OR，跨维度 AND。
