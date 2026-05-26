# Reuse Decision

## Change type
structure_refactor

## Existing code searched
- `code/frontend/src/views/challenges/ChallengeDetail.vue`
- `code/frontend/src/components/challenge/ChallengeQuestionPanel.vue`
- `code/frontend/src/components/challenge/ChallengeSolutionsPanel.vue`
- `code/frontend/src/components/challenge/ChallengeSubmissionRecordsPanel.vue`
- `code/frontend/src/components/challenge/ChallengeWriteupPanel.vue`
- `code/frontend/src/components/challenge/ChallengeActionAside.vue`
- `code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue`
- `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`
- `code/frontend/src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue`
- `code/frontend/src/components/contests/ContestChallengeWorkspacePanel.vue`
- `code/frontend/src/components/contests/ContestTeamWorkspaceSection.vue`
- `code/frontend/src/components/contests/ContestAnnouncementsWorkspaceSection.vue`

## Decision
refactor_existing

## Reason
`ChallengeDetail.vue` 当前重量已经不在题目、题解、提交记录和实例工具的业务细节，这些内容都已经各自下沉到独立 panel / aside 组件。剩余混在路由页里的主要是 tabbar、主区切换、右侧工具栏装配和对应布局样式。这里适合继续沿用仓库里已经验证过的“父页保留 route/data/action owner，子组件承接稳定 workspace shell”模式，把路由页继续收敛成 `useChallengeDetailPage` 的组合 owner，而不是重新拆 feature model。

## Files to modify
- `code/frontend/src/views/challenges/ChallengeDetail.vue`
- `code/frontend/src/components/challenge/ChallengeWorkspaceShell.vue`
- `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`
- `code/frontend/src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果这一刀完成后 `ChallengeDetail.vue` 仍超过 route view 大小阈值，下一步优先继续判断 `loading / error / empty` 状态壳是否应独立到 challenge state shell，而不是把 `useChallengeDetailPage` 的数据和交互 owner 继续切碎。
