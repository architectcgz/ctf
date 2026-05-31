# Reuse Decision

## Change type
frontend refactor / managed instance directory state owner convergence

## Existing code searched
- `code/frontend/src/features/teacher/instances/model/useInstances.ts`
- `code/frontend/src/features/platform/instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
- `code/frontend/src/shared/model/common/usePagination.ts`
- `code/frontend/src/pages/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/pages/platform/__tests__/InstanceManage.test.ts`

## Similar implementations found
- teacher / platform 两边都各自维护 `getInstanceDirectoryByRole + useAbortController + latestRequestId + debounce search + page/pageSize/total`。
- `usePagination` 只能覆盖基础分页，不覆盖 role-aware query、summary 回填和目录型 debounce owner。
- `useAwdReviewIndex`、`useStudentDirectoryQuery` 也有相似的 stale request / debounce 结构，但不是 managed instance directory 的 owner。

## Decision
refactor_existing

## Reason
当前最小正确切片不是继续在 teacher / platform 各自维护目录加载逻辑，而是把 managed instance directory 的状态 owner 收口成共享 feature：

- 两边都依赖 `getInstanceDirectoryByRole`
- 两边都维护相同的 stale request guard 和 abort cleanup
- 两边都在处理页码、pageSize、totalPages、debounce 搜索和统一错误状态
- 差异主要只剩 teacher 的班级初始化和 platform 的状态筛选 / 行映射 / summary 文案

因此本轮应：

- 新建 `features/managed-instance-directory`
- 让它承接目录加载、分页切换、debounce 搜索和 stale request 防护
- teacher / platform 只保留各自 filter shape、summary 回填和 route/page 表达

本轮不做：

- 不改 managed destroy workflow
- 不改普通 instance workflow
- 不把 class 初始化、行映射、状态文案或 route 逻辑抽进共享层

## Files to modify
- `.harness/reuse-decisions/managed-instance-directory-owner.md`
- `docs/plan/impl-plan/2026-05-31-managed-instance-directory-owner-plan.md`
- `code/frontend/src/features/managed-instance-directory/index.ts`
- `code/frontend/src/features/managed-instance-directory/model/index.ts`
- `code/frontend/src/features/managed-instance-directory/model/useManagedInstanceDirectory.ts`
- `code/frontend/src/features/managed-instance-directory/model/useManagedInstanceDirectory.test.ts`
- `code/frontend/src/features/teacher/instances/model/useInstances.ts`
- `code/frontend/src/features/platform/instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/pages/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/pages/platform/__tests__/InstanceManage.test.ts`

## After implementation
- teacher / platform 不再各自维护 managed instance directory 的请求时序、debounce 和分页 owner。
- `managed-instance-directory` 会成为目录状态的共享 owner。
- 页面本地只保留班级默认值、状态筛选、summary 和表格行表达差异。
