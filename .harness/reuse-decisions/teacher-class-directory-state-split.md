# Reuse Decision

## Change type
frontend refactor / teacher class directory state split

## Existing code searched
- `code/frontend/src/features/teacher/class-management/model/useClassManagementPage.ts`
- `code/frontend/src/pages/teacher/__tests__/ClassManagement.test.ts`
- `code/frontend/src/features/teacher/student-management/model/useTeacherStudentDirectory.ts`
- `code/frontend/src/features/platform/class-management/model/usePlatformClassManagementPage.ts`

## Similar implementations found
- `useClassManagementPage.ts` 当前同时持有班级目录请求、分页、错误状态、route builder 和导出弹窗状态。
- `useTeacherStudentDirectory.ts` 已经证明 teacher 侧目录状态可以先拆到更小的 directory owner，再让 page model 退成壳。
- 平台端 class management 也有相同趋势，但 teacher 端额外带有报告弹窗壳，适合先独立收口。

## Decision
refactor_existing

## Reason
当前最小正确切片是把教师端班级管理拆成：

- `useTeacherClassDirectory`：承接班级目录加载、分页和错误状态。
- `useClassManagementPage`：保留 route builder、概览路由和报告弹窗状态。

这样可以：

- 去掉 `useClassManagementPage` 里混合的目录数据 owner
- 保持 teacher class management 自己的 route / dialog 语义清楚

本轮不做：

- 不改 platform class management
- 不新增 teacher / platform 共用 class management page owner
- 不调整班级管理页 UI

## Files to modify
- `.harness/reuse-decisions/teacher-class-directory-state-split.md`
- `docs/plan/impl-plan/2026-05-31-teacher-class-directory-state-split-plan.md`
- `code/frontend/src/features/teacher/class-management/model/useTeacherClassDirectory.ts`
- `code/frontend/src/features/teacher/class-management/model/useTeacherClassDirectory.test.ts`
- `code/frontend/src/features/teacher/class-management/model/useClassManagementPage.ts`
- `code/frontend/src/features/teacher/class-management/model/index.ts`
- `code/frontend/src/pages/teacher/__tests__/ClassManagement.test.ts`

## After implementation
- 教师端班级目录的数据加载和分页 owner 会集中到 `useTeacherClassDirectory`。
- `useClassManagementPage` 只保留教师端页面壳相关的 route / dialog 编排。
