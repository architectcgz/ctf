# teacher student directory owner 收口计划

## Objective

- 在 `teacher/student-management` 内新增 `useTeacherStudentDirectory`，承接教师端学生目录的数据 owner。
- 让 `useStudentManagementPage` 只保留教师端页面壳的 route / dialog 编排。

## Non-goals

- 不改平台端 student management。
- 不把 teacher / platform 两套学生目录策略合成新的 shared page owner。
- 不调整 `StudentManagementPage.vue` 的 UI 和交互结构。

## Source Inputs

- `code/frontend/src/features/teacher/student-management/model/useStudentManagementPage.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentManagement.test.ts`
- `code/frontend/src/features/student-directory/model/useStudentDirectoryQuery.ts`
- `code/frontend/src/features/student-directory/model/useStudentFilters.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewDirectory.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailData.ts`

## Plan Review Result

- 教师端学生管理的第一刀应该先做 `page + directory` 拆分，不直接把 teacher / platform 强行合并。
- `useTeacherStudentDirectory` 负责 classes、preferred class、search heuristic、pagination 和 request orchestration。
- `useStudentManagementPage` 保留 route builder、班级管理入口和报告弹窗 state，更符合页面壳 owner。

## Task Slices

### Slice 1: 新建 useTeacherStudentDirectory

- 目标：收口教师端学生目录的数据状态、筛选、初始化和分页。
- 风险：
  - 如果把 route 或 dialog 一起搬过去，会重新模糊 page owner。

### Slice 2: useStudentManagementPage 改为消费 directory owner

- 目标：保留教师端 route builder 和报告弹窗壳。
- 风险：
  - 如果 page 继续直接依赖 `getClasses` / `getStudentsDirectory`，就没有真正收口目录 owner。

### Slice 3: 更新源码级和行为测试

- 目标：给新目录 owner 补直测，并更新教师端源码断言。
- 风险：
  - 不补 preferred class / 搜索归一化测试，后面还会回流进 page owner。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision teacher-student-directory-owner`
- `npm run test:run -- src/features/teacher/student-management/model/useTeacherStudentDirectory.test.ts src/pages/teacher/__tests__/TeacherStudentManagement.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `npm run typecheck`
- `git diff --check`
- `bash scripts/check-reuse-first.sh`

## Review Focus

- `useTeacherStudentDirectory` 是否只承接教师端学生目录的数据 owner。
- `useStudentManagementPage` 是否只剩教师端 route / dialog 编排。
- `TeacherStudentManagementRoutePage.vue` 是否继续只负责 route page 组合。

## Rollback / Recovery

- 如果 `useTeacherStudentDirectory` 的返回接口不顺手，可以调整字段组织，但学生目录的数据加载和筛选 owner 仍必须留在新 composable。
