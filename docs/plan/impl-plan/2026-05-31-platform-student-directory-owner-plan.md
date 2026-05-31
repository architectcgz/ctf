# platform student directory owner 收口计划

## Objective

- 在 `platform/student-management` 内新增 `usePlatformStudentDirectory`，承接平台端学生目录的数据 owner。
- 让 `usePlatformStudentManagementPage` 只保留平台端 route 和 rows 派生语义。

## Non-goals

- 不改 teacher student management。
- 不把 teacher / platform 合成新的 shared student management page owner。
- 不调整 `StudentManageRoutePage.vue` 或相关 UI 组件的结构。

## Source Inputs

- `code/frontend/src/features/platform/student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/pages/platform/__tests__/StudentManage.test.ts`
- `code/frontend/src/features/teacher/student-management/model/useTeacherStudentDirectory.ts`
- `code/frontend/src/features/student-directory/model/useStudentDirectoryQuery.ts`
- `code/frontend/src/features/student-directory/model/useStudentFilters.ts`

## Plan Review Result

- 平台端学生管理适合沿用刚刚在教师端验证过的 `page + directory` 形态。
- `usePlatformStudentDirectory` 负责 classes、keyword/class filter、pagination 和 request orchestration。
- `usePlatformStudentManagementPage` 保留 rows 派生和学员分析 route builder，更符合页面语义 owner。

## Task Slices

### Slice 1: 新建 usePlatformStudentDirectory

- 目标：收口平台端学生目录的数据状态、筛选、初始化和分页。
- 风险：
  - 如果把 rows / route 派生一起搬走，会把平台页面语义混进目录状态 owner。

### Slice 2: usePlatformStudentManagementPage 改为消费 directory owner

- 目标：保留平台端 rows、统计派生和 route builder。
- 风险：
  - 如果 page 继续直接依赖 `getClasses` / `getStudentsDirectory`，就没有真正收口目录 owner。

### Slice 3: 更新源码级和行为测试

- 目标：给新目录 owner 补直测，并更新平台端源码断言。
- 风险：
  - 不补 keyword debounce 和 class filter 测试，后面还会回流进 page owner。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision platform-student-directory-owner`
- `npm run test:run -- src/features/platform/student-management/model/usePlatformStudentDirectory.test.ts src/pages/platform/__tests__/StudentManage.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `npm run typecheck`
- `git diff --check`
- `bash scripts/check-reuse-first.sh`

## Review Focus

- `usePlatformStudentDirectory` 是否只承接平台端学生目录的数据 owner。
- `usePlatformStudentManagementPage` 是否只剩平台端 rows / route 派生。
- 平台 route page 是否继续只负责 feature 组合。

## Rollback / Recovery

- 如果 `usePlatformStudentDirectory` 的返回接口不顺手，可以调整字段组织，但学生目录的数据加载和筛选 owner 仍必须留在新 composable。
