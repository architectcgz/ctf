# Reuse Decision

## Change type
frontend refactor / entity contract boundary normalization

## Existing code searched
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/entities/challenge/model/index.ts
- code/frontend/src/entities/challenge/model/presentation.ts
- code/frontend/src/entities/challenge/ui/*.vue
- code/frontend/src/features/platform-challenge-detail/ui/AdminChallengeProfilePanel.vue
- code/frontend/src/views/challenges/__tests__/ChallengeList.test.ts
- code/frontend/src/views/platform/__tests__/challengeDetailPanelExtraction.test.ts

## Similar implementations found
- 当前仓库前端的 page / feature UI 收口已经多次把 API DTO 与展示层契约分离，页面或 feature 继续持有 API shape，entities / shared UI 只消费自己真正需要的字段。
- `entities/challenge` 目前仍直接从 `@/api/contracts` 取 category / difficulty / challenge DTO 类型，属于同一类还没完全收口的展示层边界问题。

## Decision
refactor_existing

## Reason
`commonForbiddenImportAllowlist` 里当前有 10 条，除 `components/common/InstancePanel.vue` 外，剩余 9 条都集中在 `entities/challenge`：

- `entities/challenge/model/presentation.ts`
- `entities/challenge/ui/*`

这组文件本质上只需要：

- challenge category / difficulty / status / instance sharing 的本地表示
- 几个展示组件真正读取到的少量字段

它们不需要直接持有 API DTO contract owner。最小正确改动是：

- 在 `entities/challenge/model` 内补本地展示类型
- 让 `presentation.ts` 和各个 UI 组件切到本地 entity type
- 保持外部 feature / page 仍可用结构兼容 DTO 传入，不扩大改动面
- 删除 `commonForbiddenImportAllowlist` 中 `entities/challenge` 对应 9 条例外

## Files to modify
- .harness/reuse-decisions/challenge-entity-contract-boundary-normalization.md
- docs/plan/impl-plan/2026-05-29-challenge-entity-contract-boundary-normalization-plan.md
- docs/reviews/frontend/2026-05-29-challenge-entity-contract-boundary-normalization-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/entities/challenge/model/*
- code/frontend/src/entities/challenge/ui/*
- code/frontend/src/__tests__/architectureAllowlist.ts

## After implementation
- `entities/challenge` 不再直接依赖 `@/api/contracts`。
- `commonForbiddenImportAllowlist` 将从 10 条下降到仅剩 `components/common/InstancePanel.vue` 这 1 条。
- 外部 feature / page 继续通过结构兼容对象消费 challenge entity UI，不新增 bridge。
