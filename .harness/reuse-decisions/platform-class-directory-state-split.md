# Reuse Decision

## Change type
frontend refactor / platform class directory state split

## Existing code searched
- `code/frontend/src/features/platform/class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/pages/platform/__tests__/ClassManage.test.ts`
- `code/frontend/src/features/platform/student-management/model/usePlatformStudentDirectory.ts`
- `code/frontend/src/features/teacher/class-management/model/useTeacherClassDirectory.ts`

## Similar implementations found
- `usePlatformClassManagementPage.ts` 当前同时持有班级目录请求、分页、错误状态、行映射和 route builder。
- `usePlatformStudentDirectory.ts` 已经证明 platform 侧目录状态可以独立收口到 directory owner。
- `useTeacherClassDirectory.ts` 刚完成同类拆分，说明 class management 也适合 `page + directory` 结构。

## Decision
refactor_existing

## Reason
当前最小正确切片是把平台端班级管理拆成：

- `usePlatformClassDirectory`：承接班级目录加载、分页和错误状态。
- `usePlatformClassManagementPage`：保留班级行映射、详情 route 和概览指标派生。

这样可以：

- 去掉 `usePlatformClassManagementPage` 里混合的目录数据 owner
- 保持平台端班级页自己的列表行结构和路由语义

本轮不做：

- 不改 teacher class management
- 不新增 teacher / platform 共用 class management page owner
- 不调整平台班级管理页 UI

## Files to modify
- `.harness/reuse-decisions/platform-class-directory-state-split.md`
- `docs/plan/impl-plan/2026-05-31-platform-class-directory-state-split-plan.md`
- `code/frontend/src/features/platform/class-management/model/usePlatformClassDirectory.ts`
- `code/frontend/src/features/platform/class-management/model/usePlatformClassDirectory.test.ts`
- `code/frontend/src/features/platform/class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform/class-management/model/index.ts`
- `code/frontend/src/pages/platform/__tests__/ClassManage.test.ts`

## After implementation
- 平台端班级目录的数据加载和分页 owner 会集中到 `usePlatformClassDirectory`。
- `usePlatformClassManagementPage` 只保留后台班级页的行映射、指标和 route 编排。
