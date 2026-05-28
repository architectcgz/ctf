# Reuse Decision

## Change type
frontend bugfix / image management duplicate action owner cleanup / local in-flight guard convergence

## Existing code searched
- code/frontend/src/views/platform/ImageManage.vue
- code/frontend/src/views/platform/__tests__/ImageManage.test.ts
- code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts
- code/frontend/src/features/image-management/model/useImageManagePage.ts
- code/frontend/src/features/image-management/model/useImageManageMutations.ts
- code/frontend/src/components/platform/images/ImageCreateModal.vue
- code/frontend/src/components/platform/images/ImageDirectoryPanel.vue
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `useImageManageMutations.ts` 里的 `handleCreate()` 已经通过 `creating.value` 在本地 owner 上短路重复提交。
- `duplicateActionGuardAudit.test.ts` 已把图片管理页纳入重复动作审计，但当前只覆盖创建动作，没有覆盖删除动作。
- `ImageDirectoryPanel.vue` 的删除按钮目前直接透传 `delete-image` 事件，没有消费任何删除中的状态。

## Decision
refactor_existing

## Reason
这次不需要重做图片管理页结构，也不应该把 guard 放到按钮 disabled 或确认框实现里。最小正确改动是让 `useImageManageMutations.ts` 成为唯一的删除动作 owner：在本地持有 per-image in-flight guard，并把删除中状态显式传给列表按钮。

## Files to modify
- .harness/reuse-decisions/image-manage-duplicate-submit-owner-cleanup.md
- docs/plan/impl-plan/2026-05-28-image-manage-duplicate-submit-owner-cleanup-implementation-plan.md
- docs/reviews/frontend/2026-05-28-image-manage-duplicate-submit-owner-cleanup-review.md
- code/frontend/src/features/image-management/model/useImageManageMutations.ts
- code/frontend/src/features/image-management/model/useImageManagePage.ts
- code/frontend/src/components/platform/images/ImageDirectoryPanel.vue
- code/frontend/src/views/platform/ImageManage.vue
- code/frontend/src/views/platform/__tests__/ImageManage.test.ts
- code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- 图片创建与图片删除都由 `useImageManageMutations.ts` 在本地 owner 上持有重复动作短路。
- 同一镜像的删除动作在确认弹窗阶段和实际删除请求阶段都不会重复进入。
- 列表删除按钮会消费删除中的显式状态，而不是仅靠用户手速或确认框兜底。
