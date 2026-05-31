# Reuse Decision

## Change type
frontend refactor / awd review directory owner convergence

## Existing code searched
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndexPage.ts`
- `code/frontend/src/pages/awd-review/__tests__/PlatformAWDReviewIndex.test.ts`
- `code/frontend/src/pages/awd-review/__tests__/TeacherAWDReviewIndex.test.ts`
- `code/frontend/src/features/managed-instance-directory/model/useManagedInstanceDirectory.ts`

## Similar implementations found
- `useAwdReviewIndex.ts` 自己持有 `listAwdReviewsByRole + useAbortController + latestRequestId + debounce + page/pageSize/summary`。
- 刚完成的 `managed-instance-directory` 已经证明这种目录状态 owner 可以先收口到更小的业务 owner，再由页面层消费。
- AWD 复盘目录的差异主要是 AWD 自己的 summary、status label 和 row 映射，不适合直接并进 `managed-instance-directory`。

## Decision
refactor_existing

## Reason
当前最小正确切片是把 `useAwdReviewIndex` 拆成：

- `useAwdReviewDirectory`：承接 AWD 复盘目录状态 owner
- `useAwdReviewIndex`：保留 AWD 复盘自己的 summary、rows、status options 和文案语义

这样可以：

- 去掉 `useAwdReviewIndex` 里混合的目录请求时序 owner
- 保留 AWD review feature 的业务边界，不把它泛化成另一个 shared 目录工具

本轮不做：

- 不改 AWD detail / export flow
- 不改 teacher/platform 两个 route page 的 route target owner
- 不把 AWD 目录 owner 提升到 shared 或跨 feature 通用层

## Files to modify
- `.harness/reuse-decisions/awd-review-directory-owner.md`
- `docs/plan/impl-plan/2026-05-31-awd-review-directory-owner-plan.md`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewDirectory.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewDirectory.test.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
- `code/frontend/src/features/awd-review-workspace/model/index.ts`
- `code/frontend/src/pages/awd-review/__tests__/PlatformAWDReviewIndex.test.ts`
- `code/frontend/src/pages/awd-review/__tests__/TeacherAWDReviewIndex.test.ts`

## After implementation
- AWD 复盘目录的请求时序、分页和 debounce 会集中到 `useAwdReviewDirectory`。
- `useAwdReviewIndex` 只保留 AWD 复盘的页面语义和展示派生。
