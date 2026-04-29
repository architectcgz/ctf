# 学生仪表盘行动型子页改版设计

## 目标

把学生仪表盘中以下三个子页升级为统一的行动型训练工作区：

- `/student/dashboard?panel=recommendation`
- `/student/dashboard?panel=category`
- `/student/dashboard?panel=difficulty`

这轮改版要解决的核心问题是：

- 这三个 panel 目前各自拥有独立 hero、独立卡片语言和独立信息节奏，和当前 student workspace 已经采用的共享 shell 不一致
- 页面虽然展示了推荐、分类和难度数据，但第一屏没有明确回答“现在该练什么”
- recommendation、category、difficulty 三页彼此之间缺少连续叙事，看起来像三张旧页面，而不是一个完整训练工作区

最终结果应当是：三页都优先给出训练动作和进入路径，再用简洁依据解释为什么这样安排，并与当前 student dashboard 总览页保持一致的工作区语言。

## 非目标

- 不新增路由，不改变 `?panel=` 的切换方式
- 不改动 `getMyProgress / getRecommendations / getSkillProfile` 等现有学生侧数据接口
- 不重做 `overview` 与 `timeline` 两个 panel
- 不新建独立“推荐中心”或“训练计划”页面
- 不在这一轮新增复杂推荐算法或新的后端聚合字段

## 当前背景

当前 `DashboardView` 已经通过 route query tab 把 student dashboard 分成：

- `overview`
- `recommendation`
- `category`
- `timeline`
- `difficulty`

其中总览页已经具备较新的 student workspace 风格：

- 共享 `top-tabs`
- `workspace-tab-heading__title`
- `metric-panel-grid / metric-panel-card`
- `journal-soft-surface`

但 recommendation、category、difficulty 仍存在以下问题：

### recommendation

- 第一屏仍然偏“说明页”，推荐入口不够强
- “推荐摘要”和“推荐列表”是两块分离的大卡片，动作入口没有形成主次层级
- 推荐理由分散在多个块里，用户需要先读页面，再决定是否行动

### category

- 当前主视觉是“强项 / 短板”两张大卡片，更像统计页，不像训练工作台
- 分类列表按照覆盖率排序，但没有把“应该先补哪类”变成主结论
- 去题库训练的动作只停留在页面级，没有下沉到分类行

### difficulty

- 当前结构偏“难度说明页”，解释性内容比行动性内容更重
- “训练解读”和“当前突破口”分块较多，占据大量视觉权重
- 难度页能看出分布，但不够直接告诉用户下一步该把训练推到哪个难度

## 方案比较

### 方案 A：行动导向训练台

做法：

- 每个 panel 第一屏直接给出当前训练动作
- 推荐、分类、难度都把入口前置，把分析压缩为次级依据
- 页面主体以扁平目录或行动列表为主，不再以说明卡片为主

优点：

- 最符合 student workflow
- 用户能更快从 dashboard 进入训练
- recommendation、category、difficulty 三页都能形成一致叙事

缺点：

- 需要重排现有信息层级，不能只做视觉换皮
- 需要把 category / difficulty 的动作入口做实，最好顺手补齐题库筛选跳转

### 方案 B：分析导向训练分析台

做法：

- 每个 panel 先解释当前结构，再给训练建议
- recommendation 作为 category 和 difficulty 分析后的结论页

优点：

- 分析逻辑完整
- 对已经熟悉平台的用户更有解释力

缺点：

- 第一屏动作不够强
- 更像“报告”而不是“工作台”

### 方案 C：混合型工作区

做法：

- 上半区给结论和入口，下半区给较完整分析
- recommendation、category、difficulty 使用统一骨架，但保留较多说明信息

优点：

- 平衡
- 改版风险较低

缺点：

- 对当前 student dashboard 来说仍然偏保守
- 不足以彻底解决“看起来像旧页面”的问题

## 决策

采用方案 A。

这轮把 recommendation、category、difficulty 三页统一改造成“行动导向训练台”：

- 第一屏先给结论
- 第二层给可点击的训练入口
- 第三层才给简洁依据

分析信息不消失，但不再占据页面主叙事。

## 页面级设计

### 共享骨架

三个 panel 都采用同一套页面骨架：

1. 行动标题区
   - `journal-eyebrow`
   - `workspace-tab-heading__title`
   - 一段只说明当前动作目标的简短 copy
2. 摘要条
   - 使用共享 `metric-panel` 样式栈
   - 固定为 3 个 summary card
   - 只展示当前阶段最重要的结论，不再展示泛化统计
3. 主任务区
   - recommendation 使用行动列表
   - category 使用可训练分类列表
   - difficulty 使用可推进难度列表
4. 辅助区
   - 使用更紧凑的 rail 或紧凑说明块
   - 只保留“为什么这样推荐”的依据，不再扩张成第二个主页面

### 结构原则

- 不再为每个 panel 单独做一块厚重 hero 卡
- 不再堆叠多块说明型大卡片
- 主操作始终在视觉上强于说明块
- 说明内容只解释当前动作，不写泛化设计文案

## recommendation 设计

### 页面定位

回答：现在先练什么题。

### 标题区

- 标题改为行动导向，例如“现在先练这几道”
- copy 只说明“这些题和你当前薄弱方向最相关”

### 摘要条

保留 3 个摘要项：

- 首要补强分类
- 当前目标难度
- 推荐队列数量

helper 文案必须解释指标意义，例如“当前薄弱维度里优先级最高”，而不是只有单位。

