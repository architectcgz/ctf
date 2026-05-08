# 页面设计：学生仪表盘 (Dashboard)

> 继承：../design-system/MASTER.md | 角色：学生
> 路由：`/student/dashboard?panel={overview|recommendation|category|timeline|difficulty}`

---

## 页面定位

- 学生仪表盘是训练工作区，不再是单纯的数据总览页。
- 页面使用统一的 `journal-shell user workspace` 外壳，所有 panel 通过 query tab 切换。
- `overview` 和 `timeline` 负责总览与记录，`recommendation`、`category`、`difficulty` 负责把“下一步练什么”直接落成行动入口。

---

## 1. 页面骨架

根页面 `DashboardView.vue` 只承接三类职责：

- 顶部 `top-tabs`
  - `训练总览`
  - `训练队列`
  - `分类补强`
  - `训练记录`
  - `强度推进`
- 根级错误提示和 loading skeleton
- 通过 `dashboardPanelRegistry` 和 `resolveDashboardPanelBindings()` 挂载当前 panel

切换规则：

- 当前激活 panel 由 `?panel=` 驱动。
- tab 支持键盘切换和 aria 语义。
- 非当前 panel 仍保留挂载入口，但只显示激活项。

---

## 2. 训练队列 (`recommendation`)

这一页回答：现在先练哪几道。

页面结构：

1. 标题区
   - overline 使用 `Action Queue`
   - 主标题固定为 `现在先练这几道`
   - 标题下方显示当前薄弱维度 pill；没有明显短板时显示稳定态标签

2. 摘要条
   - 固定三张 summary card
   - `当前补强方向`
   - `当前目标难度`
   - `当前行动队列`

3. 主任务区
   - 有推荐数据时显示行动列表
   - 每项展示排序号、题目名、难度标签、分类标签和推荐理由
   - 点击整行直接进入题目详情

4. 工具动作
   - `能力画像`
   - `浏览全部题目`

空状态规则：

- 不保留多余说明卡片。
- 只显示一句提示和唯一主 CTA `浏览全部题目`。

---

## 3. 分类补强 (`category`)

这一页回答：现在优先补哪个分类。

页面结构：

1. 标题区
   - 标题直接写成 `优先补这个分类：{category}`
   - 没有统计数据时退化为“先开始积累分类覆盖面”

2. 摘要条
   - `当前待补题量`
   - `整体覆盖率`
   - `当前排序依据`

3. 快捷动作
   - `先去 {分类}` 或通用 `去训练`
   - `浏览全部题目`
   - `能力画像`

4. 分类行动列表
   - 每行展示排名、分类名、完成率、已解 / 总量
   - 附带当前这一类为什么排到这里的解释文案
   - 行尾 CTA 固定为 `去做这个分类`
   - 下方进度条表达该分类当前完成率

排序规则：

- 先按完成率低的分类优先。
- 再用题量打破并列，避免极小样本直接排到最前。

---

## 4. 强度推进 (`difficulty`)

这一页回答：下一步该把训练强度推到哪一档。

页面结构：

1. 标题区
   - 标题直接写成 `先推这一档强度：{difficulty}`
   - 没有难度数据时退化为“先开始建立强度节奏”

2. 摘要条
   - `当前完成率`
   - `当前覆盖层级`
   - `推进节奏`

3. 快捷动作
   - `先做{难度}`
   - `浏览全部题目`

4. 难度行动列表
   - 固定按难度层级顺序展示
   - 当前主推档位高亮
   - 每项展示完成率、已解 / 总量、解释文案和进度条
   - 行尾 CTA 为 `去做这一档`

视觉规则：

- 不再用“说明型大卡片”解释训练哲学。
- 强度推进的主结论必须在第一屏可见。

---

## 5. 共享动作出口

- 点击推荐题目：进入题目详情页
- 点击 `浏览全部题目`：进入题库目录
- 点击分类 CTA：带 `category` query 进入题库
- 点击难度 CTA：带 `difficulty` query 进入题库
- 点击 `能力画像`：进入能力画像页

三个行动型 panel 都复用 `journal-soft-surface`、`metric-panel` 和学生侧 workspace 字体 / 间距系统，不再各自维护独立 hero 语言。

---

## 代码落点

- 根页面：
  - `code/frontend/src/views/dashboard/DashboardView.vue`
  - `code/frontend/src/features/student-dashboard/model/useStudentDashboardPage.ts`
  - `code/frontend/src/features/student-dashboard/model/useStudentDashboardPanelBindings.ts`
- 行动型 panel：
  - `code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue`
  - `code/frontend/src/components/dashboard/student/StudentCategoryProgressPage.vue`
  - `code/frontend/src/components/dashboard/student/StudentDifficultyPage.vue`
