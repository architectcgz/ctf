# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseColumn.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseConnectionPanel.vue`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAccessActions.ts`
- `code/frontend/src/features/contest-awd-workspace/model/sshAccessPresentation.ts`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`

## Similar implementations found
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackVector.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdDefenseServiceSelection.ts`
- `code/frontend/src/features/contest-awd-workspace/model/sshAccessPresentation.ts`

## Decision
refactor_existing

## Reason
`ContestAWDWorkspacePanel.vue` 当前剩余较集中的脚本 owner 是防守 SSH / 复制链路：选中服务对应的 access 解析、复制状态、打开本队服务，以及剪贴板失败提示都还堆在父页。这组逻辑已经形成独立能力域，最小安全切片是把它收口到 `useAwdDefenseAccessPanel.ts`，父页继续保留服务选择和整页装配。

## Files to modify
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdDefenseAccessPanel.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdDefenseAccessPanel.test.ts`
- `code/frontend/src/features/contest-awd-workspace/model/index.ts`
- `code/frontend/src/features/contest-awd-workspace/index.ts`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果继续收口 AWD 工作台，优先处理事件标题映射和 attack toast 格式化这组情报侧 presentation owner。
