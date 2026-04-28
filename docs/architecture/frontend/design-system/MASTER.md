# CTF 网络攻防靶场平台 — Design System Master

> 项目：CTF 网络攻防靶场平台（校园教学实训）
> 技术栈：Vue 3 + Tailwind CSS + ECharts
> 风格：Clean Dark + 技术感点缀
> 模式：Dark Mode 默认

---

## 1. 设计风格

**主风格：Clean Dark**

校园教学实训平台，核心用户是学生和教师。设计以信息清晰、操作高效为第一优先级，在此基础上通过色彩和排版体现安全/技术氛围。

- 深色背景（非纯黑），减少长时间使用的视觉疲劳
- 卡片使用实色背景 + 细边框，层级分明
- 等宽字体仅用于 Flag 输入、代码块、数字指标
- 动画克制，仅用于操作反馈（Flag 提交结果、实例状态变更）

**技术感点缀（适度）：**
- Flag 输入框使用等宽字体 + 主色边框，暗示终端输入
- 排行榜排名变动时使用简洁的位移动画
- 登录页可使用轻量几何网格背景

**反模式（避免）：**
- 不使用扫描线、霓虹光晕、矩阵雨等装饰性动画
- 不使用 glassmorphism（深色背景上效果差）
- 不使用 emoji 作为图标，统一使用 Lucide Icons
- 不使用 scale 变换的 hover 效果（避免布局抖动）
- 不使用纯黑 `#000` 背景

---

## 2. 色彩系统

### 2.1 基础色板

```
背景层级：
  bg-base:       #0f1117    -- 最底层背景（页面），接近 GitHub Dark
  bg-surface:    #161b22    -- 卡片/面板背景
  bg-elevated:   #1c2128    -- 弹窗/下拉菜单/hover 行
  bg-overlay:    #0f1117/80 -- 遮罩层

边框：
  border-default: #30363d   -- 默认边框（可见但不抢眼）
  border-subtle:  #21262d   -- 弱边框（区域分隔）
  border-active:  #0891b2   -- 选中/聚焦态边框

文字层级：
  text-primary:   #e6edf3   -- 主要文字
  text-secondary: #8b949e   -- 次要文字
  text-muted:     #6e7681   -- 辅助/禁用文字
  text-inverse:   #0f1117   -- 按钮内反色文字
```

### 2.2 主题色

```
Teal（主色 / 交互）：
  primary:       #0891b2    -- 按钮、链接、选中态（cyan-600，比 cyan-500 更沉稳）
  primary-hover: #06b6d4    -- hover 态（cyan-500）
  primary-dim:   #0891b2/15 -- 背景高亮

Violet（强调）：
  accent:        #8b5cf6    -- 首杀标识、特殊成就
  accent-dim:    #8b5cf6/15

Green（成功）：
  success:       #22c55e    -- Flag 正确、已解出
  success-dim:   #22c55e/10

Amber（警告）：
  warning:       #f59e0b    -- 即将过期、注意事项
  warning-dim:   #f59e0b/10

Red（错误 / 危险）：
  danger:        #ef4444    -- Flag 错误、删除操作
  danger-dim:    #ef4444/10
```

### 2.3 难度色标

避免与分类色标冲突，使用独立色值：

```
入门:  #34d399  (emerald-400)
简单:  #38bdf8  (sky-400)
中等:  #fbbf24  (amber-400)
困难:  #f87171  (red-400)
地狱:  #c084fc  (purple-400)
```

### 2.4 分类色标（靶场类别）

```
Web:       #06b6d4  (cyan-500)
Crypto:    #a78bfa  (violet-400)
Pwn:       #ef4444  (red-500)
Reverse:   #fb923c  (orange-400)
Misc:      #22c55e  (green-500)
Forensics: #3b82f6  (blue-500)
```

---

## 3. 字体系统

### 3.1 字体配对

```
标题/正文:  "Inter", "Noto Sans SC", system-ui, sans-serif
  -- 所有标题和正文统一使用无衬线字体，中英文混排友好

代码/数字:  "JetBrains Mono", "Cascadia Code", ui-monospace, monospace
  -- 仅用于：Flag 输入框、代码块、倒计时、分数数字、排名数字
```

### 3.2 字号规范

```
text-xs:    12px / 1.5   -- 标签、辅助信息
text-sm:    14px / 1.5   -- 次要文字、表格内容
text-base:  16px / 1.6   -- 正文
text-lg:    18px / 1.6   -- 小标题
text-xl:    20px / 1.5   -- 卡片标题
text-2xl:   24px / 1.4   -- 区块标题
text-3xl:   30px / 1.3   -- 页面标题
text-4xl:   36px / 1.2   -- Dashboard 核心指标数字（font-mono）
```

---

## 4. 间距与布局

### 4.1 间距系统

```
基础单位: 4px (Tailwind 默认)
常用间距:
  gap-1:  4px    -- 图标与文字间距
  gap-2:  8px    -- 紧凑元素间距
  gap-3:  12px   -- 表单元素间距
  gap-4:  16px   -- 卡片内边距
  gap-6:  24px   -- 区块间距
  gap-8:  32px   -- 大区块间距
```

### 4.2 布局规范

```
最大内容宽度:  max-w-7xl (1280px)
顶部导航高度:  h-14 (56px)
侧边栏:       默认折叠 w-16 (64px，仅图标)，展开 w-56 (224px)
卡片圆角:     rounded-lg (8px)
卡片边框:     border border-[#30363d]
```

### 4.3 响应式断点

