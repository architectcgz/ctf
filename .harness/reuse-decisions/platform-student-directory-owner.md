# Reuse Decision

## Change type
frontend refactor / platform student directory owner convergence

## Existing code searched
- `code/frontend/src/features/platform/student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/pages/platform/__tests__/StudentManage.test.ts`
- `code/frontend/src/features/teacher/student-management/model/useTeacherStudentDirectory.ts`
- `code/frontend/src/features/student-directory/model/useStudentDirectoryQuery.ts`
- `code/frontend/src/features/student-directory/model/useStudentFilters.ts`

## Similar implementations found
- `usePlatformStudentManagementPage.ts` 当前同时持有班级加载、目录请求编排、筛选归一化、分页和 route builder。
- 刚完成的 `useTeacherStudentDirectory.ts` 已经把同类目录状态从 page owner 中拆了出来。
- platform 端虽然和 teacher 端同属学生目录，但默认班级、筛选行为和路由目标不同，不适合这轮直接合并成 shared page owner。

## Decision
refactor_existing

## Reason
当前最小正确切片是把平台端学生管理拆成：

- `usePlatformStudentDirectory`：承接 classes、filters、目录请求、初始化和分页。
- `usePlatformStudentManagementPage`：保留平台端 route builder、rows 派生和页面语义。

这样可以：

- 去掉 `usePlatformStudentManagementPage` 里混合的目录状态 owner
- 保持 teacher / platform 两侧各自的页面策略清晰，再决定是否需要进一步共用

本轮不做：

- 不改 teacher student management
- 不新增跨 teacher / platform 的 shared student management page owner
- 不调整平台学生管理页 UI

## Files to modify
- `.harness/reuse-decisions/platform-student-directory-owner.md`
- `docs/plan/impl-plan/2026-05-31-platform-student-directory-owner-plan.md`
- `code/frontend/src/features/platform/student-management/model/usePlatformStudentDirectory.ts`
- `code/frontend/src/features/platform/student-management/model/usePlatformStudentDirectory.test.ts`
- `code/frontend/src/features/platform/student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform/student-management/model/index.ts`
- `code/frontend/src/pages/platform/__tests__/StudentManage.test.ts`

## After implementation
- 平台端学生目录的数据加载、筛选和分页 owner 会集中到 `usePlatformStudentDirectory`。
- `usePlatformStudentManagementPage` 只保留平台端 rows / route 这类页面派生语义。
