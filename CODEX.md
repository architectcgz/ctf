# CTF 项目 Codex 规范

## 毕设课题约束（强制）

项目范围必须与 [docs/毕业设计课题.md](/home/azhi/workspace/projects/ctf/docs/毕业设计课题.md) 保持一致。后续涉及功能设计、架构取舍、页面改造、导入规范和答辩口径时，默认以该课题描述为项目上位约束，不得偏离。

课题原始要求摘要：

- 平台定位：
  - 网络攻防靶场用于网络安全技术研究与校赛人才选拔
  - 通过综合云管理平台对教学实验、综合实战、仿真实战、攻防对抗进行集中监控与管理
  - 使用 Docker 技术快速生成靶机，为信息安全专业提供训练与比赛环境
- 核心功能模块：
  - 靶场管理：管理员可配置多样化漏洞环境，支持自定义拓扑与难度分级
  - 攻防演练：学员选择靶场后获取攻击目标，系统自动记录攻击步骤与漏洞利用过程，实时反馈得分
  - 竞赛管理：支持团队对抗，自动监控攻击流量、统计防守成功次数、生成实时排行榜
  - 技能评估：基于实训数据生成能力画像，推荐针对性靶场练习，导出实训报告供教学复盘

由此导出的设计约束：

- 题目管理不能只停留在“题目元数据录入”，必须服务于 Docker 靶机生成与靶场运行时配置。
- 靶场管理能力应优先覆盖：
  - 题目元数据
  - Flag 配置
  - 附件分发
  - 运行时镜像/容器信息
  - 拓扑能力的后续接入边界
- 若做“题目导入”或“题目包规范”改造，默认目标不是单纯替换表单，而是收敛到可复用、可迁移、可部署的题目包模型。
- 涉及范围取舍时，优先保证：
  - 业务闭环完整
  - 结果稳定可演示
  - 表述适合毕业设计答辩
  - 不为了过度工程化牺牲实现完成度

## 后端架构规范（强制）

后端开发默认参考并遵守以下文档：

1. `docs/architecture/01-backend-architecture-style-decision.md`
2. `/home/azhi/workspace/docs/go-code-style-guide.md`
3. `docs/architecture/02-backend-code-style-guide.md`

如果改动涉及：

- 模块拆分
- 读写分离
- composition 装配
- backend Makefile / 校验命令
- 跨模块 contracts
- handler / service / repository 重构

那么必须先阅读以上三份文档，再开始实现。

### 当前后端目标形态

CTF 后端不是多服务单仓，但要以 workspace 共享 Go 规范为基线，并在此基础上固定采用：

- `模块化单体`
- `Clean-ish`
- `DDD-lite`
- `CQRS-lite`

这里的“服务边界”在 CTF 中统一映射成“模块边界”，不是物理微服务边界。

### 不允许的后端方向

后端当前阶段不允许：

- 把模块重新揉成全局 `internal/domain` / `internal/app` / `internal/infra` 大分层
- 引入全局 CommandBus / QueryBus
- 引入全局 Generic Repository
- 为了抽象而把所有实现都套接口
- 继续无边界扩写单个超大 `service.go`

### 新增与重构模块的默认结构

新增模块或较大重构时，优先采用：

```text
internal/module/<name>/
  contracts.go
  api/http/
  application/commands/
  application/queries/
  domain/
  ports/
  infrastructure/
  runtime/
```

约束：

- 不再新增根包兼容壳或 `module.go` 转发层
- `runtime/` 是模块内唯一允许做 wiring 的位置；如果模块暂时不需要装配层，可以不建
- 旧平铺模块可以渐进迁移，但新提取的 readmodel / query 模块不应再回退到大平铺模式

### 后端评审硬约束

Code review 默认检查：

1. 业务规则是否被塞进 handler
2. 事务边界是否散落在 HTTP 层
3. 是否跨模块依赖了对方内部实现
4. 查询聚合逻辑是否该抽成 readmodel 却仍留在命令模块
5. cache key、TTL、外部依赖是否显式且集中
6. 是否继续制造新的根包兼容层或跨模块 concrete 依赖

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
