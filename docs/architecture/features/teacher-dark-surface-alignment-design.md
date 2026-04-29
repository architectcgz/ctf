# 教师端管理页夜间模式表面对齐设计

## 目标

对齐以下教师端页面在夜间模式下的表面体系，使其和现有 `TeacherDashboard` / `notifications` 一致，解决亮底、漏白、层级失衡和 Element Plus 默认浅色外露问题：

- `/academy/classes`
- `/academy/students`
- `/academy/instances`
- `/academy/awd-reviews`
- `/academy/classes/:className/students/:studentId/review-archive`

本轮允许调整控件层级和视觉细节，但不改页面结构、路由、接口和交互流程。

## 非目标

- 不重做教师端信息架构
- 不修改业务流程、筛选逻辑和导出流程
- 不引入新的全站主题系统
- 不对未点名页面做大范围样式重刷

## 当前问题

这 4 个页面已经部分使用暗色 token，但仍有三类不一致：

1. 页面之间的语义变量和 accent 写法分散，导致同属教师端却不像同一套系统
2. `ElTable`、输入框、下拉、按钮、空态、错误态和弹窗的内部包裹层未完全收敛到暗色表面
3. 旧的独立报告页已经迁移为上下文导出对话框和复盘归档页，教师端暗色表面应继续由这些承接页面保持一致

## 方案选择

### 方案 A：最小修补

只修明显漏白节点和 Element Plus 默认浅色层。

优点：

- 改动小
- 风险低

缺点：

- 页面之间仍然不够统一
- 后续再修会继续重复补洞

### 方案 B：教师端共享表面对齐加页面局部补漏

收敛教师端管理页的语义 surface token，并在 4 个页面内补齐 Element Plus 和状态容器的暗色层。

优点：

- 一致性最好
- 影响范围仍然可控
- 能顺手解决按钮、表格、筛选区、空态和错误态的细节问题

缺点：

- 改动比最小修补稍大

### 方案 C：全局 Element Plus 暗色覆盖

从全局共享样式直接覆盖所有常见 Element Plus 组件。

优点：

- 可一次影响更多页面

缺点：

- 外溢风险高
- 容易影响当前未验证页面

## 决策

采用方案 B。

实现策略：

1. 以教师端现有低对比暗色表面为基准，统一 `journal-ink`、`journal-muted`、`journal-border`、`journal-surface`、`journal-surface-subtle`
2. 抽取一层仅面向教师端管理页的共享 surface 样式
3. 各页面只保留少量 page-local accent 和布局差异
4. 对 Element Plus 泄漏层采用稳定 class + `:deep(...)` 精准覆盖，不做粗暴全局覆盖

## 设计细节

### 共享表面层

新增或抽取一份教师端 surface 样式，提供以下能力：

- 教师端页面外层 shell、hero、brief、metric、section 的统一底色、边框和阴影
- 教师端按钮的 ghost / primary / danger 层级
- 教师端表格常用变量与 wrapper 背景覆盖
- 教师端筛选控件、空态、错误态的表面 token

这层样式只服务教师端管理页，避免影响管理员页、学生页和其它业务视图。

### `/academy/classes`

保留当前 hero 和列表结构，重点修正：

- `ElTable` 的 `el-table`、`__inner-wrapper`、`__header-wrapper`、`__body-wrapper`、空态容器背景统一为 `journal-surface`
- hover、边框和表头文字对比降到教师端现有标准
- 操作按钮与空态、错误态卡片和其余教师页一致

### `/academy/students`

保留当前筛选布局和学生表格，重点修正：

- `select`、搜索输入框、学号输入框统一为同层暗面控件
- 表格 wrapper、空态、错误态、loading skeleton 与列表区表面统一
- 避免筛选区和表格区像两套不同的暗色皮肤

### `/academy/instances`

该页已接近目标，重点收敛：

- 筛选表单容器与输入控件层级
- 状态 chip 的边框和底色对比
- 表格内部 wrapper、危险按钮、loading skeleton
- 与 `TeacherDashboard` 的 hero / section / metric 语义保持完全同源

### `/academy/awd-reviews` 与学生复盘归档

不改导出流程，只处理表面系统：

- 保持 AWD 复盘目录、详情页和学生复盘归档页的 `teacher-surface` 暗色体系一致
- 班级报告导出不再依赖独立 `/academy/reports` 页面，而是在班级/学生上下文中通过 `TeacherClassReportExportDialog` 承接
- 统一导出对话框、任务状态、空态和按钮层级
- 补齐 `ElDialog`、`input`、可能的 `textarea` 内层亮底问题

## 影响范围

预计涉及：

- `code/frontend/src/components/teacher/class-management/ClassManagementPage.vue`
- `code/frontend/src/components/teacher/student-management/StudentManagementPage.vue`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue`
- `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
- `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`
- `code/frontend/src/views/teacher/TeacherStudentReviewArchive.vue`
- `code/frontend/src/components/teacher/reports/TeacherClassReportExportDialog.vue`
- 一个教师端共享 surface 样式文件或已有教师端样式文件
- 相关前端测试文件

## 测试与验证

需要完成以下验证：

1. 运行教师端相关 Vitest：
   - `ClassManagement`
   - `TeacherStudentManagement`
   - `InstanceManagement`
   - `TeacherAWDReviewIndex`
   - `TeacherStudentReviewArchive`
2. 运行 `npm run build`，确认 SFC 与样式改动无编译问题
3. 如本地可访问页面，做人工核对，重点检查：
   - `ElTable` header/body/wrapper/empty block
   - `input` / `select` / `textarea` / `dialog` / `button`
   - 空态、错误态、loading skeleton、卡片内二级面板

## 风险与控制

- 若共享样式抽取过宽，可能影响其它教师页面；因此共享层只抽共通 token 和已被这 4 页重复使用的控件表面
- 若仅改外层不改 Element Plus 内层，亮底仍会漏出；因此必须覆盖到稳定的内部 wrapper
- 当前仓库存在其它未提交前端改动；本轮只叠加教师端夜间模式修正，不回退或改写无关文件
