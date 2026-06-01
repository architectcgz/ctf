> 状态：Current
> 事实源：platform/teacher instance management page models、instance directory route views、前端架构 allowlist
> 替代：无

# Instance Management Route Target Cleanup Plan

## 目标

- 把 `usePlatformInstanceManagementPage.ts` 与 `useInstanceManagementPage.ts` 从单次 `router.push()` 收口成纯 route target contract。
- 保持平台 / 教师实例目录的数据、销毁、筛选与分页 owner 不变，同时再清掉 2 条 `featureRouterImportAllowlist`。

## 非目标

- 不处理 `useInstances.ts` 的数据加载与 stale-response owner。
- 不改实例目录的销毁确认、刷新、筛选、分页和错误态流程。
- 不继续做实例目录 UI 拆壳。

## 输入依据

- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/teacher-instances/model/useInstanceManagementPage.ts`
- `code/frontend/src/components/platform/instance/InstanceManageHeroPanel.vue`
- `code/frontend/src/components/platform/instance/InstanceManageWorkspacePanel.vue`
- `code/frontend/src/features/teacher-instances/ui/TeacherInstanceManagementPage.vue`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceHeroPanel.vue`
- `code/frontend/src/views/platform/InstanceManage.vue`
- `code/frontend/src/views/teacher/InstanceManagement.vue`
- `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/components/navigation/AppRouteLink.vue`

## 当前结论

- 平台实例目录的 router 依赖只剩“返回概览”和“进入学员分析”两处。
- 教师实例目录的 router 依赖只剩“返回教学概览”一处。
- 两边都更适合收口成 route target contract，而不是继续让 page model 持有 router。

## 设计边界

### `platformInstanceManagementRoutes.ts` 本轮负责

- 生成平台概览页 route target
- 生成平台学员分析页 route target

### `teacherInstanceManagementRoutes.ts` 本轮负责

- 生成教师 / 管理员实例目录返回概览页 route target

### `usePlatformInstanceManagementPage()` / `useInstanceManagementPage()` 本轮负责

- 实例目录数据加载、筛选、分页、销毁与本地 workflow owner
- 暴露 route target contract，不再直接导航

### `InstanceManageHeroPanel.vue` / `InstanceManageWorkspacePanel.vue` / `TeacherInstanceHeroPanel.vue` 本轮负责

- 通过共享 `AppRouteLink` 消费 route target
- 保持刷新、销毁与筛选交互 owner 不变

## 任务切片

### Slice 1：page model 去 router 化

- 目标：
  - 新增平台 / 教师实例目录 route helper
  - 两个 page model 去掉 `vue-router`
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - allowlist 是否可净减少 2 条

### Slice 2：实例目录 UI 切到 `AppRouteLink`

- 目标：
  - 平台实例页的“返回概览”“所属用户”改成共享 `AppRouteLink`
  - 教师实例页的“返回教学概览”改成共享 `AppRouteLink`
  - 实例销毁、筛选、分页与刷新保持原 owner
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/InstanceManage.test.ts src/views/teacher/__tests__/InstanceManagement.test.ts`
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

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/InstanceManage.test.ts src/views/teacher/__tests__/InstanceManagement.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/instance-management-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-instance-management-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-instance-management-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-instance-management/model/platformInstanceManagementRoutes.ts code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts code/frontend/src/features/teacher-instances/model/teacherInstanceManagementRoutes.ts code/frontend/src/features/teacher-instances/model/useInstanceManagementPage.ts code/frontend/src/components/platform/instance/InstanceManageHeroPanel.vue code/frontend/src/components/platform/instance/InstanceManageWorkspacePanel.vue code/frontend/src/features/teacher-instances/ui/TeacherInstanceManagementPage.vue code/frontend/src/components/teacher/instance-management/TeacherInstanceHeroPanel.vue code/frontend/src/views/platform/InstanceManage.vue code/frontend/src/views/teacher/InstanceManagement.vue code/frontend/src/views/platform/__tests__/InstanceManage.test.ts code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 实例目录更深层的数据查询与销毁 workflow owner 不在这轮范围内。
