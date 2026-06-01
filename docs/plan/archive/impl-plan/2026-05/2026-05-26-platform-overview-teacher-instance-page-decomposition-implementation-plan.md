> 状态：Current
> 事实源：`PlatformOverviewPage.vue`、`TeacherInstanceManagementPage.vue` 当前 owner，相关 route view / feature composable / 测试护栏，现有 workspace page 拆分模式
> 替代：无

# Platform Overview / Teacher Instance Page Decomposition Implementation Plan

## 目标

- 拆分 `PlatformOverviewPage.vue` 和 `TeacherInstanceManagementPage.vue` 这两个仍然过宽的前端 page 组件，把稳定展示块从父页壳中抽出。
- 保持 `PlatformOverview.vue`、`InstanceManagement.vue`、`usePlatformOverviewPage.ts`、`useInstanceManagementPage.ts` 继续 owning route、导航动作、加载与主业务事件。
- 同步更新 raw-source 护栏和页面行为测试，让这两页后续不再继续向单文件堆展示结构。

## 非目标

- 本轮不改 `usePlatformOverviewPage.ts`、`usePlatformOverviewWorkspace.ts`、`useInstanceManagementPage.ts`、`useInstances.ts` 的 owner 边界。
- 本轮不调整平台概览的数据结构、教师实例查询 contract、分页 contract、销毁实例交互语义。
- 本轮不做 `platform-users` feature 进一步拆分，不处理 admin / teacher 结构耦合等更高层结构债。

## 输入依据

- `code/frontend/src/components/platform/dashboard/PlatformOverviewPage.vue`
- `code/frontend/src/views/platform/PlatformOverview.vue`
- `code/frontend/src/features/platform-overview/model/usePlatformOverviewPage.ts`
- `code/frontend/src/features/platform-overview/model/usePlatformOverviewWorkspace.ts`
- `code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts`
- `code/frontend/src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue`
- `code/frontend/src/views/teacher/InstanceManagement.vue`
- `code/frontend/src/features/teacher-instances/model/useInstanceManagementPage.ts`
- `code/frontend/src/features/teacher-instances/model/useInstances.ts`
- `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/components/platform/instance/InstanceManageHeroPanel.vue`
- `code/frontend/src/components/platform/instance/InstanceManageWorkspacePanel.vue`
- `docs/reviews/frontend/README.md`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`

## 当前结论

- `PlatformOverviewPage.vue` 目前同时承载 overview hero、summary skeleton / metric strip、error alert、alerts section、hotspots section 和整段 page-level scoped style，已经超过一个展示壳应有的密度。
- `TeacherInstanceManagementPage.vue` 目前同时承载 page shell、header、summary strip、directory filters、table、pagination、error block 和整段 style owner，且目录区块边界很稳定，适合抽成独立 section。
- 这两页已经有清晰的 route view 与 feature composable owner，最小正确拆分不是继续迁 route / async owner，而是把稳定展示块切成子组件，同时让父页继续装配它们。

## 任务切片

### Slice 1：拆分平台概览页展示区块

- 目标：
  - 从 `PlatformOverviewPage.vue` 抽出 hero / summary 区和 alerts / hotspots 目录区。
  - 保留 `PlatformOverviewPage.vue` 继续作为 page shell 和 `usePlatformOverviewWorkspace()` 的消费边界。
- 预期改动：
  - `code/frontend/src/components/platform/dashboard/PlatformOverviewPage.vue`
  - `code/frontend/src/components/platform/dashboard/PlatformOverviewHeroPanel.vue`
  - `code/frontend/src/components/platform/dashboard/PlatformOverviewAlertsSection.vue`
  - `code/frontend/src/components/platform/dashboard/PlatformOverviewHotspotsSection.vue`
  - `code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts`
  - 以及引用 `PlatformOverviewPage.vue?raw` 的共享护栏
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/PlatformOverview.test.ts src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts`
- Review focus：
  - `openAuditLog` / `openCheatDetection` / `retry` 是否仍由父层 page owner 装配
  - `usePlatformOverviewWorkspace()` 的派生值是否没有被 child 组件反向接管
  - hero / alerts / hotspots 三块拆分后，surface token owner 是否仍清晰

### Slice 2：拆分教师实例页展示区块

- 目标：
  - 从 `TeacherInstanceManagementPage.vue` 抽出 hero / summary 区和实例目录区。
  - 保留 `TeacherInstanceManagementPage.vue` 继续装配 `useInstanceManagementPage()` 返回的筛选、分页、销毁和导航动作。
- 预期改动：
  - `code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue`
  - `code/frontend/src/components/teacher/instance-management/TeacherInstanceHeroPanel.vue`
  - `code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue`
  - `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
  - 以及引用 `TeacherInstanceManagementPage.vue?raw` 的教师端 / shared 护栏
- 验证：
  - `npm run test:run -- src/views/teacher/__tests__/InstanceManagement.test.ts`
- Review focus：
  - `destroy`、`changePage`、`updateKeyword`、`updateStudentNo`、`updateClassName` 是否仍由父页透传
  - `openDashboard` 与 `retry` 是否仍留在 page owner
  - 目录区块拆分后是否保持现有 teacher surface token 与 directory shell 语义

## 集成检查

- 两个 route view 仍只负责组合，不回流 API、router 细节或 confirm owner。
- 新子组件不直接接管 feature composable、router、store 或 page-level async owner。
- raw-source 护栏统一改成“父页 + 子组件组合源码”策略，避免为了过测试重新把模板塞回父页。

## 回退 / 恢复说明

- 每个页面拆分都应保持可独立回退：子组件文件、父页接线和对应测试可以按页面粒度回退。
- 本轮不涉及 API、路由命名、数据迁移或状态机改动，因此回退主要是前端组件结构回退。

## 残余风险

- `PlatformOverviewPage.vue` 与 `TeacherInstanceManagementPage.vue` 都有较多 raw-source 护栏；拆分后如果漏掉组合源测试，容易出现“实现没坏但护栏误报”的假失败。
- 教师实例页当前显示文案里还保留 `@username` 形式；这轮不改变行为，但它和 `frontend-engineer` 的更长期规则存在张力，后续如果继续碰这一块要单独判断是产品文案还是技术债。
- 平台概览页和教师实例页拆分完成后，`TD-1` 仍未结束，后续更大的组件壳还在 `layout`、`contest`、`AWD` 线上。
