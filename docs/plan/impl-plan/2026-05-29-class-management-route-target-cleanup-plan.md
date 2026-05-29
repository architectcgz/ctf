> 状态：Current
> 事实源：platform/teacher class management page models、class directory route views、前端架构 allowlist
> 替代：无

# Class Management Route Target Cleanup Plan

## 目标

- 把 `usePlatformClassManagementPage.ts` 与 `useClassManagementPage.ts` 从单次 `router.push()` 收口成纯 route target contract。
- 保持平台 / 教师班级目录的数据 owner 不变，同时再清掉 2 条 `featureRouterImportAllowlist`。

## 非目标

- 不处理 `ClassStudents`、`TeacherDashboard` 等其它页面自身的 route owner。
- 不改班级目录的数据拉取、筛选、分页和报告导出 owner。
- 不继续做班级目录 UI 拆壳。

## 输入依据

- `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/teacher-class-management/model/useClassManagementPage.ts`
- `code/frontend/src/views/platform/ClassManage.vue`
- `code/frontend/src/components/platform/class/ClassManageWorkspacePanel.vue`
- `code/frontend/src/components/navigation/AppRouteLink.vue`
- `code/frontend/src/components/teacher/class-management/TeacherClassManagementHeaderActions.vue`
- `code/frontend/src/components/teacher/class-management/TeacherClassManagementRowLink.vue`
- `code/frontend/src/features/teacher-class-management/ui/ClassManagementPage.vue`
- `code/frontend/src/views/teacher/ClassManagement.vue`
- `code/frontend/src/views/platform/__tests__/ClassManage.test.ts`
- `code/frontend/src/views/teacher/__tests__/ClassManagement.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- 平台班级目录的 router 依赖只剩“进入班级详情”这一处。
- 教师班级目录的 router 依赖只剩“进入班级详情”和“回教学概览”两处；真正需要保留在 page model 里的本地 workflow 是 `reportDialogVisible`。
- 这两条都更适合收口成 route target contract，而不是继续保留 route-aware page owner。

## 设计边界

### `platformClassManagementRoutes.ts` 本轮负责

- 生成平台班级详情页 route target

### `teacherClassManagementRoutes.ts` 本轮负责

- 生成教师班级详情页 route target
- 生成教师教学概览 route target

### `usePlatformClassManagementPage()` / `useClassManagementPage()` 本轮负责

- 班级目录数据加载、错误态、分页和本地 dialog owner
- 暴露 route target contract，不再直接导航

### `AppRouteLink.vue`、`ClassManageWorkspacePanel.vue` / `ClassManagementPage.vue` 及教师目录动作子组件本轮负责

- 在共享 route link bridge 上消费 route target，避免每个本地组件都各自直接碰 `vue-router`
- 保持刷新、分页和报告导出事件 owner 不变

## 任务切片

### Slice 1：page model 去 router 化

- 目标：
  - 新增平台 / 教师班级目录 route helper
  - 两个 page model 去掉 `vue-router`
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - allowlist 是否可净减少 2 条

### Slice 2：班级目录 UI 切到共享 route link bridge

- 目标：
  - 平台目录的“查看班级”改成共享 route link bridge
  - 教师目录的“教学概览”“进入班级”改成共享 route link bridge
  - “导出班级报告”继续保留 button + emit
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ClassManage.test.ts src/views/teacher/__tests__/ClassManagement.test.ts`
- Review focus：
  - route target contract 是否清楚，没有把非路由 workflow 一起迁走

### Slice 3：allowlist / review / backlog 收尾

- 目标：
  - 更新 allowlist、review 和 backlog
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - `featureRouterImportAllowlist` 是否净减少 2 条

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ClassManage.test.ts src/views/teacher/__tests__/ClassManagement.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/class-management-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-class-management-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-class-management-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/components/navigation/AppRouteLink.vue code/frontend/src/features/platform-class-management/model/platformClassManagementRoutes.ts code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts code/frontend/src/components/platform/class/ClassManageWorkspacePanel.vue code/frontend/src/components/teacher/class-management/TeacherClassManagementHeaderActions.vue code/frontend/src/components/teacher/class-management/TeacherClassManagementRowLink.vue code/frontend/src/features/teacher-class-management/model/teacherClassManagementRoutes.ts code/frontend/src/features/teacher-class-management/model/useClassManagementPage.ts code/frontend/src/features/teacher-class-management/ui/ClassManagementPage.vue code/frontend/src/views/platform/ClassManage.vue code/frontend/src/views/teacher/ClassManagement.vue code/frontend/src/views/platform/__tests__/ClassManage.test.ts code/frontend/src/views/teacher/__tests__/ClassManagement.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 教师教学概览、自身班级工作区等其它 route owner 不在这轮范围内。
