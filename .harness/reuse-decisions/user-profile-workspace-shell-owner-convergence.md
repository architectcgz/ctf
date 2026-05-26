# Reuse Decision

## Change type
structure_refactor

## Existing code searched
- `code/frontend/src/views/profile/UserProfile.vue`
- `code/frontend/src/views/profile/__tests__/UserProfile.test.ts`
- `code/frontend/src/components/challenge/ChallengeWorkspaceShell.vue`
- `code/frontend/src/components/scoreboard/ScoreboardWorkspaceShell.vue`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/challenge/ChallengeWorkspaceShell.vue`
- `code/frontend/src/components/scoreboard/ScoreboardWorkspaceShell.vue`

## Decision
refactor_existing

## Reason
`UserProfile.vue` 当前 719 行，已经长期处于 oversized route view allowlist。脚本侧 owner 基本集中在 `useUserProfilePage()`，页面本身主要承接加载/错误/空态分支、顶部资料壳、摘要区、账号信息区、报告区和对应局部样式。最小安全切片是沿用既有 workspace shell 模式：父页继续持有 profile 数据、报告导出、下载和错误 owner，新子组件只承接稳定的页面模板和事件转发。

## Files to modify
- `code/frontend/src/views/profile/UserProfile.vue`
- `code/frontend/src/components/profile/UserProfileWorkspaceShell.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/profile/__tests__/UserProfile.test.ts`
- `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/views/__tests__/surfaceBackground.test.ts`
- `code/frontend/src/views/__tests__/journalUserShellStyles.test.ts`
- `code/frontend/src/views/__tests__/profileJournalButtonStyles.test.ts`
- `code/frontend/src/views/__tests__/rootHeroLayout.test.ts`
- `code/frontend/src/views/__tests__/studentRootShellCleanup.test.ts`
- `code/frontend/src/views/__tests__/profileJournalUtilityStyles.test.ts`
- `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `code/frontend/src/views/__tests__/profileJournalNoteStyles.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/views/__tests__/journalUserDirectoryStyles.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果这一刀后 `UserProfile.vue` 仍然保留过多壳层模板或继续处于 oversized allowlist，下一步优先继续把账号信息区和报告区拆成更细的展示分区，而不是把 `useUserProfilePage()` 的数据/导出 owner 再拆散。
