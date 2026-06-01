# platform class directory state split 计划

## Objective

- 在 `platform/class-management` 内新增 `usePlatformClassDirectory`，承接平台端班级目录的数据 owner。
- 让 `usePlatformClassManagementPage` 只保留后台页面壳的行映射、指标和路由编排。

## Non-goals

- 不改 teacher class management。
- 不把 teacher / platform 班级目录合成新的 shared page owner。
- 不调整 `ClassManageWorkspacePanel.vue` 的 UI 和交互结构。

## Source Inputs

- `code/frontend/src/features/platform/class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/pages/platform/__tests__/ClassManage.test.ts`
- `code/frontend/src/features/platform/student-management/model/usePlatformStudentDirectory.ts`
- `code/frontend/src/features/teacher/class-management/model/useTeacherClassDirectory.ts`

## Plan Review Result

- 平台端班级管理最合理的第一刀也是 `page + directory` 拆分。
- `usePlatformClassDirectory` 负责班级目录请求、分页和错误状态。
- `usePlatformClassManagementPage` 保留后台页面需要的 summary 指标、表格行映射和班级详情 route。

## Task Slices

### Slice 1: 新建 usePlatformClassDirectory

- 目标：收口平台端班级目录加载、分页和错误状态。
- 风险：
  - 如果把 rows 或 route 一起搬走，会重新模糊 page owner。

### Slice 2: usePlatformClassManagementPage 改为消费 directory owner

- 目标：保留平台端表格行派生、指标和 route 页面语义。
- 风险：
  - 如果 page 继续直接依赖 `getClasses`，就没有真正收口目录 owner。

### Slice 3: 更新源码级和行为测试

- 目标：给新目录 owner 补直测，并更新平台端源码断言。
- 风险：
  - 不补数组响应兼容或分页测试，后面还会回流进 page owner。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision platform-class-directory-state-split`
- `npm run test:run -- src/features/platform/class-management/model/usePlatformClassDirectory.test.ts src/pages/platform/__tests__/ClassManage.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `npm run typecheck`
- `git diff --check`
- `bash scripts/check-reuse-first.sh`

## Review Focus

- `usePlatformClassDirectory` 是否只承接平台端班级目录的数据 owner。
- `usePlatformClassManagementPage` 是否只剩后台行映射、指标和 route 编排。
- `ClassManageRoutePage.vue` 是否继续只负责 route page 组合。

## Rollback / Recovery

- 如果 `usePlatformClassDirectory` 的返回接口不顺手，可以调整字段组织，但班级目录的数据加载 owner 仍必须留在新 composable。
