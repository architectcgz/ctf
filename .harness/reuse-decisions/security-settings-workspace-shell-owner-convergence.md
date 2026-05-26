# Reuse Decision

## Change type
structure_refactor

## Existing code searched
- `code/frontend/src/views/profile/SecuritySettings.vue`
- `code/frontend/src/views/profile/__tests__/SecuritySettings.test.ts`
- `code/frontend/src/components/profile/UserProfileWorkspaceShell.vue`
- `code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/profile/UserProfileWorkspaceShell.vue`
- `code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue`

## Decision
refactor_existing

## Reason
`SecuritySettings.vue` 当前 491 行，已经接近 route view 行数护栏，而且本轮重量主要来自稳定的页面壳、密码表单模板和局部样式。脚本层 owner 已经集中在 `useSecuritySettingsPage()`，最小安全切片是沿用既有 profile workspace shell 模式：父页继续持有密码修改流程、校验、提交和安全概况数据 owner，新子组件只承接稳定模板、样式和事件转发。

## Files to modify
- `code/frontend/src/views/profile/SecuritySettings.vue`
- `code/frontend/src/components/profile/SecuritySettingsWorkspaceShell.vue`
- `code/frontend/src/views/profile/__tests__/SecuritySettings.test.ts`
- `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/views/__tests__/surfaceBackground.test.ts`
- `code/frontend/src/views/__tests__/rootHeroLayout.test.ts`
- `code/frontend/src/views/__tests__/studentRootShellCleanup.test.ts`
- `code/frontend/src/views/__tests__/profileJournalButtonStyles.test.ts`
- `code/frontend/src/views/__tests__/journalEyebrowStyles.test.ts`
- `code/frontend/src/views/__tests__/profileJournalNoteStyles.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/views/__tests__/journalUserShellStyles.test.ts`
- `code/frontend/src/views/__tests__/profileJournalUtilityStyles.test.ts`
- `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果这一刀后 `SecuritySettingsWorkspaceShell.vue` 继续累积新的展示分区，下一步优先沿“安全概况 / 密码修改 / 安全提示”继续拆成更细的展示区，而不是把 `useSecuritySettingsPage()` 的校验、提交或状态 owner 再抬回父页。
