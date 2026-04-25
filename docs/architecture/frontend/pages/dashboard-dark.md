# 页面设计：Dashboard (Dark Mode)

> 继承：../design-system/MASTER.md | 技术栈：Element Plus + Tailwind CSS

---

## 布局结构

```
┌──────────────────────────────────────────────┐
│  欢迎回来，[用户名]                            │
├──────────┬──────────┬──────────┬──────────────┤
│ 已解题数  │ 总得分    │ 全站排名  │ 在线实例     │
│ 12       │ 850      │ #23      │ 1/3         │
├──────────┴──────────┴──────────┴──────────────┤
│  最近活动                                     │
│  ┌─────────────────────────────────────────┐ │
│  │ [时间] 解出 SQL 注入基础 (+100分)        │ │
│  │ [时间] 启动实例：栈溢出入门              │ │
│  └─────────────────────────────────────────┘ │
│  推荐挑战                                     │
│  [卡片] [卡片] [卡片]                         │
└──────────────────────────────────────────────┘
```

## 技术栈

**Element Plus：**
- `<ElCard>` - 统计卡片、活动列表
- `<ElTag>` - 分类标签
- `<ElEmpty>` - 空状态

**Tailwind CSS：**
- Grid 布局：`grid grid-cols-2 md:grid-cols-4 gap-4`
- 间距：`space-y-6`

## 代码示例

```vue
<div class="space-y-6">
  <h1 class="text-2xl font-bold text-[#e6edf3]">
    欢迎回来，{{ username }}
  </h1>

  <div class="grid grid-cols-2 gap-4 md:grid-cols-4">
    <ElCard>
      <div class="text-center">
        <div class="text-3xl font-mono font-bold text-[#0891b2]">12</div>
        <div class="mt-2 text-sm text-[#8b949e]">已解题数</div>
      </div>
    </ElCard>
    <!-- 其他统计卡片 -->
  </div>

  <ElCard>
    <template #header>
      <span class="text-[#e6edf3]">最近活动</span>
    </template>
    <!-- 活动列表 -->
  </ElCard>
</div>
```