### 主任务区

主体改为平铺推荐列表：

- 每一行展示序号、题目标题、简短理由、分类/难度标签
- 主动作按钮直接进入 challenge detail
- 第一项视觉上略强，其余保持同一目录节奏

### 辅助区

只保留一个紧凑依据块，解释：

- 为什么当前优先补这个分类
- 为什么目标难度是这一档
- 如果列表为空，下一步应该去题库探索

空态要显式提示，并保留“浏览全部题目”的入口。

## category 设计

### 页面定位

回答：现在优先补哪个分类。

### 标题区

- 标题直接给出当前优先补强方向，而不是泛称“分类覆盖概况”
- copy 明确这是“按当前训练价值排序”的分类清单

### 摘要条

保留 3 个摘要项：

- 当前优先分类
- 当前最稳定分类
- 总体分类覆盖率

### 主任务区

主体改为可训练分类列表，而不是“强项 / 短板”大卡：

- 每行展示分类名、完成率、已解 / 总量
- 每行保留一个行动按钮，例如“去练这个分类”
- 列表排序以“当前最值得补”优先，而不是单纯高到低

推荐排序规则：

- 先按覆盖率低
- 再参考该分类总题量，避免极小样本误导
- 最终目的是让短板分类排在更前

### 辅助区

辅助区只保留一个简洁快看模块：

- 展示 2 到 3 个高层结论
- 例如“当前短板为何是 Crypto”“为什么 Web 暂时不需要优先投入”

不再使用两张并列 highlight card 作为主结构。

## difficulty 设计

### 页面定位

回答：下一步该把训练强度推到哪一层。

### 标题区

- 标题直接给出当前建议主攻难度
- copy 明确说明这是“训练强度推进建议”

### 摘要条

保留 3 个摘要项：

- 当前建议主攻难度
- 当前最稳定难度层
- 整体难度覆盖率

### 主任务区

主体仍然展示难度阶梯，但从纯统计展示改成“可推进列表”：

- 每层难度展示当前覆盖率和已解 / 总量
- 当前建议突破的难度行需要更明确的强调
- 行尾保留一个入口，例如“去练这档题”

### 辅助区

现有“训练解读”压缩为小型提示块：

- 是否长期停留在舒适区
- 当前应该先补哪一层再继续上推

不再用整块大卡承载长段说明。

## 跳转设计

为了让行动导向真正成立，这轮顺手补齐 dashboard 到题库页的筛选跳转。

### 目标行为

- recommendation 的“浏览全部”继续进入通用题库页
- category 行内按钮可跳转到带 `category` 预设筛选的题库页
- difficulty 行内按钮可跳转到带 `difficulty` 预设筛选的题库页

### 配套要求

`ChallengeList` 当前虽然支持分类和难度筛选，但没有把筛选状态和 URL query 同步。

因此这轮需要补上：

- 从 route query 初始化 `categoryFilter / difficultyFilter`
- 用户在题库页变更筛选时同步更新 query
- dashboard 跳转到题库时可以直接落到正确筛选结果

这样 category / difficulty 页上的行动按钮才不是“看起来能用”，而是真正把用户带到对应训练集合。

## 视觉与一致性约束

这轮 UI 必须继续遵守当前 student workspace 语言：

- 继续使用 `journal-soft-surface`
- 面板标题沿用 `workspace-tab-heading__title`
- 摘要条使用完整的 `metric-panel-grid / metric-panel-card / metric-panel-label / metric-panel-value / metric-panel-helper`
- 主列表优先使用 flat row，不再做宫格式卡片堆叠
- 主按钮强调、次按钮中性，保持现有 student 操作层级
- 空态使用显式反馈，不渲染空白区域

## 数据与实现边界

这轮不改后端数据结构，前端只基于现有数据做重排与转译：

- recommendation 继续使用 `recommendations` 和 `weakDimensions`
- category 继续使用 `progress.category_stats`
- difficulty 继续使用 `progress.difficulty_stats`

必要的前端新增逻辑仅限：

- 计算当前优先分类
- 计算当前建议主攻难度
- category / difficulty 的行动排序
- dashboard 到 challenge list 的 query 跳转支持

## 测试策略

至少覆盖以下三类验证：

1. student dashboard 行为测试
   - recommendation / category / difficulty 三个 panel 仍然通过 `?panel=` 正确显示
   - 新的标题、摘要条和主任务区能被渲染
   - category / difficulty 的行动按钮能按预期触发路由跳转
2. 共享样式与结构回归测试
   - 三个 panel 采用统一的 `metric-panel` 样式栈
   - 没有重新引入旧式硬编码边框或各页各自的卡片表面
3. 题库筛选联动测试
   - challenge list 能从 query 初始化分类和难度筛选
   - dashboard 行动按钮跳转后，题库页能落到对应筛选状态

## 验收标准

完成后应满足：

1. recommendation 页第一屏就能看见推荐题目和直接进入训练的动作入口
2. category 页不再以两张大卡为主结构，而是以“优先补哪个分类”的行动列表为主
3. difficulty 页不再以说明卡为主，而是明确告诉用户下一步该推进到哪个难度层
4. 三个 panel 在标题、摘要条、列表节奏和辅助区语言上形成统一工作区风格
5. category / difficulty 的行动按钮能把用户带到带预设筛选的题库页
6. 现有 `overview / timeline`、tabs、数据接口和 student dashboard 路由行为不被破坏
