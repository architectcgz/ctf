# CTF 卡片设计系统

> 适用范围：`code/frontend` 全部页面  
> 目标：统一卡片的信息层级、视觉语言和交互边界，避免同一页面同时出现“主视觉卡”“普通统计卡”“临时拼装卡”三套风格。

## 1. 设计结论

当前前端卡片的主要问题不是“单张不好看”，而是：

- 主导卡与普通卡没有稳定的层级边界
- 很多页面直接手写 `rounded + border + gradient + shadow`，导致视觉语言漂移
- 同样是“摘要卡”，有的像 hero，有的像 table cell，有的像 panel

后续统一采用这 4 类卡片：

1. `Hero Card`
用于页面里最重要、最具指向性的摘要或行动入口。

2. `Metric Card`
用于展示单一核心指标，不承担复杂叙事。

3. `Panel Card`
用于承载列表、图表、正文区块、分析区块，是默认卡片。

4. `Action Card`
用于单个动作、建议、跳转或状态判断，不用于堆大段内容。

## 2. 使用规则

### Hero Card

适用：

- 页面首屏主摘要
- 推荐策略、训练计划、总体覆盖概况
- 需要“先看这一张”的场景

限制：

- 同一视口同一层级最多 1 张
- 不允许连续堆 2-3 张 hero
- 不允许把普通统计卡伪装成 hero

结构：

```text
┌────────────────────────────────────────────┐
│ Eyebrow                                   │
│ Title                                     │
│ Description                               │
│                                            │
│ Supporting block / chips / CTA            │
└────────────────────────────────────────────┘
```

视觉规则：

- 背景使用 `surface` 基底 + 单一 accent 方向性染色
- 允许渐变，但必须是“轻染色”，不能做高饱和大面积炫光
- 标题、描述、CTA 必须形成清晰三层
- 内部子块继续回到 `surface/base` 体系，不再叠第二套 hero 背景

### Metric Card

适用：

- 单个数字或短状态
- 排名、完成率、总得分、最强方向、最弱方向

限制：

- 一张卡只表达一个指标
- 不塞长段解释
- 不使用大面积 hero 渐变

结构：

```text
┌──────────────────────────┐
│ Label              Accent│
│ Value                    │
│ Hint                     │
└──────────────────────────┘
```

视觉规则：

- 统一使用 `surface` 底
- 类型差异只通过 accent 条、icon chip、hint 辅助表达
- Metric 卡之间必须保持同高或近似节奏

### Panel Card

适用：

- 列表、图表、时间线、信息区块
- 多项内容的容器

结构：

```text
┌────────────────────────────────────────────┐
│ Title / Subtitle                 Actions   │
├────────────────────────────────────────────┤
│ Content                                    │
└────────────────────────────────────────────┘
```

视觉规则：

- 这是默认卡片，不主动抢视觉焦点
- 强调内容本身，不强调卡片外壳
- 标题区必须稳定，方便页面扫描

### Action Card

适用：

- 建议动作
- 单条跳转入口
- 结构判断、风险提示、下一步推荐

视觉规则：

- 比 panel 更紧凑
- 比 metric 更强调点击或动作结果
- hover 只允许轻微位移和边框增强

## 3. 统一 Token

### 圆角

- Hero: `28px - 30px`
- Panel / Metric / Action: `22px - 26px`
- 内部小块：`18px - 22px`

### 阴影

- 默认：`0 18px 40px var(--color-shadow-soft)`
- Hero：`0 24px 64px var(--color-shadow-soft)`
- 不允许叠多层重阴影

### 边框

- 默认：`border-border`
- 强调：`accent 22% + border-default mix`
- 不允许纯白边框大面积出现

### 色彩

- 主强调：`var(--color-primary)`
- 成功：`var(--color-success)`
- 警告：`var(--color-warning)`
- 错误：`var(--color-danger)`
- `violet` 仅作少量辅助场景，不作为全局主卡色

## 4. 统一标题层级

- `Eyebrow`: `11px / uppercase / tracking 0.18em - 0.24em`
- `Card Title`: `18px - 32px`，取决于 card variant
- `Description`: `14px - 15px / line-height 1.65 - 1.8`
- `Hint`: `12px - 14px`

## 5. 当前问题判断

### 为什么 `Targeted Training` 成立

- 它是明确的 `Hero Card`
- 有 Eyebrow、主标题、解释文案、支持块、CTA
- 内部信息围绕一个目标组织：补短板

### 为什么 `覆盖概况` 不成立

- 它承担的是 Hero 级别的信息，但被画成了 Metric 样式
- 标题、数值、说明之间缺少真正的主次关系
- 它和下面的“最强方向/最弱方向”没有形成一个稳定的卡片家族

正确做法：

- `覆盖概况` 升级为 `Hero Card`
- `最强方向 / 最弱方向` 统一为 `Metric Card`
- `分类进度板` 继续作为 `Panel Card`

## 6. 禁用项

- 禁止在页面里继续手写新的卡片外壳
- 禁止用不同页面各自的渐变体系表达同一种卡片角色
- 禁止把一张卡同时当 hero、metric、panel 三种角色使用
- 禁止为了“显得丰富”给每张卡都加大面积彩色背景

## 7. 前端实现要求

- 公共卡片底座统一走 `AppCard`
- `SectionCard` 是 `Panel Card` 包装器
- `MetricCard` 是 `Metric Card` 包装器
- 页面中的主导卡直接使用 `AppCard variant="hero"`
- 后续新增卡片默认先判断属于哪一类，再决定是否新增变体

## 8. 本轮落地范围

- 学员端 `训练建议` 页
- 学员端 `分类进度` 页
- 公共组件：`AppCard`、`SectionCard`、`MetricCard`

## 9. 后续默认规则

后续在 CTF 前端编写卡片时，默认遵循这份文档；如果某个页面需要偏离，必须先说明它偏离的是哪条规则，以及为什么必须偏离。
