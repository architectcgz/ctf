# 页面设计：技能评估 (Dark Mode)

> 继承：../design-system/MASTER.md | 技术栈：Element Plus + Tailwind CSS

---

## 布局

```
┌──────────────────────────────────────────────┐
│  技能雷达图                                   │
│  [ECharts Radar]                             │
│  Web: ████████░░ 80%                         │
│  Pwn: ██████░░░░ 60%                         │
│  Crypto: ████░░░░░░ 40%                      │
├──────────────────────────────────────────────┤
│  技能成长趋势                                 │
│  [ECharts Line]                              │
└──────────────────────────────────────────────┘
```

## 技术栈

**Element Plus：**
- `<ElCard>` - 卡片容器
- `<ElProgress>` - 进度条

**ECharts：**
- Radar Chart - 技能雷达图
- Line Chart - 成长趋势

**Tailwind CSS：**
- Grid 布局：`grid grid-cols-1 lg:grid-cols-2 gap-6`