```
md:   768px   -- 平板（侧边栏隐藏，汉堡菜单）
lg:   1024px  -- 桌面（侧边栏折叠态）
xl:   1280px  -- 大桌面
```

移动端为次要场景（校园网 + 电脑操作靶机），优先保证 lg 以上体验。

---

## 5. 组件规范

### 5.1 按钮

```
主按钮:
  bg-[#0891b2] hover:bg-[#06b6d4] text-white font-medium
  px-4 py-2 rounded-lg transition-colors duration-150
  cursor-pointer

次要按钮:
  bg-[#21262d] hover:bg-[#30363d] text-[#e6edf3] border border-[#30363d]
  px-4 py-2 rounded-lg transition-colors duration-150
  cursor-pointer

危险按钮:
  bg-[#ef4444]/10 hover:bg-[#ef4444]/20 text-[#f87171] border border-[#ef4444]/20
  px-4 py-2 rounded-lg transition-colors duration-150
  cursor-pointer

禁用态:
  opacity-50 cursor-not-allowed pointer-events-none

尺寸:
  sm: px-3 py-1.5 text-sm
  md: px-4 py-2 text-sm (默认)
  lg: px-5 py-2.5 text-base
```

### 5.2 卡片

```
基础卡片:
  bg-[#161b22] border border-[#30363d] rounded-lg p-5

可点击卡片:
  基础卡片 + hover:border-[#0891b2]/50 transition-colors duration-150 cursor-pointer

高亮卡片（已解出/活跃）:
  bg-[#161b22] border border-green-500/30

统计卡片:
  bg-[#161b22] border border-[#30363d] rounded-lg p-5
  数字使用 font-mono text-3xl font-bold
```

### 5.3 输入框

```
基础输入:
  bg-[#0f1117] border border-[#30363d] text-[#e6edf3]
  placeholder:text-[#6e7681]
  rounded-lg px-3 py-2 text-sm
  focus:border-[#0891b2] focus:ring-1 focus:ring-[#0891b2]/50
  transition-colors duration-150

Flag 输入框:
  基础输入 + font-mono text-[#0891b2]
  border-[#0891b2]/30 focus:border-[#0891b2]
```

### 5.4 标签/徽章

```
难度标签:
  px-2 py-0.5 rounded text-xs font-medium
  bg-{难度色}/10 text-{难度色}

分类标签:
  px-2 py-0.5 rounded text-xs font-medium
  bg-{分类色}/10 text-{分类色}

状态标签:
  已解出: bg-green-500/10 text-green-400
  进行中: bg-[#0891b2]/10 text-[#06b6d4]
  未开始: bg-[#30363d] text-[#8b949e]
```

### 5.5 表格

```
表头:
  bg-[#161b22] text-[#8b949e] text-xs font-medium uppercase tracking-wider
  px-4 py-3 border-b border-[#30363d] text-left

表行:
  text-[#e6edf3] text-sm px-4 py-3 border-b border-[#21262d]
  hover:bg-[#1c2128] transition-colors duration-100

排名高亮行:
  第1名: bg-amber-500/5 border-l-2 border-amber-400
  第2名: bg-slate-300/5 border-l-2 border-slate-400
  第3名: bg-orange-400/5 border-l-2 border-orange-400
```

### 5.6 Toast / 操作反馈

```
成功: bg-green-500/10 border border-green-500/20 text-green-400
错误: bg-red-500/10 border border-red-500/20 text-red-400
警告: bg-amber-500/10 border border-amber-500/20 text-amber-400
信息: bg-[#0891b2]/10 border border-[#0891b2]/20 text-[#06b6d4]

位置: 右上角固定，top-4 right-4
动画: fade-in + slide-left，3s 后自动消失
```

### 5.7 空状态

```
居中布局:
  [Lucide 图标 48px, text-[#6e7681]]
  [主文案 text-[#e6edf3] text-base]
  [副文案 text-[#8b949e] text-sm]
  [可选：引导按钮]
```

### 5.8 加载态

```
页面级: 居中 spinner（Lucide Loader2 旋转）
区块级: skeleton 占位条，bg-[#21262d] rounded animate-pulse
按钮级: 按钮内 spinner 替换文字，按钮禁用
```

---

## 6. 图标规范

- 图标库：Lucide Icons（统一风格，支持 Vue 组件）
- 图标尺寸：16px (inline) / 20px (按钮内) / 24px (导航) / 48px (空状态)
- 图标颜色：继承父元素文字颜色，或使用主题色
- 禁止使用 emoji 作为功能图标

---

## 7. 动画规范

```
微交互（hover/focus/状态切换）:
  duration: 150ms
  easing: ease-in-out
  属性: color, background-color, border-color, opacity

排行榜排名变动:
  duration: 300ms
  easing: ease-in-out
  方式: translateY 位移

Flag 提交反馈:
  正确: 输入框边框变绿 0.3s + 得分数字 fade-in
  错误: 输入框水平抖动 0.3s (translateX ±4px)

prefers-reduced-motion:
  所有动画降级为 instant 切换
```

---

## 8. 可访问性

- 文字对比度 ≥ 4.5:1（WCAG AA）
- 所有交互元素有 focus-visible 样式：`ring-2 ring-[#0891b2] ring-offset-2 ring-offset-[#0f1117]`
- 图标按钮必须有 aria-label
- 表单输入必须关联 label（可视觉隐藏但 DOM 存在）
- 颜色不作为唯一信息传达方式（难度同时用文字 + 颜色）
- Tab 顺序与视觉顺序一致
- 触摸目标 ≥ 44x44px
