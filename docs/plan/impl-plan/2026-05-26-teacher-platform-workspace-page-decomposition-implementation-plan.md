> 状态：Current
> 事实源：`UserGovernancePage.vue`、`TeacherDashboardPage.vue`、`ClassStudentsPage.vue` 当前 owner，相关 route view / feature composable / 测试护栏，现有 workspace page 拆分模式
> 替代：无

# Teacher / Platform Workspace Page Decomposition Implementation Plan

## 目标

- 拆分 `UserGovernancePage.vue`、`TeacherDashboardPage.vue`、`ClassStudentsPage.vue` 这三个过宽页面，把稳定展示区块从 page 壳里拆出去。
- 保持 route view、feature composable、查询同步、导航动作和主要业务事件的 owner 不变，只收口页面组件内部混杂的展示职责。
- 同步更新源码护栏、架构 allowlist 和页面测试，让后续新增需求不再继续堆回这三个大文件。

## 非目标

- 本轮不改 `usePlatformUserManagePage.ts`、`useDashboardPage.ts`、`useClassStudentsPage.ts` 的 route / async owner 边界。
- 本轮不改变用户导入、班级训练时间段、教师概览 tab、用户详情弹窗这些既有交互行为。
- 本轮不把这三个页面整体迁移成新的 widgets / entities 分层，也不重做现有视觉样式系统。

## 输入依据

- `code/frontend/src/components/platform/user/UserGovernancePage.vue`
- `code/frontend/src/views/platform/UserManage.vue`
- `code/frontend/src/features/platform-user-management/model/usePlatformUserManagePage.ts`
- `code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue`
- `code/frontend/src/views/teacher/TeacherDashboard.vue`
- `code/frontend/src/features/teacher-dashboard/model/useDashboardPage.ts`
- `code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue`
- `code/frontend/src/views/teacher/TeacherClassStudents.vue`
- `code/frontend/src/views/platform/PlatformClassStudents.vue`
- `code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/platform/__tests__/UserManage.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts`

## 当前结论

- `TeacherDashboardPage.vue` 现在把 tab schema、tab 切换、overview 头部、portrait/trend/review/intervention 四个展示面板和大段样式堆在一个文件里，已经是典型的单 page 多展示区块混写。
- `ClassStudentsPage.vue` 同时承载页面壳、时间段过滤、概览头部、学生目录和四个 workspace subpanel 挂载，展示 owner 已经明显超过“一个页面壳”应有的密度。
- `UserGovernancePage.vue` 既承载工作台头部、目录区、详情弹窗，又承载独立导入面板和路由面板切换，导致展示结构和路由面板桥接耦合在一起。
- 这三处的 route view 与 feature composable 已经在仓库里承担 page owner，最小正确拆分不是再移动 route/api owner，而是把页面组件内部的稳定展示块切成清晰的子组件，并让父 page 继续装配它们。

## 任务切片

### Slice 1：拆分教师概览页展示面板

- 目标：
  - 从 `TeacherDashboardPage.vue` 抽出 overview / portrait / trend / review / intervention 中至少一组稳定展示子组件。
  - 保留 `TeacherDashboardPage.vue` 作为 tab owner、page shell 和 `useDashboardMetrics()` 消费边界。
- 预期改动：
  - `code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue`
  - `code/frontend/src/components/teacher/dashboard/*.vue`
  - `code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts`
  - 相关样式 / allowlist / 原始源码护栏
- 验证：
  - `npm run test:run -- src/views/teacher/__tests__/TeacherDashboard.test.ts`
- Review focus：
  - tab owner 是否仍留在 page
  - `useDashboardMetrics()` 输出 contract 是否没有被 child 组件反向接管
  - 样式 token 与工作台壳是否保持原语义

### Slice 2：拆分班级学生工作台展示区块

- 目标：
  - 从 `ClassStudentsPage.vue` 抽出时间段控制区、概览头部、学生目录、动态 panel 挂载中边界最清楚的块。
  - 保持 `ClassStudentsPage.vue` 继续拥有 tab 切换、事件透传和子面板装配。
- 预期改动：
  - `code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue`
  - `code/frontend/src/components/teacher/class-management/*.vue`
  - `code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts`
  - `code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts`
  - 相关源码护栏 / allowlist
- 验证：
  - `npm run test:run -- src/views/teacher/__tests__/TeacherClassStudents.test.ts src/views/platform/__tests__/PlatformClassStudents.test.ts`
- Review focus：
  - 班级时间段和学生列表过滤事件是否仍由 page 壳统一透传
  - `ClassTrendPanel` / `ClassReviewPanel` / `ClassInsightsPanel` / `InterventionPanel` 的动态挂载 contract 是否不变
  - 平台端与教师端共用页面的行为是否一致

### Slice 3：拆分用户治理工作台与导入面板

- 目标：
  - 从 `UserGovernancePage.vue` 抽出 overview 目录工作台、用户详情弹窗、导入面板中稳定且可复用的展示块。
  - 保留 `UserGovernancePage.vue` 继续 owning route `panel` 切换、详情选中态和顶层事件桥接。
- 预期改动：
  - `code/frontend/src/components/platform/user/UserGovernancePage.vue`
  - `code/frontend/src/components/platform/user/*.vue`
  - `code/frontend/src/views/platform/__tests__/UserManage.test.ts`
  - 相关样式 / allowlist / 原始源码护栏
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/UserManage.test.ts`
- Review focus：
  - `switchPanel()` 仍留在 page owner，child 组件不直接依赖 `vue-router`
  - 详情弹窗的编辑 / 删除事件链条保持不变
  - 导入面板仍是独立面板，不与 overview 目录重新耦合

## 集成检查

- 三个 route view 仍只做 page 组合，不回流 `useRoute` / `useRouter` / API 直接依赖。
- 新子组件不直接接管 feature composable、router 或非 contract API owner。
- `architectureAllowlist`、raw source 测试和页面行为测试能共同证明拆分后边界更清楚，而不是只把长模板平移出去。

## 回退 / 恢复说明

- 每个页面拆分都应保持可独立回退：子组件文件、父 page 接线和对应测试可以按页面粒度还原。
- 本轮不涉及数据迁移、接口变更和配置变更，因此回退主要是页面组件结构回退。

## 残余风险

- 现有测试大量依赖 `?raw` 源码断言；拆分后需要同步收口这些护栏，否则容易出现“结构改善但护栏误报”的假失败。
- 如果拆分时把样式变量分散到多个子组件，可能引入 token owner 漂移；需要优先让 page 壳继续持有共享 surface 变量。
- `UserGovernancePage.vue` 与 `ClassStudentsPage.vue` 都有跨平台 / 跨角色复用面，拆分时要避免在子组件里硬编码 teacher 或 platform 特有分支。
