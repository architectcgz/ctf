# teacher class directory state split 计划

## Objective

- 在 `teacher/class-management` 内新增 `useTeacherClassDirectory`，承接教师端班级目录的数据 owner。
- 让 `useClassManagementPage` 只保留教师端页面壳的 route / dialog 编排。

## Non-goals

- 不改 platform class management。
- 不把 teacher / platform 班级目录合成新的 shared page owner。
- 不调整 `ClassManagementPage.vue` 的 UI 和交互结构。

## Source Inputs

- `code/frontend/src/features/teacher/class-management/model/useClassManagementPage.ts`
- `code/frontend/src/pages/teacher/__tests__/ClassManagement.test.ts`
- `code/frontend/src/features/teacher/student-management/model/useTeacherStudentDirectory.ts`
- `code/frontend/src/features/platform/class-management/model/usePlatformClassManagementPage.ts`

## Plan Review Result

- 教师端班级管理最合理的第一刀是 `page + directory` 拆分。
- `useTeacherClassDirectory` 负责班级目录请求、分页和错误状态。
- `useClassManagementPage` 保留概览 route、班级详情 route 和报告弹窗 state。

## Task Slices

### Slice 1: 新建 useTeacherClassDirectory

- 目标：收口教师端班级目录加载、分页和错误状态。
- 风险：
  - 如果把 route 或 dialog 一起搬走，会重新模糊 page owner。

### Slice 2: useClassManagementPage 改为消费 directory owner

- 目标：保留教师端 route / dialog 页面语义。
- 风险：
  - 如果 page 继续直接依赖 `getClasses`，就没有真正收口目录 owner。

### Slice 3: 更新源码级和行为测试

- 目标：给新目录 owner 补直测，并更新教师端源码断言。
- 风险：
  - 不补分页和失败态测试，后面还会回流进 page owner。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision teacher-class-directory-state-split`
- `npm run test:run -- src/features/teacher/class-management/model/useTeacherClassDirectory.test.ts src/pages/teacher/__tests__/ClassManagement.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `npm run typecheck`
- `git diff --check`
- `bash scripts/check-reuse-first.sh`

## Review Focus

- `useTeacherClassDirectory` 是否只承接教师端班级目录的数据 owner。
- `useClassManagementPage` 是否只剩教师端 route / dialog 编排。
- `ClassManagementRoutePage.vue` 是否继续只负责 route page 组合。

## Rollback / Recovery

- 如果 `useTeacherClassDirectory` 的返回接口不顺手，可以调整字段组织，但班级目录的数据加载 owner 仍必须留在新 composable。
