# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceHudStrip.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseColumn.vue`
- `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspacePresentation.ts`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`

## Similar implementations found
- `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspacePresentation.ts`

## Decision
refactor_existing

## Reason
`ContestAWDWorkspacePanel.vue` 经过前几刀后，剩下最成组的脚本 owner 已经集中在 HUD 摘要和防守告警派生：回合状态标签、战队排名、服务数、最高分、同步时间，以及 `defenseAlerts` 这组展示数据都还在父页本地计算。这些逻辑只消费现成 workspace / scoreboard / runtime challenge 数据，不持有副作用，也不应该继续堆在父页里。最小安全切片是新增 `useAwdWorkspaceSummary.ts` 收口这组 summary / alert presentation，让父页只保留工作区数据获取、攻击提交流程和防守操作装配 owner。

## Files to modify
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceSummary.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceSummary.test.ts`
- `code/frontend/src/features/contest-awd-workspace/model/index.ts`
- `code/frontend/src/features/contest-awd-workspace/index.ts`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果 `ContestAWDWorkspacePanel.vue` 在这一刀后只剩装配、布局和副作用 owner，可把 AWD 页面这条 `TD-1` 标记为当前 touched surface 已收口。
