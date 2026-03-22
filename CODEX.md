# CTF 项目 Codex 规范

## 前端 UI 规范（强制）

### 去卡片化优先

- 默认采用去卡片化设计，**尽量不使用卡片式容器**（大圆角、重阴影、封闭块状面板）。
- 页面结构优先使用：分区标题 + 分隔线 + 列表/表格/栅格，不要堆叠卡片。
- 信息强调优先使用：左侧强调线、字号层级、间距层级、轻量背景，而不是卡片边框与阴影。

### 例外条件（必须满足其一）

- 只有在以下情况才允许局部使用卡片：
  - 独立浮层（如弹窗、下拉面板）需要与页面主层明确分离。
  - 明确的交互块需要整体点击且必须形成视觉边界。
  - 现有组件库能力限制，短期无法替换，且已记录后续去卡片化计划。

### 评审要求

- 新增或改造前端页面时，若使用卡片化样式，必须在变更说明中写明原因。
- Code review 默认以“去卡片化”为通过标准之一。

## 项目 UI 主题整体提示词

### 主题方向

- Modern Tech-Academic UI
- Professional SaaS layout
- Refined Dark/Light mode
- solid surfaces with restrained layering
- UI/UX for developers
- 1px border strokes
- Monospace font accents
- muted Slate and Indigo palette
- minimalist icons with soft glow hover
- organized navigation groups
- non-distractive hierarchy

### 可读性与背景约束（强制）

- 校内办公 / 学习场景优先高可读性，禁止为了“高级感”牺牲清晰度。
- 主内容区、主导航区、顶栏、常驻信息容器默认使用实色背景，不使用 `backdrop-filter: blur(...)`。
- 浮层（如下拉、弹窗、Toast）也应优先使用实色背景 + 1px 边框；只有确有必要时才允许极弱透明度，但仍禁止明显雾化。
- 浅色主题下禁止使用会形成“蒙雾感”的半透明白层、低对比灰层、叠色网格或大面积模糊背景。
- 层级区分优先级：
  - 实色背景
  - 1px 细边框
  - 轻阴影
  - 间距与字号层级
  - 最后才是轻量色块
- 禁止把 `Glassmorphism`、`backdrop blur`、`frosted`、`translucent overlay` 用作主工作区背景方案。
- 若需要现代感，优先使用：纯净背景、精细描边、微弱渐变、轻量光晕，而不是模糊。

### 结构与细节优化提示词

#### A. 品牌与顶部区域

Prompt: `Sidebar top branding: Compact school logo placeholder, platform name in semi-bold sans-serif combined with monospace version tag (v2.0), clean bottom divider.`

- 优化点：增加“版本号”或“系统名称”，用等宽字体（Monospace）显示，能提升工具感。

#### B. 导航项交互

Prompt: `Navigation item: Ghost button style, 8px rounded corners, Lucide-style thin line icons, active state with vertical accent bar on the left, subtle indigo background tint on hover, transition ease-in-out.`

- 优化点：强调“左侧激活条”和“圆角矩形”，优先于单纯文字变色。

#### C. 信息反馈与徽章

Prompt: `Miniature data badges in sidebar: small pill-shaped tags for "New" challenges or "Live" events, using desaturated colors (soft emerald, soft amber), high contrast text.`

- 优化点：在侧边栏加入低噪音实时状态，例如新题、Live 事件、小圆点或胶囊标签。

### 综合性参考 Prompt

Prompt:

`UI design of a sidebar for a CTF platform, internal school use. Clean and sophisticated. Color palette: Deep Charcoal (#121212) or Soft White (#F9FAFB). Features: Top brand area, main navigation list (Dashboard, Challenges, Leaderboard, Team, Docs), bottom user profile section with avatar and online status dot. Elements: Use thin line icons, "Inter" font for UI, "JetBrains Mono" for stats. Minimalist, high readability, professional engineering vibe. No neon, no flashy animations.`

### 禁止性 Prompt 纠偏

- 不要写：`Glassmorphism`, `backdrop blur`, `frosted panel`, `misty overlay`, `translucent workspace`
- 应改写为：
  - `solid background panels with crisp 1px borders`
  - `high readability, no foggy overlays`
  - `clean school-office UI, sharp surfaces, no blur`
  - `light theme with clear separation between background and content`
