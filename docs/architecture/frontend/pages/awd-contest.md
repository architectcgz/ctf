# 页面设计：AWD 竞赛页面 (AWD Contest)

> 继承：../design-system/MASTER.md | 角色：学员 | 任务：T26
> 技术栈：Element Plus (主) + Tailwind CSS (辅)
> 位置：竞赛比赛页 > AWD 模式专用布局

---

## 技术栈

**Element Plus：** `<ElTabs>`, `<ElTable>`, `<ElCard>`, `<ElTag>`, `<ElButton>`, `<ElInput>`, `<ElDialog>`, `<ElTooltip>`
**Tailwind CSS：** Grid布局、Flex布局、间距、响应式

---

## 1. AWD 比赛页主布局

AWD 模式与 Jeopardy 模式共用竞赛入口，通过竞赛模式字段区分渲染。

```
┌──────────────────────────────────────────────────────────┐
│  "2026 春季 AWD 攻防对抗赛"              轮次: 第 12/30 轮│
│  剩余: 02:15:30 (font-mono)        本轮倒计时: 02:48     │
│  [战况总览] [我的靶机] [攻击面板] [排行榜]                 │
├──────────────────────────────────────────────────────────┤
│                                                          │
│  (Tab 内容区域，见下方各 Tab 设计)                         │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

### 顶部信息栏

```
bg-surface border-b border-border px-6 py-4

竞赛名: text-2xl font-bold text-primary
轮次信息: text-sm text-secondary font-mono
  "第 12/30 轮" — 当前轮次/总轮次
总倒计时: text-lg font-mono text-primary
本轮倒计时: text-sm font-mono
  > 60s: text-secondary
  ≤ 60s: text-warning
  ≤ 10s: text-danger + pulse 动画
```

---

## 2. Tab：战况总览

```
┌──────────────────────────────────────────────────────────┐
│  队伍状态矩阵                                             │
│  ┌──────┬────────┬────────┬────────┬────────┬─────────┐  │
│  │ 队伍  │ 服务状态 │ 攻击分  │ 防守分  │ 总分    │ 排名   │  │
│  ├──────┼────────┼────────┼────────┼────────┼─────────┤  │
│  │ 我方  │ ● 正常  │ 1,200  │ 800   │ 2,000  │ #2     │  │
│  │ TeamA│ ● 正常  │ 1,500  │ 750   │ 2,250  │ #1     │  │
│  │ TeamB│ ○ 宕机  │ 600   │ 200   │ 800   │ #5     │  │
│  │ TeamC│ ● 正常  │ 900   │ 700   │ 1,600  │ #3     │  │
│  └──────┴────────┴────────┴────────┴────────┴─────────┘  │
│                                                          │
│  ┌─────────────────────┐ ┌────────────────────────────┐  │
│  │ 攻防得分趋势         │ │ 本轮攻击事件               │  │
│  │ ECharts Line        │ │ 14:32 我方 → TeamB ✓      │  │
│  │ 攻击分(实线)         │ │ 14:28 TeamA → 我方 ✗      │  │
│  │ 防守分(虚线)         │ │ 14:25 TeamC → TeamB ✓     │  │
│  │ 双Y轴              │ │ ...                        │  │
│  └─────────────────────┘ └────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

### 服务状态指示

```
正常: [Circle 8px] fill-success  "正常"
宕机: [Circle 8px] fill-danger   "宕机"
检测中: [Circle 8px] fill-warning animate-pulse  "检测中"
```

### 攻击事件流

```
bg-surface border border-border rounded-lg p-4 max-h-[300px] overflow-y-auto

每条事件:
  text-xs py-1.5 border-b border-border-subtle last:border-0
  时间: font-mono text-muted w-12
  内容: text-secondary
  攻击成功: text-success "✓"
  攻击失败/被防御: text-danger "✗"
  涉及我方: text-primary font-medium (高亮我方队名)
```

---

## 3. Tab：我的靶机

```
┌──────────────────────────────────────────────────────────┐
│  我的服务靶机                                             │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │ web-service                          状态: ● 正常  │  │
│  │ 地址: 10.10.2.42:80 [复制]                         │  │
│  │ 当前轮 Flag: flag{awd_r12_...} [复制] (font-mono)  │  │
│  │ 连续存活: 12 轮                                     │  │
│  │ 被攻击: 3 次 (本轮)                                 │  │
│  │                                                    │  │
│  │ [SSH 连接信息]  [重置服务]                           │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  服务存活历史 (近 12 轮)                                  │
│  ┌────────────────────────────────────────────────────┐  │
│  │ R1 R2 R3 R4 R5 R6 R7 R8 R9 R10 R11 R12           │  │
│  │ ■  ■  ■  □  ■  ■  ■  ■  ■  ■   ■   ■            │  │
│  └────────────────────────────────────────────────────┘  │
│  ■ 存活  □ 宕机                                         │
└──────────────────────────────────────────────────────────┘
```

