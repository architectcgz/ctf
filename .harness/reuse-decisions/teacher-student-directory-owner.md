# Reuse Decision

## Change type
frontend refactor / teacher student directory owner convergence

## Existing code searched
- `code/frontend/src/features/teacher/student-management/model/useStudentManagementPage.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentManagement.test.ts`
- `code/frontend/src/features/student-directory/model/useStudentDirectoryQuery.ts`
- `code/frontend/src/features/student-directory/model/useStudentFilters.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewDirectory.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailData.ts`

## Similar implementations found
- `useStudentManagementPage.ts` 当前同时持有班级加载、目录请求编排、筛选归一化、分页和 route builder。
- `useStudentDirectoryQuery.ts` 已经承接了学生目录的请求时序和 debounce 底座。
- 刚完成的 `useAwdReviewDirectory`、`useAwdReviewDetailData` 已经证明先拆出更小的数据 owner，再保留 page orchestration，是当前迁移的稳定做法。

## Decision
refactor_existing

## Reason
当前最小正确切片是把教师端学生管理拆成：

- `useTeacherStudentDirectory`：承接 classes、filters、目录请求、初始化、分页和搜索归一化。
- `useStudentManagementPage`：保留 route builder、班级管理 route 和报告弹窗 state。

这样可以：

- 去掉 `useStudentManagementPage` 里混合的目录状态 owner
- 保留教师端页面自己的 route / dialog 语义，不把 teacher / platform 两套策略硬抽成 shared

本轮不做：

- 不改平台端 `usePlatformStudentManagementPage`
- 不新增跨 teacher / platform 的 shared student management owner
- 不改 `StudentManagementPage.vue` 模板和交互文案

## Files to modify
- `.harness/reuse-decisions/teacher-student-directory-owner.md`
- `docs/plan/impl-plan/2026-05-31-teacher-student-directory-owner-plan.md`
- `code/frontend/src/features/teacher/student-management/model/useTeacherStudentDirectory.ts`
- `code/frontend/src/features/teacher/student-management/model/useTeacherStudentDirectory.test.ts`
- `code/frontend/src/features/teacher/student-management/model/useStudentManagementPage.ts`
- `code/frontend/src/features/teacher/student-management/model/index.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentManagement.test.ts`

## After implementation
- 教师端学生目录的数据加载、筛选和分页 owner 会集中到 `useTeacherStudentDirectory`。
- `useStudentManagementPage` 只保留教师端页面壳自己的 route / dialog 编排。
