# 页面设计：教师端夜间模式 (Teacher Dark)

> 继承：../design-system/MASTER.md | 角色：教师
> 当前覆盖：
> - `/academy/classes`
> - `/academy/students`
> - `/academy/instances`
> - `/academy/awd-reviews`
> - `/academy/classes/:className/students/:studentId/review-archive`

---

## 页面定位

- 教师端夜间模式不是单页样式补丁，而是一套共享的 `teacher-surface` 工作区表面。
- 目标是让目录页、筛选区、表格、对话框、复盘页在暗色主题下保持同源层级，不再出现白底泄漏和组件表面漂移。
- 当前共享事实以 `teacher-surface.css`、各页面 workspace 和复盘组件树为准，不再从单个页面局部样式反推整体规则。

---

## 1. 共享表面体系

教师端夜间模式统一依赖以下语义层：

- `teacher-management-shell`
- `teacher-surface`
- `teacher-surface-section`
- `teacher-surface-table`
- `teacher-surface-filter`
- `teacher-surface-error`
- `teacher-surface-dialog`

核心 token：

- `--journal-ink`
- `--journal-muted`
- `--journal-border`
- `--journal-surface`
- `--journal-surface-subtle`
- `--journal-accent`

这一层负责解决两类问题：

1. 页面自己的 hero、summary、section、directory card 保持统一深色表面。
2. `ElTable`、输入框、下拉、textarea、dialog 等内部 wrapper 继续落在同一套暗面上。

---

## 2. 目录型页面

### `/academy/classes`

- 顶部保留教师工作区标题、`教学概览` 与 `导出班级报告` 动作。
- 中部先显示目录摘要卡，再进入 `班级目录`。
- 目录区使用 `WorkspaceDirectoryToolbar + WorkspaceDataTable + WorkspaceDirectoryPagination`。
- 表格行操作只保留进入班级，不再引入亮底卡片式二级区域。

### `/academy/students`

- 页面结构和班级目录保持同一节奏：
  - 顶部标题与上下文动作
  - 摘要卡
  - 学生目录 section
- 筛选由班级切换和搜索组成，仍落在暗色目录区内部。
- 结果表格展示学号、姓名、昵称、薄弱项、做题数、得分和进入学员分析动作。

### `/academy/instances`

- 顶部只保留返回教学概览动作。
- 摘要区突出 `当前可见`、`运行中`、`即将到期` 三类指标。
- 实例目录使用紧凑过滤区而不是白色后台表单。
- 状态 chip、剩余时间和销毁动作都在同一层暗面内完成表达。

---

## 3. 复盘与评阅页面

### `/academy/awd-reviews`

- 页面使用独立的 `TeacherAWDReviewIndexWorkspace`，但仍承接同一套教师暗色工作区语义。
- 结构为：
  - workspace header
  - summary panel
  - contest directory
- 顶部动作只保留 `返回教学概览` 和 `刷新目录`。

### `/academy/classes/:className/students/:studentId/review-archive`

- 页面 owner 保持 `TeacherStudentReviewArchive.vue + TeacherReviewArchiveWorkspace.vue`。
- 页面主结构固定为：
  - `ReviewArchiveHero`
  - `TeacherReviewArchiveState`
  - `ReviewArchiveObservationStrip`
  - `TeacherReviewArchiveSummarySection`
  - `ReviewArchiveEvidencePanel`
  - `ReviewArchiveReflectionPanel`
- 导出动作继续在当前上下文完成，不回退到独立亮底报表页。

#### Hero 与总览

- `ReviewArchiveHero` 继续作为首屏入口，承接：
  - 当前学员身份信息
  - `完成题目`
  - `总提交`
  - `证据事件`
  - `复盘材料`
  - `返回学生列表`
  - `返回学员分析`
  - `导出复盘归档`
- `TeacherReviewArchiveSummarySection` 固定为双栏：
  - `训练摘要`
  - `能力画像`

#### 结构化案例区

- `ReviewArchiveEvidencePanel` 已不再使用旧的“左侧训练时间线 + 右侧证据链”双日志布局。
- 当前主体改为上下堆叠的两段案例区：
  - `练习复盘`
  - `AWD 复盘`
- 两段都使用统一案例卡语言：标题、副线、状态、指标摘要、阶段摘要、展开明细。

练习复盘区：

- 按 `challenge_id` 聚合。
- 卡片副线固定为 `练习轨迹`。
- 状态标签只使用：
  - `已形成闭环`
  - `已完成`
  - `进行中`
  - `已记录`
- 指标固定为：
  - `有效提交`
  - `复盘材料`
  - `最近活动`
- 阶段固定为：
  - `接入`
  - `利用`
  - `提交`
  - `复盘`

AWD 复盘区：

- 按 `challenge_id + victim_team_name` 聚合。
- 卡片副线显示目标队伍名，而不是训练来源。
- 状态标签只使用：
  - `攻击命中`
  - `攻击未命中`
- 指标固定为：
  - `攻击次数`
  - `累计得分`
  - `最近活动`
- 阶段固定为：
  - `攻击尝试`
  - `命中结果`
  - `得分变化`

展开行为：

- 按钮文案固定为 `展开案例 / 收起案例`。
- 展开后在当前暗色 section 内显示事件列表，不切到新的对话框或二级页面。

---

## 4. 夜间模式交互规则

- loading、error、empty state 都留在当前暗色 section 内，不切出白底提示框。
- 表格的 header、body wrapper、empty block、scroll wrapper 必须保持统一背景。
- 筛选控件、对话框和文本输入默认继承 `teacher-surface` token，不再单独写浅色默认值。
- 页面允许局部 accent 差异，但不再为不同教师页维护多套不兼容的暗色表面。

---

## 代码落点

- 共享表面：
  - `code/frontend/src/assets/styles/teacher-surface.css`
- 目录页：
  - `code/frontend/src/components/teacher/class-management/ClassManagementPage.vue`
  - `code/frontend/src/components/teacher/student-management/StudentManagementPage.vue`
  - `code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue`
- AWD 复盘目录：
  - `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
  - `code/frontend/src/widgets/teacher-awd-review/TeacherAWDReviewIndexWorkspace.vue`
- 学生复盘归档页：
  - `code/frontend/src/views/teacher/TeacherStudentReviewArchive.vue`
  - `code/frontend/src/widgets/teacher-review-archive/TeacherReviewArchiveWorkspace.vue`
  - `code/frontend/src/components/teacher/review-archive/ReviewArchiveHero.vue`
  - `code/frontend/src/widgets/teacher-review-archive/TeacherReviewArchiveSummarySection.vue`
  - `code/frontend/src/components/teacher/review-archive/ReviewArchiveEvidencePanel.vue`
  - `code/frontend/src/components/teacher/review-archive/ReviewArchiveReflectionPanel.vue`
  - `code/frontend/src/components/teacher/review-archive/reviewArchiveCases.ts`
