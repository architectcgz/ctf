# Reuse Decision

## Change type
structure_refactor

## Existing code searched
- `code/frontend/src/views/profile/SkillProfile.vue`
- `code/frontend/src/views/profile/__tests__/SkillProfile.test.ts`
- `code/frontend/src/views/profile/__tests__/skillProfileTabsAdoption.test.ts`
- `code/frontend/src/components/profile/UserProfileWorkspaceShell.vue`
- `code/frontend/src/components/scoreboard/ScoreboardWorkspaceShell.vue`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/profile/UserProfileWorkspaceShell.vue`
- `code/frontend/src/components/scoreboard/ScoreboardWorkspaceShell.vue`

## Decision
refactor_existing

## Reason
`SkillProfile.vue` 当前 741 行，仍在 oversized route view allowlist 中。脚本层 owner 已经主要集中在 `useSkillProfilePage()` 和 `useUrlSyncedTabs<SkillProfileTabKey>()`，页面本体主要承接 tab rail、教师筛选区、三块内容面板和对应局部样式。最小安全切片是沿用既有 workspace shell 模式：父页继续持有 route tab、学员切换、远端数据加载、推荐跳转和错误 owner，新子组件只承接稳定模板、样式和事件转发。

## Files to modify
- `code/frontend/src/views/profile/SkillProfile.vue`
- `code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/profile/__tests__/SkillProfile.test.ts`
- `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/views/__tests__/rootHeroLayout.test.ts`
- `code/frontend/src/views/__tests__/studentRootShellCleanup.test.ts`
- `code/frontend/src/views/__tests__/profileJournalButtonStyles.test.ts`
- `code/frontend/src/views/__tests__/profileJournalUtilityStyles.test.ts`
- `code/frontend/src/views/__tests__/profileJournalNoteStyles.test.ts`
- `code/frontend/src/views/__tests__/journalUserShellStyles.test.ts`
- `code/frontend/src/views/__tests__/journalEyebrowStyles.test.ts`
- `code/frontend/src/views/__tests__/pageTabsStyles.test.ts`
- `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果这一刀后 `SkillProfile.vue` 仍然保留明显的展示壳重量，下一步优先继续把分析面板、薄弱维度面板、推荐面板各自拆成更细的展示分区，而不是把 `useSkillProfilePage()` 的加载、推荐或跳转 owner 再拆散。