### 靶机卡片

```
bg-surface border border-border rounded-lg p-5

状态正常: border-l-2 border-success
状态宕机: border-l-2 border-danger bg-danger/3

服务名: text-lg font-medium text-primary
地址: font-mono text-sm text-secondary
Flag: font-mono text-xs text-primary/70 (每轮刷新)
```

### 存活历史格

```
inline-flex gap-1
每格: w-5 h-5 rounded-sm
  存活: bg-success/60
  宕机: bg-danger/60
tooltip: "第 4 轮: 宕机 (14:12 ~ 14:17)"
```

### 重置服务

```
条件: 连续宕机 ≥ N 轮时可用，否则按钮禁用
确认 Dialog:
  "重置将恢复服务到初始状态，你的自定义修改将丢失。"
  "本场比赛剩余重置次数: 2"
  [取消] [确认重置(danger)]
```

---

## 4. Tab：攻击面板

```
┌──────────────────────────────────────────────────────────┐
│  选择目标队伍                                             │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐         │
│  │TeamA │ │TeamB │ │TeamC │ │TeamD │ │TeamE │         │
│  │● 正常│ │○ 宕机│ │● 正常│ │● 正常│ │● 正常│         │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘         │
├──────────────────────────────────────────────────────────┤
│  目标: TeamA                                             │
│  靶机地址: 10.10.3.42:80 [复制]                           │
│                                                          │
│  提交攻击 Flag                                           │
│  ┌──────────────────────────────────┐                    │
│  │ flag{...}  (font-mono)           │  [提交]            │
│  └──────────────────────────────────┘                    │
│                                                          │
│  本轮攻击记录                                             │
│  ┌────────────────────────────────────────────────────┐  │
│  │ 14:32  TeamA  flag{awd_r12_a3f...}  ✓ +50pts     │  │
│  │ 14:28  TeamC  flag{awd_r12_c7b...}  ✗ 已过期     │  │
│  │ 14:15  TeamB  flag{awd_r12_b2e...}  ✓ +50pts     │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  本场攻击统计: 攻击成功 8 次 | 攻击得分 400 pts           │
└──────────────────────────────────────────────────────────┘
```

### 目标队伍选择

```
flex flex-wrap gap-3

每个队伍卡片:
  bg-surface border border-border rounded-lg px-4 py-3 cursor-pointer
  min-w-[100px] text-center

  队名: text-sm font-medium text-primary
  状态: 复用服务状态指示

选中态:
  border-primary bg-primary/5

宕机队伍:
  opacity-50 cursor-not-allowed (无法攻击宕机队伍)

自己队伍:
  不显示在列表中
```

### Flag 提交

```
复用 MASTER Flag 输入框样式:
  font-mono text-primary border-primary/30 focus:border-primary

提交结果:
  成功: 边框变绿 0.3s + toast "+50 pts"
  失败-Flag错误: 输入框抖动 + toast "Flag 错误"
  失败-已过期: toast "该 Flag 已过期（非当前轮次）"
  失败-重复: toast "已提交过该队伍的 Flag"

频率限制:
  触发条件: 1 分钟内错误提交 > 10 次
  按钮禁用 + text-warning text-xs 显示 "冷却中 (Xs)"
  倒计时结束后自动恢复
```

### 攻击记录

```
bg-surface border border-border rounded-lg overflow-hidden

每行: px-4 py-2 text-xs border-b border-border-subtle
  时间: font-mono text-muted w-14
  目标: text-secondary w-16
  Flag: font-mono text-muted truncate
  结果:
    成功: text-success "+50pts"
    失败: text-danger "已过期" / "错误" / "重复"
```

---

## 5. Tab：排行榜

复用 scoreboard.md 排行榜布局，增加 AWD 专属列：

```
# | 队伍 | 攻击分 | 防守分 | 总分 | 服务状态

攻击分列: text-danger font-mono (红色强调攻击属性)
防守分列: text-success font-mono (绿色强调防守属性)
总分列: font-mono font-bold

得分趋势图: 双线（攻击分 + 防守分），使用 danger/success 色
```
