# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue`
- `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts`
- `code/frontend/src/features/contest-awd-workspace/model/awdChallengeIdentity.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdDefenseAccessPanel.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackVector.ts`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`

## Similar implementations found
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackVector.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdDefenseAccessPanel.ts`
- `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts`

## Decision
refactor_existing

## Reason
`ContestAWDWorkspacePanel.vue` 当前剩余较成组的脚本 owner 已集中到情报 / 结果文案 presentation：challenge 映射、事件标题解析、方向/结果标签、服务引用文案，以及攻击结果 toast 文案格式化都还堆在父页。这组逻辑不持有远端副作用，最小安全切片是把它收口到 `useAwdWorkspacePresentation.ts`，父页继续保留工作区主数据和布局装配。实现过程中还发现 AWD runtime challenge 身份不能继续回退到 `challenge_id`，因此同刀把 AWD 工作台 touched surface 统一收紧为 `awd_service_id + awd_challenge_id` 专用语义，避免再混用历史 challenge 标识。

## Files to modify
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/features/contest-awd-workspace/model/awdChallengeIdentity.ts`
- `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts`
- `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.test.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackVector.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspacePresentation.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspacePresentation.test.ts`
- `code/frontend/src/features/contest-awd-workspace/model/index.ts`
- `code/frontend/src/features/contest-awd-workspace/index.ts`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果继续收口 AWD 工作台，只剩少量战场摘要标签和局部摘要 / presentation 尾项，可视情况决定是否再拆。
