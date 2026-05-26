# Reuse Decision

## Change type
structure_refactor

## Existing code searched
- `code/frontend/src/views/platform/ContestAwdConfig.vue`
- `code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue`
- `code/frontend/src/components/profile/UserProfileWorkspaceShell.vue`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue`
- `code/frontend/src/components/profile/UserProfileWorkspaceShell.vue`

## Decision
refactor_existing

## Reason
`ContestAwdConfig.vue` 当前 686 行，已经是剩余 oversized route view allowlist 的最后一页。脚本层 owner 已经主要集中在 `useContestAwdConfigPage()`，页面本体主要承接整个 AWD 配置工作台的模板和局部样式。最小安全切片是沿用既有 workspace shell 模式：父页继续持有路由、服务选择、预览、保存、草稿与错误 owner，新子组件只承接稳定的工作台布局模板、样式和事件转发。

## Files to modify
- `code/frontend/src/views/platform/ContestAwdConfig.vue`
- `code/frontend/src/components/platform/contest/ContestAwdConfigWorkspaceShell.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果这一刀后 `ContestAwdConfig.vue` 仍然保留过多壳层模板，下一步优先继续沿“配置画布 / 试跑调试台 / 底部动作区”拆更细的展示壳，而不是把 `useContestAwdConfigPage()` 的预览、保存或路由 owner 再拆散。
