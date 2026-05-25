# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceHudStrip.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseAlertsPanel.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/contests/awd/AWDWorkspaceHudStrip.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue`
- `code/frontend/src/components/challenge/ChallengeSubmissionRecordsPanel.vue`
- `code/frontend/src/components/platform/images/ImageDirectoryPanel.vue`

## Decision
refactor_existing

## Reason
`ContestAWDWorkspacePanel.vue` 已经沿稳定展示区块连续抽出 HUD、情报栏和防守告警。中区顶部筛选条只负责题目选择和队伍关键字输入，不直接拥有目标打开、Flag 提交、结果提示或远端请求，适合继续按“父组件持有状态，子组件承接输入壳”的模式拆出，而不是把整个攻击区一起迁走。

## Files to modify
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDAttackToolbar.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续继续拆中区攻击工作台，补 `.harness/reuse-index/` 镜像索引，记录“attack workspace owner + local input shell child” 模式。
