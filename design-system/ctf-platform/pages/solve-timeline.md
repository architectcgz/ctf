# 页面设计：解题时间线 (Solve Timeline)

> 继承：design-system/MASTER.md | 角色：学员 | 任务：T24
> 技术栈：Element Plus (主) + Tailwind CSS (辅)
> 位置：靶场详情页 > 新增"解题记录"Tab；个人主页 > 解题历史

---

## 技术栈

**Element Plus：** `<ElTabs>`, `<ElTimeline>`, `<ElEmpty>`
**Tailwind CSS：** 布局、间距

---

## 1. 靶场详情页 — 解题记录 Tab

在靶场详情页增加 Tab 切换：[题目详情] [解题记录]

仅当用户对该靶场有操作记录时显示此 Tab。

```
┌──────────────────────────────────────────────────────────┐
│  SQL注入基础                                   100 pts   │
│  [题目详情] [解题记录]                                    │
├──────────────────────────────────────────────────────────┤
│  解题时间线                          耗时: 1h 23min      │
│                                                          │
│  03-01                                                   │
│  ●─── 14:05  启动靶机实例                                │
│  │    实例 ID: inst-a3f2  地址: 10.10.1.42:8080          │
│  │                                                       │
│  ●─── 14:18  提交 Flag (失败)                            │
│  │    flag{test123}                                      │
│  │                                                       │
│  ●─── 14:25  解锁 Hint 1 (-10 pts)                      │
│  │    "注意登录表单的参数传递方式"                         │
│  │                                                       │
│  ●─── 14:52  提交 Flag (失败)                            │
│  │    flag{sql_inject_1}                                 │
│  │                                                       │
│  ◉─── 15:28  提交 Flag (成功) ✓                          │
│  │    flag{y0u_g0t_sql_1nj3ct10n}                       │
│  │    得分: 90 pts (原 100 - Hint 10)                    │
│  │                                                       │
│  ●─── 15:28  靶机实例自动销毁                             │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

---

## 2. 时间线组件样式

### 时间轴线

```
竖线: w-0.5 bg-border-subtle  (贯穿整个时间线)
节点: w-3 h-3 rounded-full border-2

日期分隔: text-xs text-muted font-medium mb-3 mt-6 (首个日期无 mt)
```

### 事件节点色标

```
启动靶机:   border-primary bg-primary/20     图标: Play 12px
Flag 成功:  border-success bg-success/20     图标: Check 12px
Flag 失败:  border-danger bg-danger/20       图标: X 12px
解锁提示:   border-warning bg-warning/20     图标: Lightbulb 12px
实例销毁:   border-muted bg-border           图标: Square 12px
实例延时:   border-primary bg-primary/20     图标: Clock 12px
```

### 事件卡片

```
ml-6 (距时间轴线的左边距)

时间: text-xs text-muted font-mono  "14:05"
标题: text-sm text-primary font-medium  "启动靶机实例"
详情: text-xs text-secondary mt-1

Flag 内容: font-mono text-xs
  成功: text-success
  失败: text-danger/60

成功事件卡片:
  bg-success/5 border border-success/20 rounded-lg p-3
  其余事件无背景色
```

### 耗时统计

```
右上角: text-xs text-muted
"耗时: 1h 23min" (从首次启动到成功提交)
未完成: "进行中" text-primary
```

---

## 3. 空状态

```
居中:
  [Clock 48px text-muted]
  "暂无解题记录"
  "启动靶机开始挑战后，操作记录将显示在这里"
```
