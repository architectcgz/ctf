> 状态：Current
> 事实源：platform/teacher student management page models、student directory route views、前端架构 allowlist
> 替代：无

# Student Management Route Target Cleanup Plan

## 目标

- 把 `usePlatformStudentManagementPage.ts` 与 `useStudentManagementPage.ts` 从单次 `router.push()` 收口成纯 route target contract。
- 保持平台 / 教师学生目录的数据 owner 不变，同时再清掉 2 条 `featureRouterImportAllowlist`。

## 非目标

- 不处理 `StudentAnalysis`、`ClassManagement` 等其它页面自身的 route owner。
- 不改学生目录的数据拉取、筛选、分页和报告导出 owner。
- 不继续做学生目录 UI 拆壳。

## 输入依据

- `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/teacher-student-management/model/useStudentManagementPage.ts`
- `code/frontend/src/components/platform/student/StudentManageWorkspacePanel.vue`
- `code/frontend/src/features/teacher-student-management/ui/StudentManagementPage.vue`
- `code/frontend/src/views/platform/StudentManage.vue`
- `code/frontend/src/views/teacher/TeacherStudentManagement.vue`
- `code/frontend/src/views/platform/__tests__/StudentManage.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/components/navigation/AppRouteLink.vue`

## 当前结论

- 平台学生目录的 router 依赖只剩“进入学员分析”这一处。
- 教师学生目录的 router 依赖只剩“进入学员分析”和“回班级管理”两处；真正需要保留在 page model 里的本地 workflow 是 `reportDialogVisible`。
- 这两条都更适合收口成 route target contract，而不是继续保留 route-aware page owner。

## 设计边界

### `platformStudentManagementRoutes.ts` 本轮负责

- 生成平台学员分析页 route target

### `teacherStudentManagementRoutes.ts` 本轮负责

- 生成教师 / 管理员学生分析页 route target
- 生成教师 / 管理员班级管理页 route target

### `usePlatformStudentManagementPage()` / `useStudentManagementPage()` 本轮负责

- 学生目录数据加载、错误态、分页和本地 dialog owner
- 暴露 route target contract，不再直接导航

### `StudentManageWorkspacePanel.vue` / `StudentManagementPage.vue` 本轮负责

- 通过共享 `AppRouteLink` 消费 route target
- 保持筛选、分页和报告导出事件 owner 不变

## 任务切片

### Slice 1：page model 去 router 化

- 目标：
  - 新增平台 / 教师学生目录 route helper
  - 两个 page model 去掉 `vue-router`
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - allowlist 是否可净减少 2 条

### Slice 2：学生目录 UI 切到 `AppRouteLink`

- 目标：
  - 平台目录的“查看学员”改成共享 `AppRouteLink`
  - 教师目录的“班级管理”“学员分析”改成共享 `AppRouteLink`
  - “导出班级报告”继续保留 button + emit
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/StudentManage.test.ts src/views/teacher/__tests__/TeacherStudentManagement.test.ts`
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

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/StudentManage.test.ts src/views/teacher/__tests__/TeacherStudentManagement.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/student-management-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-student-management-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-student-management-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-student-management/model/platformStudentManagementRoutes.ts code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts code/frontend/src/features/teacher-student-management/model/teacherStudentManagementRoutes.ts code/frontend/src/features/teacher-student-management/model/useStudentManagementPage.ts code/frontend/src/components/platform/student/StudentManageWorkspacePanel.vue code/frontend/src/features/teacher-student-management/ui/StudentManagementPage.vue code/frontend/src/views/platform/StudentManage.vue code/frontend/src/views/teacher/TeacherStudentManagement.vue code/frontend/src/views/platform/__tests__/StudentManage.test.ts code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 学员分析页、自身班级目录页等其它 route owner 不在这轮范围内。
