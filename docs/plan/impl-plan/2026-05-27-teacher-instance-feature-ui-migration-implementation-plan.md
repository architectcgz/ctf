> 状态：Current
> 事实源：`TeacherInstanceManagementPage.vue` 当前 owner、教师实例 route view 与 `teacher-instances` feature public API 边界
> 替代：无

# Teacher Instance Feature UI Migration Implementation Plan

## 目标

- 把 `TeacherInstanceManagementPage.vue` 从 `components/teacher/instance-management/` 迁到 `features/teacher-instances/ui/`。
- 让 `views/teacher/InstanceManagement.vue` 直接通过 `features/teacher-instances` public API 组合 page-sized UI 与 `useInstanceManagementPage()`。
- 收掉 `TeacherInstanceManagementPage.vue` 对应的 `legacyComponentPageAllowlist` 例外，并同步更新教师端 raw-source 测试路径。

## 非目标

- 本轮不改 `useInstanceManagementPage()` / `useInstances()` 的加载、筛选、防抖、销毁、分页或 dashboard 跳转 owner。
- 本轮不迁 `TeacherInstanceHeroPanel.vue` 与 `TeacherInstanceDirectorySection.vue` 的目录位置。
- 本轮不改教师实例页的用户可见文案、数据字段、筛选交互和分页行为。

## 输入依据

- `docs/architecture/frontend/06-components.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue`
- `code/frontend/src/views/teacher/InstanceManagement.vue`
- `code/frontend/src/features/teacher-instances/index.ts`
- `code/frontend/src/features/teacher-instances/model/useInstanceManagementPage.ts`
- `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `TeacherInstanceManagementPage.vue` 现在只承担 page shell 与事件桥接，真正的业务 owner 已经在 `useInstanceManagementPage.ts` / `useInstances.ts`。
- 这页不再持有 `useRoute/useRouter` 或 API owner，所以这轮比 `UserGovernancePage.vue` 更简单，不需要新增 route model。
- 当前残留结构债是它仍以 `components/*Page.vue` 的形式存在，并继续占用 `legacyComponentPageAllowlist`。

## 设计边界

### route view 继续负责

- 组合 `useInstanceManagementPage()` 与 `TeacherInstanceManagementPage`
- 不直接持有初始化、筛选、防抖、销毁确认、分页或 dashboard 跳转 owner

### `features/teacher-instances/model` 继续负责

- 初始化、班级/关键字/学号筛选、防抖加载、分页切换
- 销毁确认后的实例移除
- dashboard 跳转 owner

### `features/teacher-instances/ui` 本轮负责

- 教师实例页 page-sized shell
- 消费上层派生状态与事件 handler
- 不直接持有 API、router 或确认弹窗 owner

### `components/teacher/instance-management/*` 继续保留

- `TeacherInstanceHeroPanel.vue`
- `TeacherInstanceDirectorySection.vue`

## 任务切片

### Slice 1：迁移教师实例 page shell 到 feature ui

- 目标：
  - 新增 `features/teacher-instances/ui/TeacherInstanceManagementPage.vue`
  - `features/teacher-instances/index.ts` 导出 `ui`
  - `views/teacher/InstanceManagement.vue` 改从 feature public API 引用
- 预期改动：
  - `code/frontend/src/features/teacher-instances/index.ts`
  - `code/frontend/src/features/teacher-instances/ui/*`
  - `code/frontend/src/views/teacher/InstanceManagement.vue`
- 验证：
  - `npm run test:run -- src/views/teacher/__tests__/InstanceManagement.test.ts`
- Review focus：
  - route view 是否继续只保留组合壳
  - page shell 是否仍未吸回业务 owner

### Slice 2：更新 guardrail、测试与 backlog

- 目标：
  - 更新 raw-source 测试路径与自动注册类型路径
  - 清理 `legacyComponentPageAllowlist` 中对应的 page 例外
  - 记录 backlog 进展
- 预期改动：
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/components.d.ts`
  - `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherRootShellCleanup.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts`
  - `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
  - `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/InstanceManagement.test.ts src/views/teacher/__tests__/teacherRootShellCleanup.test.ts src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts src/views/teacher/__tests__/teacherSurface.test.ts src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Review focus：
  - allowlist 是否真实下降
  - 教师端 surface / shell 测试是否已经切到新 owner

## 结构收口检查

- `TeacherInstanceManagementPage.vue` 不再作为 `components/*Page.vue` 存在。
- `views/teacher/InstanceManagement.vue` 只保留 feature public API 组合壳。
- `features/teacher-instances` public API 明确导出 `model + ui`。
- touched surface 上至少移除一条 `legacyComponentPageAllowlist`。

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/InstanceManagement.test.ts src/views/teacher/__tests__/teacherRootShellCleanup.test.ts src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts src/views/teacher/__tests__/teacherSurface.test.ts src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/teacher-instance-feature-ui-migration.md docs/plan/impl-plan/2026-05-27-teacher-instance-feature-ui-migration-implementation-plan.md docs/reviews/frontend/2026-05-27-teacher-instance-feature-ui-migration-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/features/teacher-instances code/frontend/src/views/teacher/InstanceManagement.vue code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/components.d.ts code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts code/frontend/src/views/teacher/__tests__/teacherRootShellCleanup.test.ts code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts code/frontend/src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts code/frontend/src/views/__tests__/workspaceShellStyles.test.ts code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review 关注点

- `features/teacher-instances/ui` 是否成为教师实例页 page shell 的唯一 owner。
- `useInstanceManagementPage()` 是否仍然保持初始化、筛选、防抖、销毁与 dashboard 跳转 owner，不被迁移反向打散。
- 教师端 raw-source 测试与 allowlist 是否同步反映新边界。

## 回退 / 恢复说明

- 如迁移后出现问题，可把 `TeacherInstanceManagementPage.vue` 移回 `components/teacher/instance-management/` 并恢复 route view import。
- 如教师端 source 测试出现误报，可先回退测试路径切换，再单独重做 page shell 迁移。

## 残余风险

- `TeacherInstanceHeroPanel.vue` 与 `TeacherInstanceDirectorySection.vue` 仍保留在 `components/teacher/instance-management/`，本轮只处理 page shell 落位。
- 其它 teacher legacy page 仍在 backlog 中，本轮不一并迁移。
