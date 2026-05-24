# Reuse Decision

## Change type
feature / composable / test / docs

## Existing code searched
- `code/frontend/src/features/platform-users`
- `code/frontend/src/features/platform-user-management`
- `code/frontend/src/composables/usePlatformUsers.ts`
- `code/frontend/src/features/__tests__/featureBoundaries.test.ts`
- `docs/plan/impl-plan/2026-05-24-platform-user-management-feature-split-implementation-plan.md`

## Similar implementations found
- `code/frontend/src/features/platform-users/index.ts`
- `code/frontend/src/features/platform-users/model/index.ts`
- `code/frontend/src/features/platform-users/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-users/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform-users/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/platform-users/model/usePlatformUserManagePage.ts`
- `code/frontend/src/features/platform-users/model/usePlatformUsers.ts`
- `code/frontend/src/composables/usePlatformUsers.ts`
- `code/frontend/src/features/platform-user-management/index.ts`
- `code/frontend/src/features/platform-user-management/model/usePlatformUserManagePage.ts`
- `code/frontend/src/features/platform-user-management/model/usePlatformUsers.ts`

## Decision
refactor_existing

## Reason
这次不再保留空桥。`platform-user-management` 已经承接真实用户治理 owner，而 `platform-users` 与 `composables/usePlatformUsers.ts` 只剩转发壳，并且当前 runtime 搜索没有发现真实引用。继续保留这些空壳，只会让 review 中记录的 over-broad bucket 继续留在磁盘结构上，增加搜索噪音和边界判断成本。更低风险的收口方式是直接删除这些已无引用的桥接壳，同时同步边界测试和实施计划，让结构事实与当前 owner 保持一致。

## Files to modify
- `docs/plan/impl-plan/2026-05-24-platform-users-bridge-removal-implementation-plan.md`
- `docs/plan/impl-plan/2026-05-24-platform-user-management-feature-split-implementation-plan.md`
- `code/frontend/src/features/__tests__/featureBoundaries.test.ts`
- `code/frontend/src/features/platform-users/**`
- `code/frontend/src/composables/usePlatformUsers.ts`

## After implementation
- 后续若还要继续缩减其他 legacy bridge，沿用同样模式：先确认真实 owner 已落地且无 runtime 引用，再删桥并同步边界测试，不再长期保留空转发壳。
