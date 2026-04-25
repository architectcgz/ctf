# 页面设计：整体布局 (Layout)

> 继承：../design-system/MASTER.md
> 技术栈：Element Plus (主) + Tailwind CSS (辅)

---

## 技术栈

**Element Plus：** `<ElMenu>`, `<ElDropdown>`, `<ElBadge>`, `<ElAvatar>`
**Tailwind CSS：** Flex布局、响应式

---

## 1. 应用外壳结构

```
┌──────────────────────────────────────────────────┐
│  TopNav (h-14, fixed, bg-base, border-b)          │
│  [≡][Logo CTF Arena]          [通知] [用户头像▼]  │
├──────┬───────────────────────────────────────────┤
│ Side │  Main Content Area                         │
│ w-16 │  (max-w-7xl mx-auto px-6 py-6)            │
│(折叠)│                                            │
│      │                                            │
│ 图标  │                                            │
│ 图标  │                                            │
│ 图标  │                                            │
│ 图标  │                                            │
│      │                                            │
│ ──── │                                            │
│ 图标  │  (管理员可见)                               │
└──────┴───────────────────────────────────────────┘

侧边栏 hover 展开为 w-56 时：
├──────────┬───────────────────────────────────────┤
│ [图标] 仪表盘  │                                   │
│ [图标] 靶场训练 │                                   │
│ [图标] 竞赛中心 │                                   │
│ ...            │                                   │
```

## 2. 顶部导航 (TopNav)

- 固定定位，`top-0 left-0 right-0 z-50`
- 背景：`bg-[#0f1117] border-b border-[#30363d]`
- 高度：`h-14`

左侧：
- 汉堡菜单按钮（md 以下显示）
- Logo + 平台名称 "CTF Arena"，`text-base font-semibold text-[#e6edf3]`

右侧：
- 通知铃铛（Lucide: Bell, 24px）+ 未读数红点 `bg-danger`
- 用户头像（32px 圆形）+ 下拉菜单（个人中心、退出）

## 3. 侧边栏 (Sidebar)

- 固定定位，`left-0 top-14 bottom-0`
- 背景：`bg-[#0f1117] border-r border-[#30363d]`
- 默认折叠 `w-16`（仅图标），hover 展开 `w-56`（图标+文字）
- 展开时使用 `transition-all duration-200`

导航项：
```
折叠态: [Icon 24px] 居中，tooltip 显示文字
展开态: [Icon 20px] [文字 text-sm]

默认态:  text-[#8b949e] hover:text-[#e6edf3] hover:bg-[#1c2128]
活跃态:  text-[#0891b2] bg-[#0891b2]/10 border-l-2 border-[#0891b2]
```

学员导航项：
- 仪表盘 (LayoutDashboard)
- 靶场训练 (Target)
- 竞赛中心 (Trophy)
- 排行榜 (Medal)
- 我的实例 (Server)
- 能力评估 (Radar)

教师额外项（分隔线下方）：
- 班级管理 (GraduationCap)
- 学员进度 (Users)

管理员额外项：
- 靶场管理 (Settings)
- 镜像管理 (Container)
- 竞赛管理 (CalendarClock)
- 用户管理 (UserCog)
- 作弊检测 (ShieldAlert)
- 审计日志 (FileText)
- 系统监控 (Activity)

## 4. 移动端适配

- `< md`：侧边栏隐藏，通过汉堡菜单触发抽屉覆盖
- TopNav 保持固定
- 内容区域 `px-4 py-4`
