# Reuse Decision

## Change type
- component
- layout

## Existing code searched
- code/frontend/src/components/challenge/ChallengeInstanceCard.vue
- code/frontend/src/components/challenge/ChallengeActionAside.vue
- code/frontend/src/views/challenges/ChallengeDetail.vue
- code/frontend/src/views/challenges/__tests__/challengeDetailSharedShell.test.ts
- code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts

## Similar implementations found
- code/frontend/src/components/challenge/ChallengeInstanceCard.vue
- code/frontend/src/components/challenge/ChallengeActionAside.vue
- code/frontend/src/views/challenges/__tests__/challengeDetailSharedShell.test.ts

## Decision
extend_existing

## Reason
这次不是新增题目页实例卡，也不是拆新的文案组件。继续复用现有 `ChallengeInstanceCard` 的 `instance-note` 区块和共享 / 独享实例分支，只删除一条实现说明式提示文案，并在现有题目详情壳层测试里补回归断言，改动最小，也符合“可见 UI 不渲染实现说明”的既有规则。

## Files to modify
- code/frontend/src/components/challenge/ChallengeInstanceCard.vue
- code/frontend/src/views/challenges/__tests__/challengeDetailSharedShell.test.ts
