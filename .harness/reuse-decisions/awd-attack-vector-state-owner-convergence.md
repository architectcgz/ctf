# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/features/contest-awd-workspace/model/useContestAWDWorkspace.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackSubmission.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdDefenseServiceSelection.ts`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/features/contest-awd-workspace/model/useAwdDefenseServiceSelection.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackSubmission.ts`
- `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts`

## Decision
refactor_existing

## Reason
`ContestAWDWorkspacePanel.vue` 在模板壳拆分后，最集中的剩余 owner 密度落在攻击向量脚本：challenge 选择、目标筛选、Flag 输入缓存、活跃目标派生和提交后清空输入都还堆在页面里。这组逻辑已经形成单独能力域，最小安全切片是把它收口到 `useAwdWorkspaceAttackVector.ts`，由 composable 持有本地 UI state 和派生，而父页继续承接战场级布局、`useContestAWDWorkspace` 主数据源和外层装配。

## Files to modify
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackVector.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackVector.test.ts`
- `code/frontend/src/features/contest-awd-workspace/model/index.ts`
- `code/frontend/src/features/contest-awd-workspace/index.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续继续收口 AWD 工作台，优先处理防守 SSH / 复制链路和事件标题映射这两组剩余 script owner。
