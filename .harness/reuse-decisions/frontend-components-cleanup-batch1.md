# Reuse Decision

## Change type
+component / feature / docs / test

## Existing code searched
- `code/frontend/src/components/platform/PlatformPaginationControls.vue`
- `code/frontend/src/components/common/PagePaginationControls.vue`
- `code/frontend/src/components/platform/__tests__/PlatformPaginationControls.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts`
- `code/frontend/src/features/audit-log/ui/AuditLogDirectoryPanel.vue`
- `code/frontend/src/features/image-management/ui/ImageDirectoryPanel.vue`
- `code/frontend/src/features/platform/user-management/ui/UserGovernanceOverviewPanel.vue`
- `code/frontend/src/features/platform/contests/ui/PlatformContestTable.vue`
- `code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupDirectorySection.vue`
- `code/frontend/src/features/awd-inspector/ui/AWDTrafficEventTable.vue`
- `code/frontend/src/pages/platform/__tests__/platformManagementSurfaceAlignment.test.ts`

## Similar implementations found
- `PagePaginationControls.vue` 已经是通用分页原语，`PlatformPaginationControls.vue` 只是透传 `show-jump` 的薄壳，没有独立平台 owner。
- `AWDOperationsPanel` 运行时 owner 已经稳定落在 `features/contest-awd-admin/ui`，测试继续留在 `components/platform/__tests__` 只会制造旧 owner 仍存在的错觉。

## Decision
refactor_existing

## Reason
- 这次不是把 `components` 整体改名成 `shared`，而是先清理其中最明显的假 owner。
- `PlatformPaginationControls.vue` 没有业务语义，直接让 consumer 依赖 `components/common/PagePaginationControls.vue` 更符合当前分层。
- `components/platform/__tests__/AWDOperationsPanel.test.ts` 应迁到 feature 邻近，避免历史目录继续承担活动 owner。

## Files to modify
- `.harness/reuse-decisions/frontend-components-cleanup-batch1.md`
- `code/frontend/src/features/audit-log/ui/AuditLogDirectoryPanel.vue`
- `code/frontend/src/features/image-management/ui/ImageDirectoryPanel.vue`
- `code/frontend/src/features/platform/user-management/ui/UserGovernanceOverviewPanel.vue`
- `code/frontend/src/features/platform/contests/ui/PlatformContestTable.vue`
- `code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupDirectorySection.vue`
- `code/frontend/src/features/awd-inspector/ui/AWDTrafficEventTable.vue`
- `code/frontend/src/pages/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `code/frontend/src/components/platform/__tests__/PlatformPaginationControls.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDRoundInspector.test.ts`
- `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts`
- `code/frontend/src/components/platform/PlatformPaginationControls.vue`
- `code/frontend/src/pages/platform/ImageManageRoutePage.vue`

## After implementation
- 如果后续继续清理 `components/*`，这一批文件应作为“先删除假 owner，再讨论目录改名”的基准案例。
