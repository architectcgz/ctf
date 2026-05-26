# Reuse Decision

## Change type
structure_refactor

## Existing code searched
- `code/frontend/src/views/instances/InstanceList.vue`
- `code/frontend/src/views/instances/__tests__/InstanceList.test.ts`
- `code/frontend/src/components/profile/SecuritySettingsWorkspaceShell.vue`
- `code/frontend/src/components/profile/UserProfileWorkspaceShell.vue`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/profile/SecuritySettingsWorkspaceShell.vue`
- `code/frontend/src/components/profile/UserProfileWorkspaceShell.vue`

## Decision
refactor_existing

## Reason
`InstanceList.vue` 当前 493 行，已经接近 route view 护栏，页面业务 owner 已主要集中在 `useInstanceListPage()` 与 `useInstanceWarningFocus()`。剩余重量主要来自实例概况、目录表格、过期提醒弹层和对应局部样式。最小安全切片是沿用既有 workspace shell 模式：父页继续持有数据加载、定时刷新、延时/销毁/打开/复制等主动作、过期提醒状态和按钮聚焦 owner，新 shell 只承接稳定模板、局部样式和按钮 ref 透传。

## Files to modify
- `code/frontend/src/views/instances/InstanceList.vue`
- `code/frontend/src/components/instance/InstanceListWorkspaceShell.vue`
- `code/frontend/src/views/instances/__tests__/InstanceList.test.ts`
- `code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts`
- `code/frontend/src/views/__tests__/journalUserShellStyles.test.ts`
- `code/frontend/src/views/__tests__/journalEyebrowStyles.test.ts`
- `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/views/__tests__/journalUserDirectoryButtonVariants.test.ts`
- `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `code/frontend/src/views/__tests__/studentDirectoryTypographyBoundary.test.ts`
- `code/frontend/src/views/__tests__/rootHeroLayout.test.ts`
- `code/frontend/src/views/__tests__/studentRootShellCleanup.test.ts`
- `code/frontend/src/views/__tests__/journalUserDirectoryStyles.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果这一刀后 `InstanceListWorkspaceShell.vue` 继续增长，下一步优先沿“概况 / 目录列表 / 过期提醒弹层”继续拆成更细展示区，而不是把加载、定时刷新或实例操作 owner 再抬回 route view。
